package repository

import (
	"stationery-management/internal/domain"

	"gorm.io/gorm"
)

type ChatRepository interface {
	CreateChatMessage(msg *domain.ChatMessage) error
	GetChatMessagesByRequestID(requestID uint) ([]domain.ChatMessage, error)
}

type chatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) CreateChatMessage(msg *domain.ChatMessage) error {
	return r.db.Create(msg).Error
}

func (r *chatRepository) GetChatMessagesByRequestID(requestID uint) ([]domain.ChatMessage, error) {
	var messages []domain.ChatMessage
	err := r.db.Where("request_id = ?", requestID).Order("created_at ASC").Find(&messages).Error
	return messages, err
}
