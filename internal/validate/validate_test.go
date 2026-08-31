package validate

import (
	"testing"
	"time"
)

type validStruct struct {
	Period1  time.Time `validate:"required"`
	Interval string    `validate:"omitempty,oneof=1d 1wk 1mo"`
}

type invalidStruct struct {
	Period1  time.Time `validate:"required"`
	Interval string    `validate:"omitempty,oneof=1d 1wk 1mo"`
}

func TestStructValid(t *testing.T) {
	v := validStruct{
		Period1:  time.Now(),
		Interval: "1d",
	}
	if err := Struct(v); err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
}

func TestStructInvalid(t *testing.T) {
	v := invalidStruct{
		Period1:  time.Time{},
		Interval: "bad",
	}
	if err := Struct(v); err == nil {
		t.Fatal("expected error, got nil")
	}
}
