package plugin

import (
	"testing"
)

func TestFloatConversion(t *testing.T) {
	tests := []struct {
		in interface{}

		value float64
		err   error
	}{
		{1, 1.0, nil},
		{1.00000001, 1.00000001, nil},
		{"1", 1.0, nil},
		{"1.00000001", 1.00000001, nil},
		{"1487630856.389418293", 1487630856.389418, nil},
	}

	for _, test := range tests {
		got, err := i2f(test.in)
		if test.value != got || test.err != err {
			t.Errorf("Expected %f (%s) got %f (%s)", test.value, test.err, got, err)
		}
	}
}
