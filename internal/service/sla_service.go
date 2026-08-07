package service

import (
	"fmt"
	"math"
	"stationery-management/internal/domain"
	"stationery-management/internal/repository"
	"time"
)

type SlaService interface {
	GetSlaSettings() (*domain.SlaSettings, error)
	UpdateSlaSettings(actorID uint, actorEmail, ip string, input domain.UpdateSlaSettingsDTO) (*domain.SlaSettings, error)
	GetDelayedOrders(departmentFilter string) ([]domain.DelayedOrderDTO, error)
}

type slaService struct {
	slaRepo   repository.SlaRepository
	reqRepo   *repository.RequestRepository
	auditRepo *repository.AuditRepository
}

func NewSlaService(
	slaRepo repository.SlaRepository,
	reqRepo *repository.RequestRepository,
	auditRepo *repository.AuditRepository,
) SlaService {
	return &slaService{
		slaRepo:   slaRepo,
		reqRepo:   reqRepo,
		auditRepo: auditRepo,
	}
}

func (s *slaService) GetSlaSettings() (*domain.SlaSettings, error) {
	return s.slaRepo.GetSlaSettings()
}

func (s *slaService) UpdateSlaSettings(actorID uint, actorEmail, ip string, input domain.UpdateSlaSettingsDTO) (*domain.SlaSettings, error) {
	settings, err := s.slaRepo.GetSlaSettings()
	if err != nil {
		return nil, err
	}

	settings.MaxApproveDays = input.MaxApproveDays
	settings.MaxDeliveryDays = input.MaxDeliveryDays
	settings.MaxVerifyDays = input.MaxVerifyDays

	if err := s.slaRepo.UpdateSlaSettings(settings); err != nil {
		return nil, err
	}

	s.auditRepo.Create(&domain.AuditLog{
		UserID:     &actorID,
		UserName:   actorEmail,
		Action:     "UPDATE_SLA_SETTINGS",
		EntityType: "SYSTEM",
		EntityID:   "1",
		IPAddress:  ip,
		Details:    fmt.Sprintf("Updated SLA Max Days: Approve=%d, Delivery=%d, Verify=%d", input.MaxApproveDays, input.MaxDeliveryDays, input.MaxVerifyDays),
	})

	return settings, nil
}

func (s *slaService) GetDelayedOrders(departmentFilter string) ([]domain.DelayedOrderDTO, error) {
	settings, err := s.slaRepo.GetSlaSettings()
	if err != nil {
		return nil, err
	}

	requests, _, err := s.reqRepo.FindAll(nil, nil, departmentFilter, "", 1, 1000)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	var delayedOrders []domain.DelayedOrderDTO

	for _, req := range requests {
		switch req.Status {
		case "SUBMITTED":
			daysElapsed := int(math.Floor(now.Sub(req.SubmittedAt).Hours() / 24))
			if daysElapsed > settings.MaxApproveDays {
				delayedOrders = append(delayedOrders, domain.DelayedOrderDTO{
					Request:        req,
					DelayDays:      daysElapsed - settings.MaxApproveDays,
					DelayedStage:   "Pending Approval",
					TargetRole:     "APPROVER",
					MaxAllowedDays: settings.MaxApproveDays,
				})
			}
		case "APPROVED", "PARTIALLY_DELIVERED":
			refTime := req.SubmittedAt
			if req.ApprovedAt != nil {
				refTime = *req.ApprovedAt
			}
			daysElapsed := int(math.Floor(now.Sub(refTime).Hours() / 24))
			if daysElapsed > settings.MaxDeliveryDays {
				delayedOrders = append(delayedOrders, domain.DelayedOrderDTO{
					Request:        req,
					DelayDays:      daysElapsed - settings.MaxDeliveryDays,
					DelayedStage:   "Pending Delivery",
					TargetRole:     "AGENCY",
					MaxAllowedDays: settings.MaxDeliveryDays,
				})
			}
		case "DELIVERED":
			refTime := req.SubmittedAt
			if len(req.Deliveries) > 0 {
				refTime = req.Deliveries[len(req.Deliveries)-1].DeliveredDate
			}
			daysElapsed := int(math.Floor(now.Sub(refTime).Hours() / 24))
			if daysElapsed > settings.MaxVerifyDays {
				delayedOrders = append(delayedOrders, domain.DelayedOrderDTO{
					Request:        req,
					DelayDays:      daysElapsed - settings.MaxVerifyDays,
					DelayedStage:   "Pending Verification",
					TargetRole:     "BRANCH_REQUESTER",
					MaxAllowedDays: settings.MaxVerifyDays,
				})
			}
		}
	}

	return delayedOrders, nil
}
