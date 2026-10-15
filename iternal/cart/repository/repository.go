package repository

import (
	"sync"

	"github.com/El-Donte/microservices-grps/iternal/cart/domain"
)

type Repository struct {
	mu       sync.RWMutex
	database map[uint64]*domain.Cart
}

func NewRepository() *Repository {
	return &Repository{database: make(map[uint64]*domain.Cart)}
}

func (repo *Repository) Get(userId uint64) (*domain.Cart, error) {
	repo.mu.RLock()
	defer repo.mu.RUnlock()

	if cart, ok := repo.database[userId]; ok {
		return cart, nil
	}
	return nil, domain.CartNotFound
}

func (repo *Repository) Create(userId uint64) *domain.Cart {
	repo.mu.Lock()
	defer repo.mu.Unlock()

	cart := domain.NewCart()
	repo.database[userId] = cart
	return cart
}
