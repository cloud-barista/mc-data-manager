package models

// YcsbDiagnosticRequest is the request body for the NRDB go-ycsb diagnose endpoint.
// Host/Port/Username/Password are only used for ncp/alibaba (mongodb); Region is
// only used for aws (dynamodb), whose credentials are resolved server-side.
type YcsbDiagnosticRequest struct {
	ProviderParams
	RegionParams
	Host     string `json:"host" form:"host"`
	Port     string `json:"port" form:"port"`
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`

	RecordCount      int64   `json:"recordCount" form:"recordCount"` // records inserted during the load phase
	ThreadCount      int64   `json:"threadCount" form:"threadCount"`
	ReadProportion   float64 `json:"readProportion" form:"readProportion"`     // run phase, 0~1
	UpdateProportion float64 `json:"updateProportion" form:"updateProportion"` // run phase, 0~1
}

// YcsbOperationReport holds one operation's final summary line printed after
// "Run finished" in go-ycsb output.
type YcsbOperationReport struct {
	Operation string  `json:"operation"` // INSERT (load) / READ, UPDATE (run)
	TakesSec  float64 `json:"takesSec"`
	Count     int64   `json:"count"`
	OPS       float64 `json:"ops"`
	AvgMs     float64 `json:"avgMs"`
	MinMs     float64 `json:"minMs"`
	MaxMs     float64 `json:"maxMs"`
	P50Ms     float64 `json:"p50Ms"`
	P90Ms     float64 `json:"p90Ms"`
	P95Ms     float64 `json:"p95Ms"`
	P99Ms     float64 `json:"p99Ms"`
	P999Ms    float64 `json:"p999Ms"`  // 99.9th
	P9999Ms   float64 `json:"p9999Ms"` // 99.99th
}

// YcsbParsed is the parsed result of a go-ycsb load+run diagnose. TOTAL and
// *_ERROR summary lines are excluded from Operations but remain in Raw.
type YcsbParsed struct {
	Operations []YcsbOperationReport `json:"operations"` // INSERT (load), READ/UPDATE (run)
	Raw        string                `json:"raw"`
}
