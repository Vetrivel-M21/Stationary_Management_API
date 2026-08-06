package service

import (
	"errors"
	"stationery-management/internal/domain"
	"stationery-management/internal/repository"
)

type ProductService struct {
	productRepo *repository.ProductRepository
	auditRepo   *repository.AuditRepository
}

func NewProductService(productRepo *repository.ProductRepository, auditRepo *repository.AuditRepository) *ProductService {
	return &ProductService{productRepo: productRepo, auditRepo: auditRepo}
}

func (s *ProductService) GetAllProducts(search, category string, page, limit int) ([]domain.Product, int64, error) {
	return s.productRepo.FindAll(search, category, page, limit)
}

func (s *ProductService) CreateProduct(req *domain.CreateProductRequest, actorID uint, actorName, ip string) (*domain.Product, error) {
	product := &domain.Product{
		Name:        req.Name,
		Category:    req.Category,
		Unit:        req.Unit,
		UnitPrice:   req.UnitPrice,
		Description: req.Description,
		Status:      "ACTIVE",
	}

	if err := s.productRepo.Create(product); err != nil {
		return nil, err
	}

	s.auditRepo.Create(&domain.AuditLog{
		UserID:     &actorID,
		UserName:   actorName,
		Action:     "CREATE_PRODUCT",
		EntityType: "PRODUCT",
		EntityID:   product.Name,
		IPAddress:  ip,
	})

	return product, nil
}

func (s *ProductService) UpdateProduct(id uint, req *domain.UpdateProductRequest, actorID uint, actorName, ip string) (*domain.Product, error) {
	product, err := s.productRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("product not found")
	}

	if req.Name != "" {
		product.Name = req.Name
	}
	if req.Category != "" {
		product.Category = req.Category
	}
	if req.Unit != "" {
		product.Unit = req.Unit
	}
	if req.UnitPrice >= 0 {
		product.UnitPrice = req.UnitPrice
	}
	if req.Description != "" {
		product.Description = req.Description
	}
	if req.Status != "" {
		product.Status = req.Status
	}

	if err := s.productRepo.Update(product); err != nil {
		return nil, err
	}

	s.auditRepo.Create(&domain.AuditLog{
		UserID:     &actorID,
		UserName:   actorName,
		Action:     "UPDATE_PRODUCT",
		EntityType: "PRODUCT",
		EntityID:   product.Name,
		IPAddress:  ip,
	})

	return product, nil
}

func (s *ProductService) DeleteProduct(id uint, actorID uint, actorName, ip string) error {
	product, err := s.productRepo.FindByID(id)
	if err != nil {
		return errors.New("product not found")
	}

	if err := s.productRepo.Delete(id); err != nil {
		return err
	}

	s.auditRepo.Create(&domain.AuditLog{
		UserID:     &actorID,
		UserName:   actorName,
		Action:     "SOFT_DELETE_PRODUCT",
		EntityType: "PRODUCT",
		EntityID:   product.Name,
		IPAddress:  ip,
	})

	return nil
}
