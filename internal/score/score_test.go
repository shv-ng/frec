package score

import (
	"testing"
	"time"
)

func Test_calculateScore(t *testing.T) {
	now := time.Now().Unix()

	daysAgo := func(d int) int64 {
		return now - int64(d*24*60*60)
	}

	tests := []struct {
		name      string
		firstSeen int64
		visits    []int64
		count     int64
		starred   bool
		want      int64
	}{
		{
			name:      "Basic recent user",
			firstSeen: daysAgo(2),
			visits:    []int64{daysAgo(1)},
			count:     10,
			starred:   false,
			want:      500,
		},
		{
			name:      "Starred multiplier (2x)",
			firstSeen: daysAgo(2),
			visits:    []int64{daysAgo(1)},
			count:     10,
			starred:   true,
			want:      1000,
		},
		{
			name:      "Averaging different weights",
			firstSeen: daysAgo(10),
			visits: []int64{
				daysAgo(2),
				daysAgo(40),
			},
			count:   20,
			starred: false,
			want:    130,
		},
		{
			name:      "Long term user, low activity",
			firstSeen: daysAgo(200),
			visits:    []int64{daysAgo(100)},
			count:     5,
			starred:   false,
			want:      0,
		},
		{
			name:      "No visits uses default avg 1.0",
			firstSeen: daysAgo(5),
			visits:    []int64{},
			count:     1,
			starred:   false,
			want:      20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CalculateScore(tt.firstSeen, tt.visits, tt.count, tt.starred)
			if got != tt.want {
				t.Errorf("%s: got %v, want %v", tt.name, got, tt.want)
			}
		})
	}
}

func Test_getWeight(t *testing.T) {
	tests := []struct {
		days int64
		want float64
	}{
		{0, 1.0},
		{4, 1.0},
		{10, 0.7},
		{30, 0.5},
		{60, 0.3},
		{200, 0.1},
	}

	for _, tt := range tests {
		if got := getWeight(tt.days); got != tt.want {
			t.Errorf("getWeight(%d) = %v, want %v", tt.days, got, tt.want)
		}
	}
}
