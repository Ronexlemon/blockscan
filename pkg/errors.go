package pkg

import "fmt"

type ChainError struct {
	Message string
	Err     error
}

func (e *ChainError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

func (e *ChainError) Unwrap() error {
	return e.Err
}


func WrapError(msg string, err error) error {
	if err == nil {
		return nil
	}
	return &ChainError{
		Message: msg,
		Err:     err,
	}
}