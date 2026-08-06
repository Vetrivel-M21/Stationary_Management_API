package service

import (
	"errors"
	"fmt"
	"stationery-management/internal/domain"
	"stationery-management/internal/repository"
	"time"
)

type ChatService interface {
	GetChatMessages(requestID uint) ([]domain.ChatMessage, error)
	SendChatMessage(requestID uint, senderID uint, messageText string) (*domain.ChatMessage, error)
}

type chatService struct {
	chatRepo  repository.ChatRepository
	reqRepo   *repository.RequestRepository
	userRepo  *repository.UserRepository
	auditRepo *repository.AuditRepository
}

func NewChatService(
	chatRepo repository.ChatRepository,
	reqRepo *repository.RequestRepository,
	userRepo *repository.UserRepository,
	auditRepo *repository.AuditRepository,
) ChatService {
	return &chatService{
		chatRepo:  chatRepo,
		reqRepo:   reqRepo,
		userRepo:  userRepo,
		auditRepo: auditRepo,
	}
}

func (s *chatService) GetChatMessages(requestID uint) ([]domain.ChatMessage, error) {
	_, err := s.reqRepo.FindByID(requestID)
	if err != nil {
		return nil, errors.New("request not found")
	}
	return s.chatRepo.GetChatMessagesByRequestID(requestID)
}

func (s *chatService) SendChatMessage(requestID uint, senderID uint, messageText string) (*domain.ChatMessage, error) {
	sender, err := s.userRepo.FindByID(senderID)
	if err != nil {
		return nil, errors.New("sender user not found")
	}

	req, err := s.reqRepo.FindByID(requestID)
	if err != nil {
		return nil, errors.New("request not found")
	}

	senderRole := sender.Role.Name
	targetRole := "APPROVER"

	switch req.Status {
	case "SUBMITTED":
		if senderRole == "MONITOR" {
			targetRole = "APPROVER"
		} else {
			targetRole = "MONITOR"
		}
	case "APPROVED", "PARTIALLY_DELIVERED":
		if senderRole == "MONITOR" {
			targetRole = "AGENCY"
		} else {
			targetRole = "MONITOR"
		}
	case "DELIVERED":
		if senderRole == "MONITOR" {
			targetRole = "BRANCH_REQUESTER"
		} else {
			targetRole = "MONITOR"
		}
	default:
		targetRole = "MONITOR"
	}

	msg := &domain.ChatMessage{
		RequestID:  requestID,
		SenderID:   sender.ID,
		SenderName: sender.Name,
		SenderRole: senderRole,
		TargetRole: targetRole,
		Message:    messageText,
		CreatedAt:  time.Now(),
	}

	if err := s.chatRepo.CreateChatMessage(msg); err != nil {
		return nil, err
	}

	s.auditRepo.Create(&domain.AuditLog{
		UserID:     &sender.ID,
		UserName:   sender.Name,
		Action:     "SEND_CHAT_MESSAGE",
		EntityType: "REQUEST",
		EntityID:   fmt.Sprintf("%d", requestID),
		Details:    fmt.Sprintf("Sent chat message to %s regarding Request #%s", targetRole, req.RequestNo),
	})

	return msg, nil
}
