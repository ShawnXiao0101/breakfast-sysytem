package order

import (
	"sync"

	"breakfast-system/pkg/protocol"
)

type Store struct {
	mu     sync.RWMutex
	nextID int
	orders map[int]protocol.Order
}

func NewStore() *Store {
	return &Store{
		nextID: 1,
		orders: make(map[int]protocol.Order),
	}
}

func (s *Store) Create(items []string) protocol.Order {
	s.mu.Lock()
	defer s.mu.Unlock()

	order := protocol.Order{
		ID:     s.nextID,
		Items:  append([]string(nil), items...),
		Status: protocol.Pending,
	}

	s.orders[order.ID] = order
	s.nextID++

	return order
}

func (s *Store) Get(id int) (protocol.Order, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	order, ok := s.orders[id]
	if !ok {
		return protocol.Order{}, false
	}

	order.Items = append([]string(nil), order.Items...)
	return order, true
}

func (s *Store) ListAll() []protocol.Order {
	s.mu.RLock()
	defer s.mu.RUnlock()

	orders := make([]protocol.Order, 0, len(s.orders))
	for i := 1; i < s.nextID; i++ {
		order, ok := s.orders[i]
		if !ok {
			continue
		}
		order.Items = append([]string(nil), order.Items...)
		orders = append(orders, order)
	}

	return orders
}

func (s *Store) UpdateStatus(id int, status protocol.Status) (protocol.Order, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, ok := s.orders[id]
	if !ok {
		return protocol.Order{}, false
	}

	order.Status = status
	s.orders[id] = order
	order.Items = append([]string(nil), order.Items...)
	return order, true
}
