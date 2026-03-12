package score

import (
	"time"
)

type weightRange struct {
	maxDays int64
	weight  float64
}

var scoreWeights = []weightRange{
	{maxDays: 4, weight: 1.0},
	{maxDays: 14, weight: 0.7},
	{maxDays: 31, weight: 0.5},
	{maxDays: 90, weight: 0.3},
	{maxDays: 99999, weight: 0.1},
}

func CalculateScore(firstSeen int64, visits []int64, count int64, starred bool) int64 {
	days := max(1, DaysSince(firstSeen))

	var sum float64
	for _, v := range visits {
		d := max(1, DaysSince(v))
		sum += getWeight(d)
	}

	avg := 1.0
	if len(visits) > 0 {
		avg = sum / float64(len(visits))
	}

	score := (float64(count) * avg * 100) / float64(days)
	if starred {
		score *= 2 // magic number
	}

	return int64(score)
}

func getWeight(d int64) float64 {
	for _, w := range scoreWeights {
		if d <= w.maxDays {
			return w.weight
		}
	}
	return 0
}

func DaysSince(ts int64) int64 {
	return int64(time.Since(time.Unix(ts, 0)).Hours() / 24)
}
