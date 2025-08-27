package internal

import "time"

type QuoteResponse struct {
	Quote string `json:"quote"`
}

type ErrorResponse struct {
	Error     string `json:"error"`
	Timestamp string `json:"timestamp"`
}

func NewErrorResponse(err error) ErrorResponse {
	return ErrorResponse{
		Error:     err.Error(),
		Timestamp: time.Now().Format(time.RFC3339),
	}
}

type Item struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Count      int    `json:"count"`
	CreateDate string `json:"create_date"`
	UpdateDate string `json:"update_date"`
}
