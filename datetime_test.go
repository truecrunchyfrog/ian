package ian

import (
	"testing"
	"time"
)

func TestIsTimeWithinPeriod(t *testing.T) {
	r := time.Now()

	var tests = []struct {
		name          string
		start, t, end time.Time
		want          bool
	}{
		{
			"before",
			r.Add(5),
			r.Add(3),
			r.Add(10),
			false,
		},
		{
			"after",
			r.Add(0),
			r.Add(11),
			r.Add(10),
			false,
		},
		{
			"inside",
			r.Add(0),
			r.Add(5),
			r.Add(10),
			true,
		},
		{
			"on start border",
			r.Add(5),
			r.Add(5),
			r.Add(10),
			true,
		},
		{
			"on end border",
			r.Add(0),
			r.Add(5),
			r.Add(5),
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ans := IsTimeWithinPeriod(tt.t, tt.start, tt.end)
			if ans != tt.want {
				t.Errorf("got %v, want %v", ans, tt.want)
			}
		})
	}
}

func TestDoPeriodsMeet(t *testing.T) {
	r := time.Now()

	var tests = []struct {
		name           string
		s1, e1, s2, e2 time.Time
		want           bool
	}{
		{
			// s1111e
			//    s2222e

			"meet in middle",
			r.Add(0),
			r.Add(10),
			r.Add(5),
			r.Add(20),
			true,
		},
		{
			// s1111e
			//       s2222e

			"barely touch",
			r.Add(0),
			r.Add(1),
			r.Add(2),
			r.Add(3),
			false,
		},
		{
			// s1111e
			//      s2222e

			"edge 1",
			r.Add(0),
			r.Add(1),
			r.Add(2),
			r.Add(3),
			true,
		},
		{
			//      s1111e
			// s2222e

			"edge 2",
			r.Add(2),
			r.Add(3),
			r.Add(0),
			r.Add(1),
			true,
		},
		{
			//        s1111e
			// s2222e

			"distant",
			r.Add(2),
			r.Add(3),
			r.Add(0),
			r.Add(1),
			false,
		},
		{
			//    s11111111111111111111111111111111111111111e
			// s2222e

			"big and small touch",
			r.Add(3),
			r.Add(500),
			r.Add(0),
			r.Add(5),
			true,
		},
		{
			// s1111e
			// s2222e

			"parallel",
			r.Add(0),
			r.Add(1),
			r.Add(0),
			r.Add(1),
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ans := DoPeriodsMeet(tt.s1, tt.e1, tt.s2, tt.e2)
			if ans != tt.want {
				t.Errorf("got %v, want %v", ans, tt.want)
			}
		})
	}
}
