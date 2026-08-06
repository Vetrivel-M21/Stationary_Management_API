package repository

import (
	"stationery-management/internal/domain"

	"gorm.io/gorm"
)

type BranchRepository struct {
	db *gorm.DB
}

func NewBranchRepository(db *gorm.DB) *BranchRepository {
	return &BranchRepository{db: db}
}

func (r *BranchRepository) FindByID(id uint) (*domain.Branch, error) {
	var branch domain.Branch
	err := r.db.Where("id = ?", id).First(&branch).Error
	if err != nil {
		return nil, err
	}
	return &branch, nil
}

func (r *BranchRepository) FindAll(search string, page, limit int) ([]domain.Branch, int64, error) {
	var branches []domain.Branch
	var total int64

	query := r.db.Model(&domain.Branch{})
	if search != "" {
		s := "%" + search + "%"
		query = query.Where("name LIKE ? OR code LIKE ?", s, s)
	}

	query.Count(&total)
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("id desc").Find(&branches).Error
	return branches, total, err
}

func (r *BranchRepository) Create(branch *domain.Branch) error {
	return r.db.Create(branch).Error
}

func (r *BranchRepository) Update(branch *domain.Branch) error {
	return r.db.Save(branch).Error
}
