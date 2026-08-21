package service

import (
	"errors"
	"stationery-management/internal/domain"
	"stationery-management/internal/repository"
	"time"
)

type RequestService struct {
	reqRepo   *repository.RequestRepository
	userRepo  *repository.UserRepository
	auditRepo *repository.AuditRepository
}

func NewRequestService(reqRepo *repository.RequestRepository, userRepo *repository.UserRepository, auditRepo *repository.AuditRepository) *RequestService {
	return &RequestService{reqRepo: reqRepo, userRepo: userRepo, auditRepo: auditRepo}
}

func (s *RequestService) CreateRequest(requester *domain.User, dto *domain.CreateRequestDTO, actorName, ip string) (*domain.Request, error) {
	requesterID := requester.ID
	reqNo := s.reqRepo.GenerateUniqueRequestNo()

	department := dto.Department
	if requester.Role.Name == "BRANCH_REQUESTER" {
		if requester.Department == "" {
			return nil, errors.New("Your account has no department assigned. Contact an admin.")
		}
		department = requester.Department
	}

	var items []domain.RequestItem
	for _, item := range dto.Items {
		items = append(items, domain.RequestItem{
			ProductID:    item.ProductID,
			RequestedQty: item.RequestedQty,
			UnitPrice:    item.UnitPrice,
		})
	}

	targetBranchID := dto.BranchID
	if targetBranchID == 0 && dto.BranchName != "" {
		b, err := s.reqRepo.FindOrCreateBranchByName(dto.BranchName)
		if err == nil && b != nil {
			targetBranchID = b.ID
		}
	}
	if targetBranchID == 0 {
		targetBranchID = 1
	}

	req := &domain.Request{
		RequestNo:       reqNo,
		BranchID:        targetBranchID,
		RequesterID:     requesterID,
		ApplicantName:   dto.ApplicantName,
		ApplicantMobile: dto.ApplicantMobile,
		ApplicantEmail:  dto.ApplicantEmail,
		Department:      department,
		Location:        dto.Location,
		Status:          "SUBMITTED",
		Items:           items,
		SubmittedAt:     time.Now(),
	}

	if err := s.reqRepo.Create(req); err != nil {
		return nil, err
	}

	s.auditRepo.Create(&domain.AuditLog{
		UserID:     &requesterID,
		UserName:   actorName,
		Action:     "SUBMIT_REQUEST",
		EntityType: "REQUEST",
		EntityID:   reqNo,
		IPAddress:  ip,
	})

	return s.reqRepo.FindByID(req.ID)
}

func (s *RequestService) GetRequests(user *domain.User, status string, page, limit int) ([]domain.Request, int64, error) {
	var branchID *uint
	var requesterID *uint
	var department string

	if user.Role.Name == "BRANCH_REQUESTER" {
		requesterID = &user.ID
	} else if user.Role.Name == "APPROVER" {
		if user.ApproverAccessType == "SINGLE_BRANCH" && user.BranchID != nil {
			branchID = user.BranchID
		}
		if user.Department != "" {
			department = user.Department
		}
	} else if user.Role.Name == "MONITOR" {
		if user.Department != "" {
			department = user.Department
		}
	}

	return s.reqRepo.FindAll(branchID, requesterID, department, status, page, limit)
}

func (s *RequestService) GetRequestByID(id uint) (*domain.Request, error) {
	return s.reqRepo.FindByID(id)
}

func (s *RequestService) ProcessApproval(requestID uint, approverID uint, dto *domain.ProcessApprovalDTO, actorName, ip string) (*domain.Request, error) {
	req, err := s.reqRepo.FindByID(requestID)
	if err != nil {
		return nil, errors.New("request not found")
	}

	if req.Status != "SUBMITTED" {
		return nil, errors.New("request cannot be processed in its current status")
	}

	now := time.Now()
	if dto.Action == "REJECT" {
		req.Status = "REJECTED"
		req.ApprovedAt = &now
		if err := s.reqRepo.Update(req); err != nil {
			return nil, err
		}
	} else {
		// Action == APPROVE
		req.Status = "APPROVED"
		req.ApprovedAt = &now

		var approvalItems []domain.ApprovalItem
		for _, item := range req.Items {
			approvedQty := item.RequestedQty
			remarks := ""

			// Check if overridden in DTO
			for _, dtoItem := range dto.Items {
				if dtoItem.RequestItemID == item.ID {
					if dtoItem.Remove {
						approvedQty = 0
					} else {
						approvedQty = dtoItem.ApprovedQty
					}
					remarks = dtoItem.Remarks
					break
				}
			}

			approvalItems = append(approvalItems, domain.ApprovalItem{
				RequestItemID: item.ID,
				ApprovedQty:   approvedQty,
				ApprovedBy:    approverID,
				Remarks:       remarks,
			})
		}

		if err := s.reqRepo.CreateApprovalItems(approvalItems); err != nil {
			return nil, err
		}

		if err := s.reqRepo.Update(req); err != nil {
			return nil, err
		}
	}

	s.auditRepo.Create(&domain.AuditLog{
		UserID:     &approverID,
		UserName:   actorName,
		Action:     "PROCESS_APPROVAL_" + dto.Action,
		EntityType: "REQUEST",
		EntityID:   req.RequestNo,
		IPAddress:  ip,
	})

	return s.reqRepo.FindByID(req.ID)
}

func (s *RequestService) ProcessDelivery(requestID uint, agencyID uint, dto *domain.ProcessDeliveryDTO, actorName, ip string) (*domain.Request, error) {
	req, err := s.reqRepo.FindByID(requestID)
	if err != nil {
		return nil, errors.New("request not found")
	}

	if req.Status != "APPROVED" && req.Status != "PARTIALLY_DELIVERED" {
		return nil, errors.New("request is not pending delivery")
	}

	totalDeliveredMap := make(map[uint]int)
	totalUnavailableMap := make(map[uint]int)

	for _, d := range req.Deliveries {
		for _, di := range d.Items {
			totalDeliveredMap[di.ProductID] += di.DeliveredQty
			totalUnavailableMap[di.ProductID] += di.UnavailableQty
		}
	}

	var deliveryItems []domain.DeliveryItem
	for _, item := range dto.Items {
		totalDeliveredMap[item.ProductID] += item.DeliveredQty
		totalUnavailableMap[item.ProductID] += item.UnavailableQty

		deliveryItems = append(deliveryItems, domain.DeliveryItem{
			ProductID:      item.ProductID,
			ApprovedQty:    item.ApprovedQty,
			DeliveredQty:   item.DeliveredQty,
			UnavailableQty: item.UnavailableQty,
			UnitPrice:      item.UnitPrice,
			Remarks:        item.Remarks,
		})

		// Update RequestItem and Product unit prices in DB when agency enters price
		if item.UnitPrice > 0 {
			_ = s.reqRepo.UpdateItemUnitPrice(requestID, item.ProductID, item.UnitPrice)
			_ = s.reqRepo.UpdateProductUnitPrice(item.ProductID, item.UnitPrice)
		}
	}

	allFullyDelivered := true
	for _, reqItem := range req.Items {
		approvedQty := reqItem.RequestedQty
		if reqItem.ApprovalItem != nil {
			approvedQty = reqItem.ApprovalItem.ApprovedQty
		}
		if approvedQty <= 0 {
			continue
		}
		cumDelivered := totalDeliveredMap[reqItem.ProductID]
		cumUnavailable := totalUnavailableMap[reqItem.ProductID]
		if cumDelivered+cumUnavailable < approvedQty {
			allFullyDelivered = false
			break
		}
	}

	delivery := &domain.Delivery{
		RequestID:     requestID,
		AgencyUser:    agencyID,
		DeliveredDate: time.Now(),
		Status:        "DELIVERED",
		BillUrl:       dto.BillUrl,
		BillNotes:     dto.BillNotes,
		Items:         deliveryItems,
	}

	if err := s.reqRepo.CreateDelivery(delivery); err != nil {
		return nil, err
	}

	if allFullyDelivered {
		req.Status = "DELIVERED"
	} else {
		req.Status = "PARTIALLY_DELIVERED"
	}

	if err := s.reqRepo.Update(req); err != nil {
		return nil, err
	}

	s.auditRepo.Create(&domain.AuditLog{
		UserID:     &agencyID,
		UserName:   actorName,
		Action:     "PROCESS_DELIVERY",
		EntityType: "REQUEST",
		EntityID:   req.RequestNo,
		IPAddress:  ip,
	})

	return s.reqRepo.FindByID(req.ID)
}

func (s *RequestService) ProcessVerification(requestID uint, verifier *domain.User, dto *domain.ProcessVerificationDTO, actorName, ip string) (*domain.Request, error) {
	verifierID := verifier.ID
	req, err := s.reqRepo.FindByID(requestID)
	if err != nil {
		return nil, errors.New("request not found")
	}

	if verifier.Role.Name == "BRANCH_REQUESTER" && req.RequesterID != verifierID {
		return nil, errors.New("You can only verify requests you submitted.")
	}

	var verifications []domain.VerificationItem
	for _, item := range dto.Items {
		verifications = append(verifications, domain.VerificationItem{
			DeliveryItemID: item.DeliveryItemID,
			AcceptedQty:    item.AcceptedQty,
			DamagedQty:     item.DamagedQty,
			NotReceivedQty: item.NotReceivedQty,
			Remarks:        item.Remarks,
		})
	}

	if err := s.reqRepo.CreateVerifications(verifications); err != nil {
		return nil, err
	}

	now := time.Now()
	req.Status = "COMPLETED"
	req.CompletedAt = &now

	if err := s.reqRepo.Update(req); err != nil {
		return nil, err
	}

	s.auditRepo.Create(&domain.AuditLog{
		UserID:     &verifierID,
		UserName:   actorName,
		Action:     "VERIFY_DELIVERY",
		EntityType: "REQUEST",
		EntityID:   req.RequestNo,
		IPAddress:  ip,
	})

	return s.reqRepo.FindByID(req.ID)
}
