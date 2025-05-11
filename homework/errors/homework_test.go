package main

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// go test -v homework_test.go

type MultiError struct {
	errors []error
}

func (e *MultiError) Error() string {
	if e == nil || len(e.errors) == 0 {
		return ""
	}

	header := "errors"
	if len(e.errors) == 1 {
		header = "error"
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%d %s occured:\n", len(e.errors), header))

	for _, err := range e.errors {
		if err != nil {
			sb.WriteString("\t* " + err.Error())
		}
	}

	sb.WriteString("\n")

	return sb.String()
}

func Append(err error, errs ...error) *MultiError {
	me := &MultiError{}
	if err != nil {
		var existing *MultiError
		if errors.As(err, &existing) {
			me = existing
		} else {
			me = &MultiError{errors: []error{err}}
		}
	}

	for _, err := range errs {
		if err == nil {
			continue
		}

		var multiError *MultiError

		if errors.As(err, &multiError) {
			me.errors = append(me.errors, multiError.errors...)
		} else {
			me.errors = append(me.errors, err)
		}
	}

	if len(me.errors) == 0 {
		return nil
	}

	return me
}

func TestMultiError(t *testing.T) {
	var err error
	err = Append(err, errors.New("error 1"))
	err = Append(err, errors.New("error 2"))

	expectedMessage := "2 errors occured:\n\t* error 1\t* error 2\n"
	assert.EqualError(t, err, expectedMessage)
}
