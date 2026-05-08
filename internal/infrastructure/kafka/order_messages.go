package kafka

import "strconv"

type OrderRequest struct {
	EventID  int64  `json:"event_id"`
	OrderID  int64  `json:"order_id"`
	Username string `json:"username"`
}

type OrderResponse struct {
	EventID        string `json:"event_id"`
	OrderID        int64  `json:"order_id"`
	Username       string `json:"username"`
	FullName       string `json:"full_name"`
	FollowersCount int    `json:"followers_count"`
	Status         int    `json:"status"`
}

func ResponseEventID(orderID int64, status int) string {
	return strconv.FormatInt(orderID, 10) + ":" + strconv.Itoa(status)
}
