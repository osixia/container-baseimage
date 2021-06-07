package helpers

import (
	"fmt"
	"strconv"
)

type NilBool struct {
	Value *bool
}

func (b *NilBool) Set(s string) error {

	if s == "" {
		s = "true"
	}

	v, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	b.Value = &v

	return nil
}

func (b *NilBool) String() string {
	if b.Value == nil {
		return ""
	}
	return fmt.Sprintf("%t", *b.Value)
}

func (b *NilBool) Type() string {
	return "bool"
}
