package httptransport

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) getDisplayOrders(c *gin.Context) {
	c.JSON(http.StatusOK, h.orders.ListDisplayOrders())
}
