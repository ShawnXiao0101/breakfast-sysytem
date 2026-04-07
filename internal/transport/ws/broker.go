package ws

import (
	"log"
	"net/http"
	"strconv"
	"sync"

	"breakfast-system/pkg/protocol"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type Broker struct {
	mu        sync.RWMutex
	owners    map[*websocket.Conn]struct{}
	displays  map[*websocket.Conn]struct{}
	customers map[int]map[*websocket.Conn]struct{}
}

func NewBroker() *Broker {
	return &Broker{
		owners:    make(map[*websocket.Conn]struct{}),
		displays:  make(map[*websocket.Conn]struct{}),
		customers: make(map[int]map[*websocket.Conn]struct{}),
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (b *Broker) ServeOwner(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("upgrade websocket failed: %v", err)
		return
	}

	b.AddOwner(conn)
	defer b.RemoveOwner(conn)

	drainConnection(conn)
}

func (b *Broker) ServeDisplay(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("upgrade websocket failed: %v", err)
		return
	}

	b.AddDisplay(conn)
	defer b.RemoveDisplay(conn)

	drainConnection(conn)
}

func (b *Broker) ServeCustomer(c *gin.Context) {
	orderID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order id"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("upgrade websocket failed: %v", err)
		return
	}

	b.AddCustomer(orderID, conn)
	defer b.RemoveCustomer(orderID, conn)

	drainConnection(conn)
}

func drainConnection(conn *websocket.Conn) {
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			return
		}
	}
}

func (b *Broker) AddOwner(conn *websocket.Conn) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.owners[conn] = struct{}{}
}

func (b *Broker) RemoveOwner(conn *websocket.Conn) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.owners, conn)
	_ = conn.Close()
}

func (b *Broker) AddDisplay(conn *websocket.Conn) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.displays[conn] = struct{}{}
}

func (b *Broker) RemoveDisplay(conn *websocket.Conn) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.displays, conn)
	_ = conn.Close()
}

func (b *Broker) AddCustomer(orderID int, conn *websocket.Conn) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.customers[orderID] == nil {
		b.customers[orderID] = make(map[*websocket.Conn]struct{})
	}
	b.customers[orderID][conn] = struct{}{}
}

func (b *Broker) RemoveCustomer(orderID int, conn *websocket.Conn) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if subscribers, ok := b.customers[orderID]; ok {
		delete(subscribers, conn)
		if len(subscribers) == 0 {
			delete(b.customers, orderID)
		}
	}
	_ = conn.Close()
}

func (b *Broker) BroadcastOwner(event protocol.Event) {
	b.mu.RLock()
	conns := make([]*websocket.Conn, 0, len(b.owners))
	for conn := range b.owners {
		conns = append(conns, conn)
	}
	b.mu.RUnlock()

	b.broadcastToConnections(conns, event, b.RemoveOwner)
}

func (b *Broker) BroadcastDisplay(event protocol.Event) {
	b.mu.RLock()
	conns := make([]*websocket.Conn, 0, len(b.displays))
	for conn := range b.displays {
		conns = append(conns, conn)
	}
	b.mu.RUnlock()

	b.broadcastToConnections(conns, event, b.RemoveDisplay)
}

func (b *Broker) BroadcastCustomer(orderID int, event protocol.Event) {
	b.mu.RLock()
	subscribers := b.customers[orderID]
	conns := make([]*websocket.Conn, 0, len(subscribers))
	for conn := range subscribers {
		conns = append(conns, conn)
	}
	b.mu.RUnlock()

	b.broadcastToConnections(conns, event, func(conn *websocket.Conn) {
		b.RemoveCustomer(orderID, conn)
	})
}

func (b *Broker) broadcastToConnections(conns []*websocket.Conn, event protocol.Event, onFailure func(*websocket.Conn)) {
	for _, conn := range conns {
		if err := conn.WriteJSON(event); err != nil {
			if onFailure != nil {
				onFailure(conn)
				continue
			}
			_ = conn.Close()
		}
	}
}
