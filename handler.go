package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

var orders = []Order{}
var nextID = 1

func createOrder(c *gin.Context) {
	var req struct {
		Items []string `json:"items"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order := Order{
		ID:     nextID,
		Items:  req.Items,
		Status: Pending,
	}

	nextID++
	orders = append(orders, order)

	c.JSON(http.StatusOK, order)
}

func getOrders(c *gin.Context) {
	c.JSON(http.StatusOK, orders)
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

	for i := range orders {
		if orders[i].ID == id {
			orders[i].Status = req.Status
			c.JSON(http.StatusOK, orders[i])
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
}