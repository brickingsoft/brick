package errors_test

import (
	"fmt"
	"io"
	"testing"

	"github.com/brickingsoft/brick/rpc/errors"
)

func TestIs(t *testing.T) {
	target := io.EOF

	err := errors.New("test error")
	err = errors.Join(err, target)

	t.Log(errors.Is(err, target))
	t.Log(errors.Is(err, fmt.Errorf("test %s", "error")))
	t.Log(errors.Is(err, fmt.Errorf("test %s", "error1")))

}

var (
	gErr = errors.New("global")
)

func TestError_Format(t *testing.T) {
	err := errors.New("test error")
	t.Log(fmt.Sprintf("%+v", err))
	t.Log(fmt.Sprintf("%+v", gErr))
	t.Log(fmt.Sprintf("%+v", errors.Wrap(gErr)))
}

func TestWrap(t *testing.T) {
	err := io.EOF
	t.Log(fmt.Sprintf("%+v", errors.Wrap(err)))
}

func TestJoin(t *testing.T) {
	err := errors.New("err 0")
	err = errors.Join(err, errors.New("err 1"))
	err = errors.Join(err, errors.Join(errors.New("err 2"), errors.New("err 2.1")))
	err = errors.Join(err, errors.Join(errors.New("err 3"), errors.New("err 3.1"), errors.New("err 3.2")))

	t.Log(fmt.Sprintf("%+v", err))
}

func TestJoin_Circular(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Log(r)
		}
	}()

	err := errors.New("err")
	err1 := errors.Join(err, errors.New("err 1"))

	t.Log(fmt.Sprintf("%+v", errors.Join(err, err1)))
}

func TestJoin_Nil(t *testing.T) {
	defer func() {
		r := recover()
		if r != nil {
			t.Log(r)
		}
	}()

	err := errors.New("err")
	err1 := errors.Join(nil, errors.New("err 1"))

	t.Log(fmt.Sprintf("%+v", errors.Join(err, err1)))
}
