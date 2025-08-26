package errors

import (
	"errors"
)

func Wrap(target error, attrs ...Attribute) error {
	if target == nil {
		return nil
	}
	err := wrap(target)
	if len(attrs) > 0 {
		err.Attrs = append(err.Attrs, attrs...)
	}
	err.sourcing(3, true)
	return err
}

func wrap(target error) *Error {
	if target == nil {
		return nil
	}
	var err *Error
	// check *Error
	if ok := errors.As(target, &err); ok {
		return err
	}
	// check errors.Join()
	if unwrapped, ok := target.(interface{ Unwrap() []error }); ok {
		errs := unwrapped.Unwrap()
		errsLen := len(errs)
		if errsLen == 0 {
			return nil
		}
		for i := 0; i < errsLen; i++ {
			err.Wrapped = append(err.Wrapped, wrap(errs[i]))
		}
		return err
	}
	// wrap
	err = &Error{
		Message: target.Error(),
	}
	// check unwrapped
	if unwrapped := errors.Unwrap(target); unwrapped != nil {
		err.Wrapped = append(err.Wrapped, wrap(unwrapped))
	}
	return err
}
