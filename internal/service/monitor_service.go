package service

import (
	"fmt"
	"strings"
	"stationery-management/internal/domain"
	"stationery-management/internal/repository"
	"stationery-management/pkg/email"
)

type MonitorService struct {
	reqRepo   *repository.RequestRepository
	userRepo  *repository.UserRepository
	emailSvc  *email.EmailService
	auditRepo *repository.AuditRepository
}

func NewMonitorService(reqRepo *repository.RequestRepository, userRepo *repository.UserRepository, emailSvc *email.EmailService, auditRepo *repository.AuditRepository) *MonitorService {
	return &MonitorService{reqRepo: reqRepo, userRepo: userRepo, emailSvc: emailSvc, auditRepo: auditRepo}
}

func (s *MonitorService) SendReminder(dto *domain.SendReminderDTO, actorID uint, actorName, ip string) error {
	req, err := s.reqRepo.FindByID(dto.RequestID)
	if err != nil {
		return err
	}

	var recipients []string
	var subject string
	var message string

	switch dto.Target {
	case "REQUESTER":
		if req.Requester.Email == "" {
			return fmt.Errorf("requester email not found for request %s", req.RequestNo)
		}
		recipients = []string{req.Requester.Email}
		subject = fmt.Sprintf("Reminder: Action required for Request %s", req.RequestNo)
		message = fmt.Sprintf("<p>Dear %s,</p><p>This is a reminder regarding your stationery request <strong>%s</strong> for branch %s.</p><p>%s</p>",
			req.Requester.Name, req.RequestNo, req.Branch.Name, dto.Message)

	case "APPROVER":
		// Look up real APPROVER users from DB — prefer those covering this branch
		approvers, err := s.userRepo.FindByRoleName("APPROVER")
		if err != nil || len(approvers) == 0 {
			return fmt.Errorf("no active approver found in the system to send reminder to")
		}
		for _, a := range approvers {
			// Include approvers that cover all branches OR this specific branch
			if a.ApproverAccessType == "ALL_BRANCHES" {
				recipients = append(recipients, a.Email)
			} else if a.ApproverAccessType == "SINGLE_BRANCH" && a.BranchID != nil && *a.BranchID == req.BranchID {
				recipients = append(recipients, a.Email)
			}
		}
		if len(recipients) == 0 {
			// Fallback: email all approvers if none matched branch
			for _, a := range approvers {
				recipients = append(recipients, a.Email)
			}
		}
		subject = fmt.Sprintf("Reminder: Pending Approval for Request %s", req.RequestNo)
		message = fmt.Sprintf("<p>Dear Approver,</p><p>Stationery request <strong>%s</strong> from branch <strong>%s</strong> is awaiting your approval.</p><p>%s</p>",
			req.RequestNo, req.Branch.Name, dto.Message)

	case "AGENCY":
		// Look up all active AGENCY users from DB
		agencies, err := s.userRepo.FindByRoleName("AGENCY")
		if err != nil || len(agencies) == 0 {
			return fmt.Errorf("no active delivery agency found in the system to send reminder to")
		}
		for _, a := range agencies {
			recipients = append(recipients, a.Email)
		}
		subject = fmt.Sprintf("Reminder: Delivery Pending for Request %s", req.RequestNo)
		message = fmt.Sprintf("<p>Dear Delivery Agency,</p><p>Approved request <strong>%s</strong> from branch <strong>%s</strong> is pending fulfillment and delivery.</p><p>%s</p>",
			req.RequestNo, req.Branch.Name, dto.Message)
	}

	if len(recipients) == 0 {
		return fmt.Errorf("no recipient email resolved for target '%s'", dto.Target)
	}

	// Send to all resolved recipients
	var sendErrors []string
	for _, to := range recipients {
		if err := s.emailSvc.SendEmail(to, subject, message); err != nil {
			sendErrors = append(sendErrors, fmt.Sprintf("%s: %v", to, err))
		}
	}
	if len(sendErrors) > 0 {
		return fmt.Errorf("reminder send errors: %s", strings.Join(sendErrors, "; "))
	}

	s.auditRepo.Create(&domain.AuditLog{
		UserID:     &actorID,
		UserName:   actorName,
		Action:     "SEND_REMINDER_EMAIL",
		EntityType: "REQUEST",
		EntityID:   req.RequestNo,
		IPAddress:  ip,
	})

	return nil
}

