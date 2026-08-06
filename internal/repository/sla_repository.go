package repository

import (
	"stationery-management/internal/domain"
	"time"

	"gorm.io/gorm"
)

type SlaRepository interface {
	GetSlaSettings() (*domain.SlaSettings, error)
	UpdateSlaSettings(settings *domain.SlaSettings) error
}

type slaRepository struct {
	db *gorm.DB
}

func NewSlaRepository(db *gorm.DB) SlaRepository {
	return &slaRepository{db: db}
}

func (r *slaRepository) GetSlaSettings() (*domain.SlaSettings, error) {
	var settings domain.SlaSettings
	err := r.db.First(&settings, 1).Error
	if err != nil {
		// Return default settings if none found
		return &domain.SlaSettings{
			ID:              1,
			MaxApproveDays:  2,
			MaxDeliveryDays: 3,
			MaxVerifyDays:   2,
			UpdatedAt:       time.Now(),
		}, nil
	}
	return &settings, nil
}

func (r *slaRepository) UpdateSlaSettings(settings *domain.SlaSettings) error {
	settings.ID = 1
	settings.UpdatedAt = time.Now()
	return r.db.Save(settings).Error
}
