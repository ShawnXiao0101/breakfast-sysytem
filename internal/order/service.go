package order

import (
	"errors"
	"strings"

	"breakfast-system/pkg/protocol"
)

var (
	ErrEmptyItems        = errors.New("items cannot be empty")
	ErrOrderNotFound     = errors.New("order not found")
	ErrInvalidStatus     = errors.New("invalid status")
	ErrInvalidTransition = errors.New("invalid status transition")
)

type Service struct {
	store *Store
}

func NewService(store *Store) *Service {
	return &Service{store: store}
}

func (s *Service) CreateOrder(items []string) (protocol.Order, error) {
	cleaned := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		cleaned = append(cleaned, item)
	}

	if len(cleaned) == 0 {
		return protocol.Order{}, ErrEmptyItems
	}

	return s.store.Create(cleaned), nil
}

func (s *Service) GetOrder(id int) (protocol.Order, error) {
	order, ok := s.store.Get(id)
	if !ok {
		return protocol.Order{}, ErrOrderNotFound
	}

	return order, nil
}

func (s *Service) ListOwnerOrders() []protocol.Order {
	return s.store.ListAll()
}

func (s *Service) ListDisplayOrders() []protocol.Order {
	all := s.store.ListAll()
	visible := make([]protocol.Order, 0, len(all))
	for _, order := range all {
		if order.Status == protocol.Ready || order.Status == protocol.Done {
			visible = append(visible, order)
		}
	}
	return visible
}

func (s *Service) UpdateOrderStatus(id int, next protocol.Status) (protocol.Order, error) {
	if !next.IsValid() {
		return protocol.Order{}, ErrInvalidStatus
	}

	current, ok := s.store.Get(id)
	if !ok {
		return protocol.Order{}, ErrOrderNotFound
	}

	if !canTransition(current.Status, next) {
		return protocol.Order{}, ErrInvalidTransition
	}

	updated, _ := s.store.UpdateStatus(id, next)
	return updated, nil
}

func canTransition(from, to protocol.Status) bool {
	switch from {
	case protocol.Pending:
		return to == protocol.Cooking
	case protocol.Cooking:
		return to == protocol.Ready
	case protocol.Ready:
		return to == protocol.Done
	case protocol.Done:
		return false
	default:
		return false
	}
}
