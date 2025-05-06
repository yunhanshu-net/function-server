package db

import (
	"context"
	"gorm.io/gorm"
)

type OrderDTO struct {
	UserID string
	Amount string
	Status string
}

type Order struct {
	UserID string
	Amount string
	Status string
}

type S struct {
	db   *gorm.DB
	repo *repo
}

type repo struct {
	db *gorm.DB
}

func (r *repo) Create(ctx context.Context, tree *OrderDTO) error {
	return GetContextDB(ctx, r.db).Create(tree).Error
}

// service/user_order_service.go
func (s *S) CreateOrderWithPayment(ctx context.Context, orderDTO *OrderDTO) error {
	txManager := NewTxManager(s.db)
	return txManager.WithTransaction(ctx, func(txCtx context.Context) error {

		err := s.repo.Create(txCtx, orderDTO)
		if err != nil {
			return err
		}
		return nil
	})
}
