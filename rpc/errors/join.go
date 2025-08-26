package errors

import "fmt"

func Join(errs ...error) error {
	errsLen := len(errs)
	if errsLen == 0 {
		return nil
	}
	head := wrap(errs[0])
	if head == nil {
		head = &Error{}
		head.sourcing(3, true)
		panic(fmt.Errorf(
			"[ errors: join nil ] [ %s ] [ %s:%d ] ",
			head.Source.Function, head.Source.File, head.Source.Line,
		))
		return nil
	}
	head.sourcing(3, true)

	for i := 1; i < errsLen; i++ {
		next := wrap(errs[i])
		if next == nil {
			continue
		}
		if head == next {
			panic(fmt.Errorf(
				"[ errors: circular join ] [ %s ] [ %s:%d ] ",
				head.Source.Function, head.Source.File, head.Source.Line,
			))
			return nil
		}
		head.Wrapped = append(head.Wrapped, wrap(errs[i]))
	}

	return head
}
