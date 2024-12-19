package main

import "fmt"

func wrapError(wrapper string, err error) error {
	if err != nil {
		return fmt.Errorf("%s: %w", wrapper, err)
	}
	return nil
}
