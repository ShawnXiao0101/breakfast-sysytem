package main

type Status string

const (
	Pending Status = "待接单"
	Cooking Status = "制作中"
	Ready   Status = "待取餐"
	Done    Status = "已完成"
)

type Order struct {
	ID     int      `json:"id"`
	Items  []string `json:"items"`
	Status Status   `json:"status"`
}

type EventType string

const (
	EventOrderCreated EventType = "order_created"
	EventOrderUpdated EventType = "order_updated"
)

type Event struct {
	Type EventType `json:"type"`
	Data Order     `json:"data"`
}
