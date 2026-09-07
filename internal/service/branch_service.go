package service

import (
	"errors"
	"fmt"
	"stationery-management/internal/domain"
	"stationery-management/internal/repository"
	"strconv"
	"time"
)

type BranchService struct {
	branchRepo *repository.BranchRepository
	auditRepo  *repository.AuditRepository
}

func NewBranchService(branchRepo *repository.BranchRepository, auditRepo *repository.AuditRepository) *BranchService {
	return &BranchService{branchRepo: branchRepo, auditRepo: auditRepo}
}

func (s *BranchService) GetAllBranches(search, department string, page, limit int) ([]domain.Branch, int64, error) {
	return s.branchRepo.FindAll(search, department, page, limit)
}

func (s *BranchService) CreateBranch(req *domain.CreateBranchRequest, actorID uint, actorName, ip string) (*domain.Branch, error) {
	branch := &domain.Branch{
		Name:       req.Name,
		Code:       "TMP" + strconv.FormatInt(time.Now().UnixNano(), 36),
		Address:    req.Address,
		Department: req.Department,
		Status:     "ACTIVE",
	}

	if err := s.branchRepo.Create(branch); err != nil {
		return nil, err
	}

	// Auto-increment branch code derived from the DB-assigned primary key,
	// guaranteed unique without requiring manual entry from the admin.
	branch.Code = fmt.Sprintf("BR-%03d", branch.ID)
	if err := s.branchRepo.Update(branch); err != nil {
		return nil, err
	}

	s.auditRepo.Create(&domain.AuditLog{
		UserID:     &actorID,
		UserName:   actorName,
		Action:     "CREATE_BRANCH",
		EntityType: "BRANCH",
		EntityID:   branch.Code,
		IPAddress:  ip,
	})

	return branch, nil
}

func (s *BranchService) UpdateBranch(id uint, req *domain.UpdateBranchRequest, actorID uint, actorName, ip string) (*domain.Branch, error) {
	branch, err := s.branchRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("branch not found")
	}

	if req.Name != "" {
		branch.Name = req.Name
	}
	if req.Code != "" {
		branch.Code = req.Code
	}
	if req.Address != "" {
		branch.Address = req.Address
	}
	if req.Department != "" {
		branch.Department = req.Department
	}
	if req.Status != "" {
		branch.Status = req.Status
	}

	if err := s.branchRepo.Update(branch); err != nil {
		return nil, err
	}

	s.auditRepo.Create(&domain.AuditLog{
		UserID:     &actorID,
		UserName:   actorName,
		Action:     "UPDATE_BRANCH",
		EntityType: "BRANCH",
		EntityID:   branch.Code,
		IPAddress:  ip,
	})

	return branch, nil
}
