package mosses

import (
	"bytes"
	"encoding/json"
	"fmt"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fatih/color"
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

type Source struct {
	Function string
	File     string
	Line     int
}

func (src *Source) RelativePath() string {
	i := strings.LastIndex(src.Function, "/")
	if i == -1 {
		if mod := modulePath(); mod != "" {
			i = strings.LastIndex(src.File, mod)
			if i == -1 {
				return src.File
			}
			return src.File[i:]
		}
		return src.File
	}
	pkg := src.Function[:i]
	i = strings.LastIndex(src.File, pkg)
	if i == -1 {
		if mod := modulePath(); mod != "" {
			i = strings.LastIndex(src.File, mod)
			if i == -1 {
				return src.File
			}
			return src.File[i:]
		}
		return src.File
	}
	return src.File[i:]
}

func (src *Source) relativePathByFunction() string {
	i := strings.LastIndex(src.Function, "/")
	if i == -1 {
		return ""
	}
	pkg := src.Function[:i]
	i = strings.LastIndex(src.File, pkg)
	if i == -1 {
		return ""
	}
	return src.File[i:]
}

type Record struct {
	Level   Level
	Time    time.Time
	Message string
	PC      uintptr
	Group   Group
}

func (record Record) Source() Source {
	fs := runtime.CallersFrames([]uintptr{record.PC})
	f, _ := fs.Next()
	return Source{
		Function: f.Function,
		File:     f.File,
		Line:     f.Line,
	}
}

type RecordEncoder interface {
	Encode(r *Record) []byte
}

func NewTextRecordEncoder() RecordEncoder {
	return &TextRecordEncoder{
		buffers: bufferPool{},
	}
}

type TextRecordEncoder struct {
	buffers bufferPool
}

// Encode
//
// [LEVEL] [TIME] [FILE:LINE] MESSAGE [GROUP: KEY=VAL...] [GROUP: KEY=VAL...]
func (encoder *TextRecordEncoder) Encode(r *Record) []byte {
	buf := encoder.buffers.acquire()

	buf.WriteByte('[')
	buf.WriteString(fmt.Sprintf("%5s", r.Level.String()))
	buf.WriteByte(']')
	buf.WriteByte(' ')

	buf.WriteByte('[')
	buf.WriteString(r.Time.Format(time.StampMilli))
	buf.WriteByte(']')
	buf.WriteByte(' ')

	if r.PC != 0 {
		const (
			fileLineCap = 30
		)
		src := r.Source()
		file := src.RelativePath()
		fileLen := len(file)
		if fileLen < fileLineCap {
			file = fmt.Sprintf("%*s", fileLineCap, file)
		} else {
			file = file[fileLen-fileLineCap:]
			if lastDirIdx := strings.LastIndexByte(file, '/'); lastDirIdx != -1 {
				if lastDirIdx = strings.LastIndexByte(file[:lastDirIdx-1], '/'); lastDirIdx != -1 {
					prefix := strings.Repeat("*", lastDirIdx)
					file = prefix + file[lastDirIdx:]
				}
			}
		}
		buf.WriteByte('[')
		buf.WriteString(file)
		buf.WriteByte(':')
		buf.WriteString(strconv.Itoa(src.Line))
		buf.WriteByte(']')
		buf.WriteByte(' ')
	}

	buf.WriteString(r.Message)

	group := &r.Group
	for group != nil {
		if group.Name != "" {
			buf.WriteByte(' ')
			buf.WriteByte('[')
			buf.WriteString(group.Name)
		}

		if len(group.Attrs) > 0 {
			if group.Name == "" {
				buf.WriteByte(' ')
				buf.WriteByte('[')
			} else {
				buf.WriteByte(':')
				buf.WriteByte(' ')
			}
			for i, attr := range group.Attrs {
				if i > 0 {
					buf.WriteByte(',')
					buf.WriteByte(' ')
				}
				buf.WriteString(attr.Key)
				buf.WriteByte('=')
				buf.WriteString(fmt.Sprintf("%v", attr.Value))
			}
		}

		if group.Name != "" || len(group.Attrs) > 0 {
			buf.WriteByte(']')
		}

		group = group.Parent
	}

	b := buf.Bytes()
	v := make([]byte, len(b))
	copy(v, b)
	encoder.buffers.release(buf)
	return v
}

func NewColorfulTextRecordEncoder() RecordEncoder {
	return &ColorfulTextRecordEncoder{
		colorful: Colorful{
			paint: newColorfulPaint(),
		},
		buffers: bufferPool{},
	}
}

type Colorful struct {
	paint *ColorfulPaint
}

func newColorfulPaint() *ColorfulPaint {
	cp := &ColorfulPaint{
		symbol:       color.New(color.Reset, color.FgWhite, color.Bold),
		debugLevel:   color.New(color.Reset, color.FgHiCyan, color.Bold),
		infoLevel:    color.New(color.Reset, color.FgHiGreen, color.Bold),
		warnLevel:    color.New(color.Reset, color.FgHiYellow, color.Bold),
		errorLevel:   color.New(color.Reset, color.FgHiRed, color.Bold),
		time:         color.New(color.Reset, color.FgHiBlack, color.Bold),
		function:     color.New(color.Reset, color.FgHiBlack, color.Bold),
		fileLine:     color.New(color.Reset, color.FgBlue, color.Underline),
		group:        color.New(color.Reset, color.FgWhite, color.Bold),
		attrKey:      color.New(color.Reset, color.FgWhite, color.Bold),
		attrVal:      color.New(color.Reset, color.FgHiBlack),
		message:      color.New(color.Reset, color.FgWhite),
		errorBorder:  color.New(color.Reset, color.BgHiRed, color.Bold),
		errorContent: color.New(color.Reset, color.FgWhite),
	}
	cp.symbol.EnableColor()
	cp.debugLevel.EnableColor()
	cp.infoLevel.EnableColor()
	cp.warnLevel.EnableColor()
	cp.errorLevel.EnableColor()
	cp.time.EnableColor()
	cp.function.EnableColor()
	cp.fileLine.EnableColor()
	cp.group.EnableColor()
	cp.attrKey.EnableColor()
	cp.attrVal.EnableColor()
	cp.message.EnableColor()
	cp.errorBorder.EnableColor()
	cp.errorContent.EnableColor()
	return cp
}

type ColorfulPaint struct {
	symbol       *color.Color
	debugLevel   *color.Color
	infoLevel    *color.Color
	warnLevel    *color.Color
	errorLevel   *color.Color
	time         *color.Color
	function     *color.Color
	fileLine     *color.Color
	group        *color.Color
	attrKey      *color.Color
	attrVal      *color.Color
	message      *color.Color
	errorBorder  *color.Color
	errorContent *color.Color
}

type ColorfulTextRecordEncoder struct {
	colorful Colorful
	buffers  bufferPool
}

// Encode
//
// [LEVEL] [TIME] [FUNC]
// [FILE:LINE]
// [-    ] key=value ...  // group 取出 最长的 ，然后 padding
// [GROUP] key=value ...
// MESSAGE
// >>> ERROR
// {ERR}
// <<<
func (encoder *ColorfulTextRecordEncoder) Encode(r *Record) []byte {
	buf := encoder.buffers.acquire()

	_, _ = encoder.colorful.paint.symbol.Fprint(buf, "[")
	switch r.Level {
	case DebugLevel:
		_, _ = encoder.colorful.paint.debugLevel.Fprintf(buf, "%5s", r.Level.String())
		break
	case InfoLevel:
		_, _ = encoder.colorful.paint.infoLevel.Fprintf(buf, "%5s", r.Level.String())
		break
	case WarnLevel:
		_, _ = encoder.colorful.paint.warnLevel.Fprintf(buf, "%5s", r.Level.String())
		break
	case ErrorLevel:
		_, _ = encoder.colorful.paint.errorLevel.Fprintf(buf, "%5s", r.Level.String())
		break
	default:
		_, _ = encoder.colorful.paint.errorLevel.Fprintf(buf, "%5s", r.Level.String())
		break
	}
	_, _ = encoder.colorful.paint.symbol.Fprint(buf, "]")

	buf.WriteByte(' ')
	_, _ = encoder.colorful.paint.symbol.Fprint(buf, "[")
	const (
		format = "2006/01/02 15:04:05.000 MST"
	)
	_, _ = encoder.colorful.paint.time.Fprint(buf, r.Time.Format(format))
	_, _ = encoder.colorful.paint.symbol.Fprint(buf, "]")

	if r.PC != 0 {
		src := r.Source()
		buf.WriteByte(' ')
		_, _ = encoder.colorful.paint.symbol.Fprint(buf, "[")
		_, _ = encoder.colorful.paint.function.Fprint(buf, src.Function)
		_, _ = encoder.colorful.paint.symbol.Fprint(buf, "]")
		buf.WriteByte('\n')
		file := src.RelativePath()
		_, _ = encoder.colorful.paint.symbol.Fprint(buf, "[ ")
		_, _ = encoder.colorful.paint.fileLine.Fprintf(buf, "%s:%d", file, src.Line)
		_, _ = encoder.colorful.paint.symbol.Fprint(buf, " ]")
	}
	buf.WriteByte('\n')

	var err any
	var groups []*Group
	groupNameMaxLen := 0
	group := &r.Group
	for group != nil {
		if group.Name != "" || len(group.Attrs) > 0 {
			if nameLen := len(group.Name); nameLen > groupNameMaxLen {
				groupNameMaxLen = nameLen
			}
			groups = append(groups, group)
		}
		group = group.Parent
	}
	for i := len(groups); i > 0; i-- {
		group = groups[i-1]
		name := group.Name
		if name == "" {
			name = "-"
		}
		_, _ = encoder.colorful.paint.symbol.Fprint(buf, "[ ")
		_, _ = encoder.colorful.paint.group.Fprintf(buf, "%*s", groupNameMaxLen, name)
		_, _ = encoder.colorful.paint.symbol.Fprint(buf, " ]")
		for _, attr := range group.Attrs {
			if attr.Key == errAttrKey {
				err = attr.Value
				continue
			}
			_, _ = encoder.colorful.paint.attrKey.Fprintf(buf, " %s", attr.Key)
			_, _ = encoder.colorful.paint.symbol.Fprint(buf, "=")
			_, _ = encoder.colorful.paint.attrVal.Fprintf(buf, "%v", attr.Value)
		}
		buf.WriteByte('\n')
	}

	_, _ = encoder.colorful.paint.message.Fprint(buf, r.Message)

	if err != nil {
		const (
			errBorderBeg = ">>> ERROR"
			errBorderEnd = "<<<"
		)
		buf.WriteByte('\n')
		_, _ = encoder.colorful.paint.errorBorder.Fprint(buf, errBorderBeg)
		buf.WriteByte('\n')
		content := fmt.Sprintf("%+v", err)
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			_, _ = encoder.colorful.paint.errorContent.Fprintln(buf, line)
		}
		_, _ = encoder.colorful.paint.errorBorder.Fprint(buf, errBorderEnd)
	}

	b := buf.Bytes()
	v := make([]byte, len(b))
	copy(v, b)
	encoder.buffers.release(buf)
	return v
}

func NewJsonRecordEncoder() RecordEncoder {
	return &JsonRecordEncoder{
		buffers: bufferPool{},
	}
}

type JsonRecordEncoder struct {
	buffers bufferPool
}

func (encoder *JsonRecordEncoder) Encode(r *Record) []byte {
	buf := encoder.buffers.acquire()
	buf.WriteByte('{')
	// level
	buf.WriteByte('"')
	buf.WriteString("level")
	buf.WriteByte('"')
	buf.WriteByte(':')
	buf.WriteByte(' ')
	buf.WriteByte('"')
	buf.WriteString(r.Level.String())
	buf.WriteByte('"')
	buf.WriteByte(',')
	buf.WriteByte(' ')
	// time
	buf.WriteByte('"')
	buf.WriteString("time")
	buf.WriteByte('"')
	buf.WriteByte(' ')
	buf.WriteByte(':')
	buf.WriteByte('"')
	buf.WriteString(r.Time.Format(time.RFC3339Nano))
	buf.WriteByte('"')
	buf.WriteByte(',')
	buf.WriteByte(' ')
	// message
	buf.WriteByte('"')
	buf.WriteString("message")
	buf.WriteByte('"')
	buf.WriteByte(':')
	buf.WriteByte(' ')
	message, _ := json.Marshal(r.Message)
	buf.Write(message)
	// source
	if r.PC != 0 {
		src := r.Source()
		buf.WriteByte(',')
		buf.WriteByte(' ')
		buf.WriteByte('"')
		buf.WriteString("source")
		buf.WriteByte('"')
		buf.WriteByte(':')
		buf.WriteByte(' ')
		buf.WriteByte('{')
		// fn
		buf.WriteByte('"')
		buf.WriteString("function")
		buf.WriteByte('"')
		buf.WriteByte(':')
		buf.WriteByte(' ')
		buf.WriteByte('"')
		buf.WriteString(src.Function)
		buf.WriteByte('"')
		buf.WriteByte(',')
		buf.WriteByte(' ')
		// file
		buf.WriteByte('"')
		buf.WriteString("file")
		buf.WriteByte('"')
		buf.WriteByte(':')
		buf.WriteByte(' ')
		buf.WriteByte('"')
		file := src.RelativePath()
		buf.WriteString(file)
		buf.WriteByte('"')
		buf.WriteByte(',')
		buf.WriteByte(' ')
		// line
		buf.WriteByte('"')
		buf.WriteString("line")
		buf.WriteByte('"')
		buf.WriteByte(':')
		buf.WriteByte(' ')
		buf.WriteString(strconv.Itoa(src.Line))
		buf.WriteByte('}')
	}
	// group
	if r.Group.Name != "" || len(r.Group.Attrs) > 0 {
		// name
		if r.Group.Name != "" {
			buf.WriteByte(',')
			buf.WriteByte(' ')
			buf.WriteByte('"')
			buf.WriteString("group")
			buf.WriteByte('"')
			buf.WriteByte(':')
			buf.WriteByte(' ')
			buf.WriteByte('"')
			buf.WriteString(r.Group.Name)
			buf.WriteByte('"')
		}
		// attr
		if len(r.Group.Attrs) > 0 {
			buf.WriteByte(',')
			buf.WriteByte(' ')
			writeAttrsToJsonBuffer(r.Group.Attrs, buf)
		}
		// parent
		if r.Group.Parent != nil {
			buf.WriteByte(',')
			buf.WriteByte(' ')
			writeParentGroupToJsonBuffer(r.Group.Parent, buf)
		}
	}
	buf.WriteByte('}')
	b := buf.Bytes()
	v := make([]byte, len(b))
	copy(v, b)
	encoder.buffers.release(buf)
	return v
}

func writeParentGroupToJsonBuffer(group *Group, buf *bytes.Buffer) {
	buf.WriteByte('"')
	buf.WriteString("attached")
	buf.WriteByte('"')
	buf.WriteByte(':')
	buf.WriteByte(' ')
	buf.WriteByte('{')
	// name
	if group.Name != "" {
		buf.WriteByte('"')
		buf.WriteString("group")
		buf.WriteByte('"')
		buf.WriteByte(':')
		buf.WriteByte(' ')
		buf.WriteByte('"')
		buf.WriteString(group.Name)
		buf.WriteByte('"')
	}
	// attr
	if len(group.Attrs) > 0 {
		if group.Name != "" {
			buf.WriteByte(',')
			buf.WriteByte(' ')
		}
		writeAttrsToJsonBuffer(group.Attrs, buf)
	}
	// parent
	if group.Parent != nil {
		if group.Parent.Name != "" || len(group.Parent.Attrs) > 0 {
			buf.WriteByte(',')
			buf.WriteByte(' ')
		}
		writeParentGroupToJsonBuffer(group.Parent, buf)
	}
	buf.WriteByte('}')
}

func writeAttrsToJsonBuffer(attrs []Attribute, buf *bytes.Buffer) {
	if len(attrs) == 0 {
		return
	}
	buf.WriteByte('"')
	buf.WriteString("attrs")
	buf.WriteByte('"')
	buf.WriteByte(':')
	buf.WriteByte(' ')
	buf.WriteByte('{')
	for i, attr := range attrs {
		if i > 0 {
			buf.WriteByte(',')
			buf.WriteByte(' ')
		}
		buf.WriteByte('"')
		buf.WriteString(attr.Key)
		buf.WriteByte('"')
		buf.WriteByte(':')
		buf.WriteByte(' ')
		b, err := json.Marshal(attr.Value)
		if err != nil {
			buf.WriteByte('"')
			buf.WriteString(fmt.Sprintf("!#FAILED(%v)", err))
			buf.WriteByte('"')
			continue
		}
		buf.Write(b)
	}
	buf.WriteByte('}')
}

type bufferPool struct {
	pool sync.Pool
}

func (encoder *bufferPool) acquire() *bytes.Buffer {
	v := encoder.pool.Get()
	if v == nil {
		return bytes.NewBuffer(nil)
	}
	return v.(*bytes.Buffer)
}

func (encoder *bufferPool) release(b *bytes.Buffer) {
	b.Reset()
	encoder.pool.Put(b)
}
