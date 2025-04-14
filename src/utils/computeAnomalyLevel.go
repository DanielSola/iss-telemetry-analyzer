package utils

import "math"

type AnomalyLevel string

const (
	NO_ANOMALY AnomalyLevel = "NO_ANOMALY"
	MEDIUM     AnomalyLevel = "MEDIUM"
	ANOMALY    AnomalyLevel = "ANOMALY"
)

func ComputeAnomalyLevel(anomalyScore, stdScore, avgScore float64) AnomalyLevel {

	scoreDeviation := math.Abs(anomalyScore - avgScore)

	if scoreDeviation <= stdScore {
		return NO_ANOMALY
	}

	if scoreDeviation > stdScore && scoreDeviation <= 3*stdScore {
		return MEDIUM
	}

	return ANOMALY
}
