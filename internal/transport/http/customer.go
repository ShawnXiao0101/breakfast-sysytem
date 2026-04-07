package httptransport

import (
	"errors"
	"net/http"
	"strconv"

	"breakfast-system/internal/order"
	"breakfast-system/pkg/protocol"

	"github.com/gin-gonic/gin"
)

func (h *Handler) createCustomerOrder(c *gin.Context) {
	var req protocol.CreateOrderRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	newOrder, err := h.orders.CreateOrder(req.Items)
	if err != nil {
		status := http.StatusBadRequest
		if !errors.Is(err, order.ErrEmptyItems) {
			status = http.StatusInternalServerError
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	h.broker.BroadcastOwner(protocol.Event{
		Type: protocol.EventOrderCreated,
		Data: newOrder,
	})

	c.JSON(http.StatusOK, newOrder)
}

func (h *Handler) getCustomerOrder(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	foundOrder, err := h.orders.GetOrder(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, foundOrder)
}
