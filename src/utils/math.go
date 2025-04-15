package utils

import "math"

func Average(xs []float64) float64 {
	total := 0.0
	for _, v := range xs {
		total += v
	}
	return total / float64(len(xs))
}

func StandardDeviation(xs []float64) float64 {
	if len(xs) == 0 {
		return 0.0
	}

	mean := Average(xs)
	var varianceSum float64

	for _, v := range xs {
		varianceSum += math.Pow(v-mean, 2)
	}

	variance := varianceSum / float64(len(xs))
	return math.Sqrt(variance)
}
