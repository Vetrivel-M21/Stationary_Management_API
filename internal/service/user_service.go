package service

import (
	"errors"
	"stationery-management/internal/domain"
	"stationery-management/internal/repository"
	"stationery-management/pkg/hash"
)

type UserService struct {
	userRepo  *repository.UserRepository
	auditRepo *repository.AuditRepository
}

func NewUserService(userRepo *repository.UserRepository, auditRepo *repository.AuditRepository) *UserService {
	return &UserService{userRepo: userRepo, auditRepo: auditRepo}
}

func (s *UserService) GetAllUsers(search string, page, limit int) ([]domain.User, int64, error) {
	return s.userRepo.FindAll(search, page, limit)
}

func (s *UserService) GetUserByID(id uint) (*domain.User, error) {
	return s.userRepo.FindByID(id)
}

func (s *UserService) CreateUser(req *domain.CreateUserRequest, actorID uint, actorName, ip string) (*domain.User, error) {
	hashedPassword, err := hash.HashPassword(req.DefaultPassword)
	if err != nil {
		return nil, err
	}

	approverAccess := req.ApproverAccessType
	if approverAccess == "" {
		approverAccess = "ALL_BRANCHES"
	}

	user := &domain.User{
		Name:               req.Name,
		Email:              req.Email,
		Mobile:             req.Mobile,
		Password:           hashedPassword,
		RoleID:             req.RoleID,
		BranchID:           req.BranchID,
		Department:         req.Department,
		ApproverAccessType: approverAccess,
		Status:             "ACTIVE",
		FirstLogin:         false,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	s.auditRepo.Create(&domain.AuditLog{
		UserID:     &actorID,
		UserName:   actorName,
		Action:     "CREATE_USER",
		EntityType: "USER",
		EntityID:   user.Email,
		IPAddress:  ip,
	})

	return s.userRepo.FindByID(user.ID)
}

func (s *UserService) UpdateUser(id uint, req *domain.UpdateUserRequest, actorID uint, actorName, ip string) (*domain.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("user not found")
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}
	if req.Mobile != "" {
		user.Mobile = req.Mobile
	}
	if req.RoleID != 0 {
		user.RoleID = req.RoleID
	}
	if req.BranchID != nil {
		user.BranchID = req.BranchID
	}
	if req.Department != "" {
		user.Department = req.Department
	}
	if req.ApproverAccessType != "" {
		user.ApproverAccessType = req.ApproverAccessType
	}
	if req.Status != "" {
		user.Status = req.Status
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, err
	}

	s.auditRepo.Create(&domain.AuditLog{
		UserID:     &actorID,
		UserName:   actorName,
		Action:     "UPDATE_USER",
		EntityType: "USER",
		EntityID:   user.Email,
		IPAddress:  ip,
	})

	return s.userRepo.FindByID(user.ID)
}

func (s *UserService) ResetPassword(userID uint, newPassword string, actorID uint, actorName, ip string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	newHash, err := hash.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.Password = newHash
	user.FirstLogin = true

	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	s.auditRepo.Create(&domain.AuditLog{
		UserID:     &actorID,
		UserName:   actorName,
		Action:     "RESET_USER_PASSWORD",
		EntityType: "USER",
		EntityID:   user.Email,
		IPAddress:  ip,
	})

	return nil
}
