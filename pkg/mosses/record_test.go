package mosses_test

import (
	"errors"
	"fmt"
	"io"
	"runtime"
	"testing"
	"time"

	"github.com/brickingsoft/brick/pkg/mosses"
)

var (
	records = []*mosses.Record{
		{
			Level:   mosses.DebugLevel,
			Time:    time.Now(),
			Message: "debug...",
			PC:      0,
			Group: mosses.Group{
				Name:   "",
				Attrs:  []mosses.Attribute{{Key: "s", Value: "sss"}, {Key: "i", Value: 1}},
				Parent: nil,
			},
		},
		{
			Level:   mosses.InfoLevel,
			Time:    time.Now(),
			Message: "info...",
			PC:      0,
			Group: mosses.Group{
				Name:   "info",
				Attrs:  []mosses.Attribute{{Key: "s", Value: "sss"}, {Key: "i", Value: 1}},
				Parent: nil,
			},
		},
		{
			Level:   mosses.WarnLevel,
			Time:    time.Now(),
			Message: "warning...",
			PC:      0,
			Group: mosses.Group{
				Name:  "warning",
				Attrs: []mosses.Attribute{{Key: "s", Value: "sss"}, {Key: "i", Value: 1}},
				Parent: &mosses.Group{
					Name:   "parent",
					Attrs:  nil,
					Parent: nil,
				},
			},
		},
		{
			Level:   mosses.ErrorLevel,
			Time:    time.Now(),
			Message: "err...",
			PC:      0,
			Group: mosses.Group{
				Name:  "err",
				Attrs: []mosses.Attribute{{Key: "s", Value: "sss"}, {Key: "i", Value: 1}},
				Parent: &mosses.Group{
					Name:  "parent1",
					Attrs: nil,
					Parent: &mosses.Group{
						Name:   "parent2",
						Attrs:  []mosses.Attribute{{Key: "s", Value: "sss"}, {Key: "i", Value: 1}},
						Parent: nil,
					},
				},
			},
		},
	}
)

func TestNewTextRecordEncoder(t *testing.T) {
	encoder := mosses.NewTextRecordEncoder()

	var pcs [1]uintptr
	runtime.Callers(1, pcs[:])

	record := &mosses.Record{
		Level:   mosses.InfoLevel,
		Time:    time.Now(),
		Message: "hello...world",
		PC:      pcs[0],
		Group: mosses.Group{
			Name:  "",
			Attrs: []mosses.Attribute{{Key: "s", Value: "sss"}, {Key: "i", Value: 1}},
			Parent: &mosses.Group{
				Name:   "moss",
				Attrs:  []mosses.Attribute{{Key: "error", Value: errors.Join(io.EOF, io.ErrNoProgress)}},
				Parent: nil,
			},
		},
	}

	p := encoder.Encode(record)
	t.Log(string(p))

	t.Log(fmt.Sprintf("%+v", errors.Join(io.EOF, io.ErrNoProgress)))

}

func TestNewColorfulTextRecordEncoder(t *testing.T) {
	encoder := mosses.NewColorfulTextRecordEncoder()
	var pcs [1]uintptr
	runtime.Callers(1, pcs[:])

	record := &mosses.Record{
		Level:   mosses.InfoLevel,
		Time:    time.Now(),
		Message: "hello...world",
		PC:      pcs[0],
		Group: mosses.Group{
			Name:  "",
			Attrs: []mosses.Attribute{{Key: "s", Value: "sss"}, {Key: "i", Value: 1}},
			Parent: &mosses.Group{
				Name:   "moss",
				Attrs:  []mosses.Attribute{{Key: "error", Value: errors.Join(io.EOF, io.ErrUnexpectedEOF)}},
				Parent: nil,
			},
		},
	}

	p := encoder.Encode(record)

	fmt.Println(string(p))
}

func TestNewJsonRecordEncoder(t *testing.T) {
	encoder := mosses.NewJsonRecordEncoder()
	for _, record := range records {
		b := encoder.Encode(record)
		t.Log(string(b))
	}
}
