package ian

import (
	"testing"
	"time"
)

func TestIsTimeWithinPeriod(t *testing.T) {
	r := time.Now()

	var tests = []struct {
		name string
		t    time.Time
		p    TimeRange
		want bool
	}{
		{
			"before",
			r.Add(3),
			TimeRange{r.Add(5), r.Add(10)},
			false,
		},
		{
			"after",
			r.Add(11),
			TimeRange{r.Add(0), r.Add(10)},
			false,
		},
		{
			"inside",
			r.Add(5),
			TimeRange{r.Add(0), r.Add(10)},
			true,
		},
		{
			"on start border",
			r.Add(5),
			TimeRange{r.Add(5), r.Add(10)},
			true,
		},
		{
			"on end border",
			r.Add(5),
			TimeRange{r.Add(0), r.Add(5)},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ans := IsTimeWithinPeriod(tt.t, tt.p)
			if ans != tt.want {
				t.Errorf("got %v, want %v", ans, tt.want)
			}
		})
	}
}

func TestDoPeriodsMeet(t *testing.T) {
	r := time.Now()

	var tests = []struct {
		name   string
		p1, p2 TimeRange
		want   bool
	}{
		{
			// s1111e
			//    s2222e

			"meet in middle",
			TimeRange{r.Add(0), r.Add(10)},
			TimeRange{r.Add(5), r.Add(20)},
			true,
		},
		{
			// s1111e
			//       s2222e

			"barely touch",
			TimeRange{r.Add(0), r.Add(1)},
			TimeRange{r.Add(2), r.Add(3)},
			false,
		},
		{
			// s1111e
			//      s2222e

			"edge 1",
			TimeRange{r.Add(0), r.Add(1)},
			TimeRange{r.Add(1), r.Add(2)},
			true,
		},
		{
			//      s1111e
			// s2222e

			"edge 2",
			TimeRange{r.Add(1), r.Add(2)},
			TimeRange{r.Add(0), r.Add(1)},
			true,
		},
		{
			//        s1111e
			// s2222e

			"distant",
			TimeRange{r.Add(2), r.Add(3)},
			TimeRange{r.Add(0), r.Add(1)},
			false,
		},
		{
			//    s11111111111111111111111111111111111111111e
			// s2222e

			"big and small touch",
			TimeRange{r.Add(3), r.Add(500)},
			TimeRange{r.Add(0), r.Add(5)},
			true,
		},
		{
			// s1111e
			// s2222e

			"parallel",
			TimeRange{r.Add(0), r.Add(1)},
			TimeRange{r.Add(0), r.Add(1)},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ans := DoPeriodsMeet(tt.p1, tt.p2)
			if ans != tt.want {
				t.Errorf("got %v, want %v", ans, tt.want)
			}
		})
	}
}
