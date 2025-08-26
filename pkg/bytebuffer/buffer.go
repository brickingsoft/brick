package bytebuffer

import (
	"errors"
	"io"
)

type Buffer struct {
	r int64
	b []byte
}

func (buf *Buffer) Len() int {
	if len(buf.b) == 0 {
		return 0
	}
	return len(buf.b[buf.r:])
}

func (buf *Buffer) Peek() []byte {
	return buf.b[buf.r:]
}

func (buf *Buffer) Read(p []byte) (n int, err error) {
	if len(p) == 0 {
		return
	}
	if buf.Len() == 0 {
		err = io.EOF
		return
	}
	n = copy(p, buf.b[buf.r:])
	buf.r += int64(n)
	return
}

func (buf *Buffer) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(buf.b[buf.r:])
	buf.r += int64(n)
	return int64(n), err
}

func (buf *Buffer) Write(p []byte) (int, error) {
	buf.b = append(buf.b, p...)
	return len(p), nil
}

func (buf *Buffer) WriteByte(c byte) error {
	buf.b = append(buf.b, c)
	return nil
}

func (buf *Buffer) WriteString(s string) (int, error) {
	buf.b = append(buf.b, s...)
	return len(s), nil
}

func (buf *Buffer) Set(p []byte) {
	buf.b = append(buf.b[:0], p...)
	buf.r = 0
}

func (buf *Buffer) SetString(s string) {
	buf.b = append(buf.b[:0], s...)
	buf.r = 0
}

func (buf *Buffer) ReadFrom(r io.Reader) (int64, error) {
	p := buf.b
	nStart := int64(len(p))
	nMax := int64(cap(p))
	n := nStart
	if nMax == 0 {
		nMax = 64
		p = make([]byte, nMax)
	} else {
		p = p[:nMax]
	}
	for {
		if n == nMax {
			nMax *= 2
			bNew := make([]byte, nMax)
			copy(bNew, p)
			p = bNew
		}
		nn, err := r.Read(p[n:])
		n += int64(nn)
		if err != nil {
			buf.b = p[:n]
			n -= nStart
			if err == io.EOF {
				return n, nil
			}
			return n, err
		}
	}
}

func (buf *Buffer) String() string {
	return string(buf.b[buf.r:])
}

func (buf *Buffer) Reset() {
	buf.r = 0
	buf.b = buf.b[:0]
}

func (buf *Buffer) Seek(offset int64, whence int) (n int64, err error) {
	switch whence {
	case io.SeekCurrent:
		n = buf.r + offset
		if n < 0 || int64(len(buf.b)) < n {
			n = 0
			err = errors.New("failed to seek cause offset out of range")
			return
		}
		buf.r = n
		return
	case io.SeekStart:
		if offset < 0 || int64(len(buf.b)) < offset {
			err = errors.New("failed to seek cause offset out of range")
			return
		}
		buf.r = offset
		n = offset
		return
	case io.SeekEnd:
		n = int64(len(buf.b)) + offset
		if n < 0 || int64(len(buf.b)) < n {
			n = 0
			err = errors.New("failed to seek cause offset out of range")
			return
		}
		buf.r = n
		return
	default:
		err = errors.New("failed to seek cause invalid whence")
		return
	}
}

func (buf *Buffer) Clone() []byte {
	bLen := buf.Len()
	if bLen == 0 {
		return nil
	}
	dst := make([]byte, bLen)
	copy(dst, buf.b)
	return dst
}
