package binoculars

type Metric struct {
	Name   string                 `json:"name"`
	Tags   map[string]string      `json:"tags"`
	Fields map[string]interface{} `json:"fields"`
}

type Metrics []Metric

type ServerResponse struct {
	RequestIntervalInMinutes int `json:"requestIntervalInMinutes"`
}
