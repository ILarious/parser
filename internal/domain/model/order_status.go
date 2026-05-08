package model

type OrderStatus int

const (
	OrderStatusNew OrderStatus = iota
	OrderStatusProcessing
	OrderStatusDone
	OrderStatusFailed
)
