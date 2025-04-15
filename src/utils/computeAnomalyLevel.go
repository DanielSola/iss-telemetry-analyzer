package utils

import "math"

type AnomalyLevel string

const (
	NO_ANOMALY AnomalyLevel = "NO_ANOMALY"
	MEDIUM     AnomalyLevel = "MEDIUM"
	ANOMALY    AnomalyLevel = "ANOMALY"
)

func (a AnomalyLevel) String() string {
	switch a {
	case NO_ANOMALY:
		return "NO_ANOMALY"
	case MEDIUM:
		return "MEDIUM"
	case ANOMALY:
		return "ANOMALY"
	default:
		return "Unknown"
	}
}

func ComputeAnomalyLevel(anomalyScore, stdScore, avgScore float64) AnomalyLevel {

	scoreDeviation := math.Abs(anomalyScore - avgScore)

	if scoreDeviation <= 2*stdScore {
		return NO_ANOMALY
	}

	if scoreDeviation > 2*stdScore && scoreDeviation <= 3*stdScore {
		return MEDIUM
	}

	return ANOMALY
}
