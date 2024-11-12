package save

import (
	"context"
	resp "passvault/internal/lib/api/response"
)

type Request struct {
	EntryType string `json:"entry_type" validate:"required"`
	EntryData string `json:"entry_data" validate:"required"`
}

type Response struct {
	resp.Response
	ID int64 `json:"id"`
}

type EntrySaver interface {
	SaveEntry(ctx context.Context, entryType, entryData string) (int64, error)
}
