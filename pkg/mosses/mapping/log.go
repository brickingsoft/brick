package mapping

import (
	"bytes"
	"context"
	"log"

	"github.com/brickingsoft/brick/pkg/mosses"
)

func SetDefaultLogger(ctx context.Context, logger mosses.Logger) {
	log.Default().SetOutput(&LoggerWriter{
		ctx:   ctx,
		moss:  logger.CallerSkipShift(3),
		proxy: log.Default(),
	})
}

func Logger(ctx context.Context, logger mosses.Logger) (v *log.Logger) {
	w := &LoggerWriter{ctx: ctx, moss: logger.CallerSkipShift(3)}
	v = log.New(w, "", 0)
	w.proxy = v
	return
}

type LoggerWriter struct {
	ctx   context.Context
	moss  mosses.Logger
	proxy *log.Logger
}

func (w *LoggerWriter) Write(p []byte) (n int, err error) {
	if len(p) == 0 {
		return
	}
	flags := w.proxy.Flags()
	// prefix
	if flags&log.Lmsgprefix == 0 {
		p = p[len(w.proxy.Prefix()):]
	}
	// date
	if flags&(log.Ldate|log.Ltime|log.Lmicroseconds) != 0 {
		if flags&log.Ldate != 0 {
			p = p[11:]
		}
		if flags&(log.Ltime|log.Lmicroseconds) != 0 {
			p = p[8:]
			if flags&log.Lmicroseconds != 0 {
				p = p[7:]
			}
			p = p[1:]
		}
	}
	// file
	if flags&(log.Lshortfile|log.Llongfile) != 0 {
		i0 := bytes.IndexByte(p, ':')
		i1 := bytes.IndexByte(p[i0+1:], ':')
		p = p[i0+i1+3:]
	}
	// message prefix
	if flags&log.Lmsgprefix != 0 {
		p = p[len(w.proxy.Prefix()):]
	}
	// line
	if p[len(p)-1] == '\n' {
		p = p[:len(p)-1]
	}
	// write
	w.moss.Debug(w.ctx, string(p))
	n = len(p)
	return
}
