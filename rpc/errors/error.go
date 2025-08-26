package errors

import (
	"errors"
	"fmt"
	"io"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"unsafe"

	"github.com/brickingsoft/brick/pkg/bytebuffer"
)

var (
	_sourcing int64 = 1
)

func DisableSourcing() {
	atomic.StoreInt64(&_sourcing, 0)
}

func EnableSourcing() {
	atomic.StoreInt64(&_sourcing, 1)
}

func SourcingEnabled() bool {
	return atomic.LoadInt64(&_sourcing) == 1
}

type Source struct {
	Function string `json:"function"`
	File     string `json:"file"`
	Line     int    `json:"line"`
}

type Attribute struct {
	Key   string
	Value string
}

func Attr(key string, value string) Attribute {
	return Attribute{
		Key:   key,
		Value: value,
	}
}

type Error struct {
	Message string      `json:"message"`
	Source  Source      `json:"source"`
	Attrs   []Attribute `json:"attrs"`
	Wrapped []*Error    `json:"wrapped"`
}

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Unwrap() []error {
	if len(e.Wrapped) == 0 {
		return nil
	}
	errs := make([]error, len(e.Wrapped))
	for i, err := range e.Wrapped {
		errs[i] = err
	}
	return errs
}

func (e *Error) Is(target error) bool {
	if e == nil || target == nil {
		return false
	}
	return e.Message == target.Error()
}

const (
	unknown = "???"
)

var (
	_mod     = ""
	_modOnce = sync.Once{}
)

func modulePath() string {
	_modOnce.Do(func() {
		if info, ok := debug.ReadBuildInfo(); ok {
			_mod = info.Main.Path
		}
	})
	return _mod
}

func (e *Error) sourcing(skip int, retry bool) {
	if SourcingEnabled() && (e.Source.Line == 0 || retry || e.Source.Function == unknown) {
		var pcs [1]uintptr
		runtime.Callers(skip, pcs[:])
		fs := runtime.CallersFrames(pcs[:])
		f, _ := fs.Next()
		e.Source.Function = f.Function
		e.Source.File = f.File
		e.Source.Line = f.Line

		if e.Source.File == "" {
			e.Source.File = unknown
		} else {
			if e.Source.Function == "" {
				if mod := modulePath(); mod != "" {
					if i := strings.LastIndex(e.Source.File, mod); i > -1 {
						e.Source.File = e.Source.File[i:]
					}
				}
			} else {
				if i := strings.LastIndex(e.Source.Function, "/"); i > -1 {
					pkg := e.Source.Function[:i]
					if i = strings.LastIndex(e.Source.File, pkg); i > -1 {
						e.Source.File = e.Source.File[i:]
					} else {
						if mod := modulePath(); mod != "" {
							if i := strings.LastIndex(e.Source.File, mod); i > -1 {
								e.Source.File = e.Source.File[i:]
							}
						}
					}
				} else {
					if mod := modulePath(); mod != "" {
						if i := strings.LastIndex(e.Source.File, mod); i > -1 {
							e.Source.File = e.Source.File[i:]
						}
					}
				}
			}
		}

		if e.Source.Function == "" {
			e.Source.Function = unknown
		}

	}
	return
}

func (e *Error) WriteTo(w io.Writer) (n int64, err error) {
	s := e.Error()
	b := unsafe.Slice(unsafe.StringData(s), len(s))
	nn, wErr := w.Write(b)
	if wErr != nil {
		err = wErr
		return
	}
	n = int64(nn)
	return
}

func (e *Error) Format(state fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case state.Flag('+'):
			buf := bytebuffer.Acquire()
			format(buf, e, nil)
			b := buf.Peek()
			bytebuffer.Release(buf)
			_, _ = state.Write(b)
			break
		default:
			_, _ = e.WriteTo(state)
			break
		}
	case 's':
		_, _ = e.WriteTo(state)
		break
	default:
		_, _ = e.WriteTo(state)
		break
	}
}

const (
	errorKey    = "[ error    ]"
	positionKey = "[ position ]"
	messageKey  = "[ message  ]"
	wrappedKey  = "|---"
	spaceKey    = "     "
)

func format(buf *bytebuffer.Buffer, err *Error, layers []int) {
	/* fmt
		[ error    ] [key=value, ...]
		[ position ] [ fn ] [ file:line ]
		[ message  ] ....
		|--- [ 1 ]
	         [ error ]
	         |--- [ 1-1 ]
	              [ error ] [key=value, ...]
	              |--- [ 1-1-1 ]
	    |--- [ 2 ]
	         |--- [ 2-1 ]
	*/
	head := strings.Repeat(spaceKey, len(layers))
	_, _ = buf.WriteString(head)
	_, _ = buf.WriteString(errorKey)

	if len(err.Attrs) > 0 {
		_ = buf.WriteByte(' ')
		_ = buf.WriteByte('[')
		_ = buf.WriteByte(' ')
		for i, attr := range err.Attrs {
			if i > 0 {
				_ = buf.WriteByte(',')
				_ = buf.WriteByte(' ')
			}
			_, _ = buf.WriteString(attr.Key)
			_ = buf.WriteByte('=')
			_, _ = buf.WriteString(attr.Value)
		}
		_ = buf.WriteByte(' ')
		_ = buf.WriteByte(']')
	}

	_ = buf.WriteByte('\n')

	if err.Source.Line > 0 {
		_, _ = buf.WriteString(head)
		_, _ = buf.WriteString(positionKey)
		_ = buf.WriteByte(' ')
		_ = buf.WriteByte('[')
		_ = buf.WriteByte(' ')
		_, _ = buf.WriteString(err.Source.Function)
		_ = buf.WriteByte(' ')
		_ = buf.WriteByte(']')
		_ = buf.WriteByte(' ')
		_ = buf.WriteByte('[')
		_ = buf.WriteByte(' ')
		_, _ = buf.WriteString(err.Source.File)
		_ = buf.WriteByte(':')
		_, _ = buf.WriteString(strconv.Itoa(err.Source.Line))
		_ = buf.WriteByte(' ')
		_ = buf.WriteByte(']')
		_ = buf.WriteByte('\n')
	}

	_, _ = buf.WriteString(head)
	_, _ = buf.WriteString(messageKey)
	_ = buf.WriteByte(' ')
	_, _ = buf.WriteString(err.Message)

	if wLen := len(err.Wrapped); wLen > 0 {

		for i, wrapped := range err.Wrapped {
			_ = buf.WriteByte('\n')
			_, _ = buf.WriteString(head)
			_, _ = buf.WriteString(wrappedKey)
			_ = buf.WriteByte(' ')
			_ = buf.WriteByte('[')
			_ = buf.WriteByte(' ')
			for _, layer := range layers {
				_, _ = buf.WriteString(strconv.Itoa(layer))
				_ = buf.WriteByte('-')
			}
			_, _ = buf.WriteString(strconv.Itoa(i + 1))
			_ = buf.WriteByte(' ')
			_ = buf.WriteByte(']')
			_ = buf.WriteByte('\n')
			format(buf, wrapped, append(layers, i+1))
		}

	}

}

func New(message string, attrs ...Attribute) error {
	e := &Error{
		Message: message,
		Attrs:   attrs,
	}
	e.sourcing(3, false)
	return e
}

func Is(err error, target error) bool {
	return errors.Is(err, target)
}

func As(err error, target any) bool {
	return errors.As(err, target)
}
