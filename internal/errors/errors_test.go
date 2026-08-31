package errors

import "testing"

func TestBadRequestError(t *testing.T) {
	e := &BadRequestError{Message: "invalid input"}
	want := "bad request: invalid input"
	if got := e.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestHTTPError(t *testing.T) {
	e := &HTTPError{StatusCode: 404, Body: "not found"}
	want := "HTTP 404: not found"
	if got := e.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestInvalidOptionsError(t *testing.T) {
	e := &InvalidOptionsError{Field: "symbol", Msg: "is required"}
	want := "invalid option symbol: is required"
	if got := e.Error(); got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestFailedValidationError(t *testing.T) {
	e := &FailedValidationError{Errors: []string{"field A failed", "field B failed"}}
	got := e.Error()
	if got == "" {
		t.Error("Error() returned empty string")
	}
	for _, msg := range []string{"field A failed", "field B failed"} {
		if !contains(got, msg) {
			t.Errorf("Error() = %q, want to contain %q", got, msg)
		}
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && searchString(s, sub)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
