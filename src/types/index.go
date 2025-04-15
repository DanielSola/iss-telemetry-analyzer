package types

type TelemetryData struct {
	Name      string `json:"name"`
	Value     string `json:"value"`
	Timestamp string `json:"timestamp"`
}

type StoreAnomalyScoreResult struct {
	Score             float64
	Average           float64
	StandardDeviation float64
	Error             error
}

type DynamoData struct {
	BucketKey *string         `dynamodbav:"BucketKey"`
	Data      []TelemetryData `dynamodbav:"Data"`
}

type ProcessedData struct {
	Timestamp       string  `json:"timestamp"` // Start of 5s bucket
	FLOWRATE        float64 `json:"flowrate"`
	PRESSURE        float64 `json:"pressure"`
	TEMPERATURE     float64 `json:"temperature"`
	FlowChangeRate  float64 `json:"flow_change_rate"`
	PressChangeRate float64 `json:"press_change_rate"`
	TempChangeRate  float64 `json:"temp_change_rate"`
	AnomalyScore    float64 `json:"anomaly_score"`
	AnomalyLevel    string  `json:"anomaly_level"`
}
