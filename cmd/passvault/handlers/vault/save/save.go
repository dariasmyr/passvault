package save

type Request struct {
	EntryType string `json:"entry_type" validate:"required"`
	EntryData string `json:"entry_data" validate:"required"`
}

type Response struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	ID     int64  `json:"id"`
}
