package models

// NRDBTableRequest is the request body for NRDBMS table create/delete endpoints.
type NRDBTableRequest struct {
	TargetPoint ProviderConfig `json:"targetPoint"`
	TableName   string         `json:"tableName"`
}

// NRDBTableListResponse is the response body for the NRDBMS table listing endpoint.
type NRDBTableListResponse struct {
	Tables []string `json:"tables"`
}
