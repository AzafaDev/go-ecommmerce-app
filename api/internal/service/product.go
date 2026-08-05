package service

import "go-ecommerce-app/internal/repository"

type ProductService struct {
	repo repository.Querier
}

func NewProductService(repo repository.Querier) *ProductService {
	return &ProductService{
		repo: repo,
	}
}
