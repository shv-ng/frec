package score

import (
	"time"
)

type WeightRange struct {
	MaxDays int64
	Weight  float64
}

var scoreWeights = []WeightRange{
	{MaxDays: 4, Weight: 1.0},
	{MaxDays: 14, Weight: 0.7},
	{MaxDays: 31, Weight: 0.5},
	{MaxDays: 90, Weight: 0.3},
	{MaxDays: 99999, Weight: 0.1},
}

func CalculateScore(firstSeen int64, visits []int64, count int64, starred bool) int64 {
	days := daysSince(firstSeen)

	var sum float64
	for _, v := range visits {
		d := daysSince(v)
		sum += getWeight(d)
	}

	avg := 1.0
	if len(visits) > 0 {
		avg = sum / float64(len(visits))
	}

	score := (float64(count) * avg * 100) / float64(days)
	if starred {
		score *= 1.6 // magic number
	}

	return int64(score)
}

func getWeight(d int64) float64 {
	for _, w := range scoreWeights {
		if d <= w.MaxDays {
			return w.Weight
		}
	}
	return 0
}

func daysSince(ts int64) int64 {
	return max(1, int64(time.Since(time.Unix(ts, 0)).Hours()/24))
}
