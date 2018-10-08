package cron

import (
	"testing"
	"time"
)

func TestTimeListNext(t *testing.T) {
	tests := []struct {
		startTime string
		times     []string
		expected  []string
	}{
		// Simple cases
		{
			"2018-09-01 08:01:02",
			[]string{"2018-09-01 10:01:02"},
			[]string{"2018-09-01 10:01:02"},
		},

		// sort list
		{
			"2018-09-01 08:01:02",
			[]string{"2018-09-01 10:01:02", "2018-09-02 10:01:02"},
			[]string{"2018-09-01 10:01:02", "2018-09-02 10:01:02"},
		},

		// sort list with middle start time
		{
			"2018-09-01 10:11:02",
			[]string{"2018-09-01 10:01:02", "2018-09-02 10:01:02"},
			[]string{"2018-09-02 10:01:02"},
		},

		// unsorted list
		{
			"2018-07-01 08:01:02",
			[]string{"2018-09-01 10:01:00", "2018-08-01 10:00:00", "2018-09-01 10:00:00", "2018-08-02 10:01:02"},
			[]string{"2018-08-01 10:00:00", "2018-08-02 10:01:02", "2018-09-01 10:00:00", "2018-09-01 10:01:00"},
		},

		// unsorted list with middle start time
		{
			"2018-08-03 12:00:00",
			[]string{"2018-09-01 10:01:00", "2018-08-01 10:00:00", "2018-09-01 10:00:00", "2018-08-02 10:01:02"},
			[]string{"2018-09-01 10:00:00", "2018-09-01 10:01:00"},
		},
	}

	for _, c := range tests {
		tls := At(getAtTimes(c.times))
		nextTime := getAtTime(c.startTime)
		for _, trun := range c.expected {
			actual := tls.Next(nextTime)
			expected := getAtTime(trun)
			if actual != expected {
				t.Errorf("%s, \"%s\": (expected) %v != %v (actual)",
					c.startTime, c.times, expected, actual)
			}
			nextTime = actual
		}
		if actual := tls.Next(nextTime); !actual.IsZero() {
			t.Errorf("%s, \"%s\": next time should be zero, but got %v (actual)",
				c.startTime, c.times, actual)
		}

	}
}

func getAtTime(value string) time.Time {
	if value == "" {
		panic("time string is empty")
	}

	t, err := time.Parse("2006-01-02 15:04:05", value)
	if err != nil {
		panic(err)
	}

	return t
}

func getAtTimes(values []string) []time.Time {
	tl := []time.Time{}
	for _, v := range values {
		tl = append(tl, getAtTime(v))
	}
	return tl
}
