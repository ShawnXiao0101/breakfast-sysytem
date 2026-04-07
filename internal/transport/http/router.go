package httptransport

import (
	"breakfast-system/internal/order"
	"breakfast-system/internal/transport/ws"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	orders *order.Service
	broker *ws.Broker
}

func NewRouter(broker *ws.Broker) *gin.Engine {
	service := order.NewService(order.NewStore())
	handler := &Handler{
		orders: service,
		broker: broker,
	}

	r := gin.Default()

	customer := r.Group("/customer")
	{
		customer.POST("/orders", handler.createCustomerOrder)
		customer.GET("/orders/:id", handler.getCustomerOrder)
		customer.GET("/ws/orders/:id", broker.ServeCustomer)
	}

	owner := r.Group("/owner")
	{
		owner.GET("/orders", handler.getOwnerOrders)
		owner.POST("/orders/:id/status", handler.updateOwnerOrderStatus)
		owner.GET("/ws", broker.ServeOwner)
	}

	display := r.Group("/display")
	{
		display.GET("/orders", handler.getDisplayOrders)
		display.GET("/ws", broker.ServeDisplay)
	}

	return r
}
