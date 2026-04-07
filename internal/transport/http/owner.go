package httptransport

import (
	"errors"
	"net/http"
	"strconv"

	"breakfast-system/internal/order"
	"breakfast-system/pkg/protocol"

	"github.com/gin-gonic/gin"
)

func (h *Handler) getOwnerOrders(c *gin.Context) {
	c.JSON(http.StatusOK, h.orders.ListOwnerOrders())
}

func (h *Handler) updateOwnerOrderStatus(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	var req protocol.UpdateStatusRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updatedOrder, err := h.orders.UpdateOrderStatus(id, req.Status)
	if err != nil {
		switch {
		case errors.Is(err, order.ErrOrderNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		case errors.Is(err, order.ErrInvalidStatus), errors.Is(err, order.ErrInvalidTransition):
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	event := protocol.Event{
		Type: protocol.EventOrderUpdated,
		Data: updatedOrder,
	}

	h.broker.BroadcastOwner(event)
	h.broker.BroadcastCustomer(updatedOrder.ID, event)
	if updatedOrder.Status == protocol.Ready || updatedOrder.Status == protocol.Done {
		h.broker.BroadcastDisplay(event)
	}

	c.JSON(http.StatusOK, updatedOrder)
}
