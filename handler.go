package main

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

var orders = []Order{}
var nextID = 1
var ordersMu sync.Mutex

func createOrder(c *gin.Context) {
	var req struct {
		Items []string `json:"items"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ordersMu.Lock()
	order := Order{
		ID:     nextID,
		Items:  req.Items,
		Status: Pending,
	}

	nextID++
	orders = append(orders, order)
	ordersMu.Unlock()

	hub.Broadcast(Event{
		Type: EventOrderCreated,
		Data: order,
	})

	c.JSON(http.StatusOK, order)
}

func getOrders(c *gin.Context) {
	ordersMu.Lock()
	ordersSnapshot := append([]Order(nil), orders...)
	ordersMu.Unlock()

	c.JSON(http.StatusOK, ordersSnapshot)
}

func updateStatus(c *gin.Context) {
	idStr := c.Param("id")
	id, _ := strconv.Atoi(idStr)

	var req struct {
		Status Status `json:"status"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ordersMu.Lock()
	for i := range orders {
		if orders[i].ID == id {
			orders[i].Status = req.Status
			updatedOrder := orders[i]
			ordersMu.Unlock()

			hub.Broadcast(Event{
				Type: EventOrderUpdated,
				Data: updatedOrder,
			})

			c.JSON(http.StatusOK, updatedOrder)
			return
		}
	}
	ordersMu.Unlock()

	c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
}
