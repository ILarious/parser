package model

import "time"

type ParsingStatus int

const (
	ParsingStatusNew ParsingStatus = iota
	ParsingStatusProcessing
	ParsingStatusDone
	ParsingStatusFailed
)

type ParsingOrders struct {
	ID        int64
	EventID   int64
	OrderID   int64
	Username  string
	Status    ParsingStatus
	ErrorText string
	AccountID int64
	CreatedAt time.Time
	UpdatedAt time.Time
}
