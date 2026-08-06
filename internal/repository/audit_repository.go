package repository

import (
	"stationery-management/internal/domain"

	"gorm.io/gorm"
)

type AuditRepository struct {
	db *gorm.DB
}

func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

func (r *AuditRepository) Create(log *domain.AuditLog) error {
	if r.db == nil {
		return nil
	}
	return r.db.Create(log).Error
}

func (r *AuditRepository) FindAll(page, limit int) ([]domain.AuditLog, int64, error) {
	var logs []domain.AuditLog
	var total int64

	query := r.db.Model(&domain.AuditLog{})
	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("id desc").Find(&logs).Error
	return logs, total, err
}
