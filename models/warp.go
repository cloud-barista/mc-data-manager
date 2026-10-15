package models

// WarpDiagnosticRequest is the request body for the object storage warp diagnose endpoint.
type WarpDiagnosticRequest struct {
	ProviderParams
	RegionParams
	BucketId    string `json:"bucketId" form:"bucketId"`
	DurationSec int64  `json:"durationSec" form:"durationSec"`
	ObjectKib   int64  `json:"objectKib" form:"objectKib"` // in KB
	ObjectCount int64  `json:"objectCount" form:"objectCount"`
}

// WarpOperationReport holds the parsed "Report: <OP>" block for one operation
// type (DELETE, GET, PUT, STAT). Fields that don't apply to a given operation
// (size/throughput for DELETE/STAT, TTFB for DELETE/STAT) are left zero.
type WarpOperationReport struct {
	Operation     string `json:"operation"`
	Requests      int64  `json:"requests"`
	DurationSec   int64  `json:"durationSec"`
	ObjectsPerReq int64  `json:"objectsPerReq"`
	Size          string `json:"size"` // raw, e.g. "65536 bytes"; empty if not applicable
	Concurrency   int64  `json:"concurrency"`

	AvgThroughput string  `json:"avgThroughput"` // raw, e.g. "3.45 MiB/s"; empty if not applicable
	AvgObjPS      float64 `json:"avgObjPS"`

	ReqAvgMs     float64 `json:"reqAvgMs"`
	ReqP50Ms     float64 `json:"reqP50Ms"`
	ReqP90Ms     float64 `json:"reqP90Ms"`
	ReqP99Ms     float64 `json:"reqP99Ms"`
	ReqFastestMs float64 `json:"reqFastestMs"`
	ReqSlowestMs float64 `json:"reqSlowestMs"`
	ReqStdDevMs  float64 `json:"reqStdDevMs"`

	TTFBAvgMs    float64 `json:"ttfbAvgMs"`
	TTFBBestMs   float64 `json:"ttfbBestMs"`
	TTFBP25Ms    float64 `json:"ttfbP25Ms"`
	TTFBMedianMs float64 `json:"ttfbMedianMs"`
	TTFBP75Ms    float64 `json:"ttfbP75Ms"`
	TTFBP90Ms    float64 `json:"ttfbP90Ms"`
	TTFBP99Ms    float64 `json:"ttfbP99Ms"`
	TTFBWorstMs  float64 `json:"ttfbWorstMs"`
	TTFBStdDevMs float64 `json:"ttfbStdDevMs"`

	ThroughputFastestObjPS float64 `json:"throughputFastestObjPS"`
	ThroughputMedianObjPS  float64 `json:"throughputMedianObjPS"`
	ThroughputSlowestObjPS float64 `json:"throughputSlowestObjPS"`
	ThroughputFastest      string  `json:"throughputFastest"` // raw, e.g. "3.9MiB/s"; empty if not applicable
	ThroughputMedian       string  `json:"throughputMedian"`
	ThroughputSlowest      string  `json:"throughputSlowest"`
}

// WarpParsed is the parsed result of a warp mixed benchmark run.
type WarpParsed struct {
	Operations []WarpOperationReport `json:"operations"` // DELETE, GET, PUT, STAT
	Raw        string                `json:"raw"`
}
