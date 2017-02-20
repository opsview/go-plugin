package plugin

import (
	"testing"
)

func TestExitCode(t *testing.T) {
	tests := []struct {
		in  Status
		out int
	}{
		{OK, 0},
		{WARNING, 1},
		{CRITICAL, 2},
		{UNKNOWN, 3},
	}

	for _, test := range tests {
		code := test.in.ExitCode()
		if test.out != code {
			t.Errorf("Expected %s got %s", test.out, code)
		}
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		in  Status
		out string
	}{
		{OK, "OK"},
		{WARNING, "WARNING"},
		{CRITICAL, "CRITICAL"},
		{UNKNOWN, "UNKNOWN"},
	}

	for _, test := range tests {
		text := test.in.String()
		if test.out != text {
			t.Errorf("Expected %s got %s", test.out, text)
		}
	}
}
