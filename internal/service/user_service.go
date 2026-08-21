package service

import (
	"errors"
	"fmt"
	"strings"
	"stationery-management/internal/domain"
	"stationery-management/internal/repository"
	"stationery-management/pkg/hash"
)

// friendlyUserDBError converts raw MySQL duplicate key errors into user-friendly messages.
func friendlyUserDBError(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	if strings.Contains(msg, "1062") {
		if strings.Contains(msg, "uni_users_mobile") || strings.Contains(msg, "mobile") {
			return errors.New("This mobile number is already associated with another account. Please use a different number.")
		}
		if strings.Contains(msg, "uni_users_email") || strings.Contains(msg, "email") {
			return errors.New("This email address is already associated with another account. Please use a different email.")
		}
		if strings.Contains(msg, "uni_users_name") || strings.Contains(msg, "name") {
			return errors.New("A user with this name already exists. Please choose a different name.")
		}
		return errors.New("A duplicate entry was detected. Please check your inputs and try again.")
	}
	return err
}

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
	// Role-specific constraints & validations
	switch req.RoleID {
	case 2: // BRANCH_REQUESTER (Max 1 Requester per department)
		if req.Department == "" {
			return nil, errors.New("Department selection is required for Requester role.")
		}
		count, err := s.userRepo.CountByRoleAndDepartment(2, req.Department, 0)
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, errors.New(fmt.Sprintf("A Requester account already exists for the %s department. Only 1 Requester per department is allowed.", req.Department))
		}
		req.BranchID = nil // Requester accounts are not branch-scoped
	case 4: // AGENCY (Global delivery agency)
		count, err := s.userRepo.CountByRoleID(4, 0)
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, errors.New("An Agency account already exists in the system. Only 1 Agency account is allowed.")
		}
		req.BranchID = nil
	case 3: // APPROVER (Max 1 Approver per department)
		if req.Department == "" {
			return nil, errors.New("Department selection is required for Approver role.")
		}
		count, err := s.userRepo.CountByRoleAndDepartment(3, req.Department, 0)
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, errors.New(fmt.Sprintf("An Approver account already exists for the %s department. Only 1 Approver per department is allowed.", req.Department))
		}
	case 5: // MONITOR (Max 1 Monitor per department)
		if req.Department == "" {
			return nil, errors.New("Department selection is required for Monitor role.")
		}
		count, err := s.userRepo.CountByRoleAndDepartment(5, req.Department, 0)
		if err != nil {
			return nil, err
		}
		if count > 0 {
			return nil, errors.New(fmt.Sprintf("A Monitor account already exists for the %s department. Only 1 Monitor per department is allowed.", req.Department))
		}
		req.BranchID = nil
	}

	exists, err := s.userRepo.ExistsByName(req.Name, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("A user with this name already exists. Please choose a different name.")
	}

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
		return nil, friendlyUserDBError(err)
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

	targetRoleID := user.RoleID
	if req.RoleID != 0 {
		targetRoleID = req.RoleID
	}
	targetDept := user.Department
	if req.Department != "" {
		targetDept = req.Department
	}

	switch targetRoleID {
	case 2:
		if targetDept != "" {
			count, err := s.userRepo.CountByRoleAndDepartment(2, targetDept, id)
			if err == nil && count > 0 {
				return nil, errors.New(fmt.Sprintf("A Requester account already exists for the %s department.", targetDept))
			}
		}
	case 4:
		count, err := s.userRepo.CountByRoleID(4, id)
		if err == nil && count > 0 {
			return nil, errors.New("An Agency account already exists in the system.")
		}
	case 3:
		if targetDept != "" {
			count, err := s.userRepo.CountByRoleAndDepartment(3, targetDept, id)
			if err == nil && count > 0 {
				return nil, errors.New(fmt.Sprintf("An Approver account already exists for the %s department.", targetDept))
			}
		}
	case 5:
		if targetDept != "" {
			count, err := s.userRepo.CountByRoleAndDepartment(5, targetDept, id)
			if err == nil && count > 0 {
				return nil, errors.New(fmt.Sprintf("A Monitor account already exists for the %s department.", targetDept))
			}
		}
	}

	if req.Name != "" {
		exists, err := s.userRepo.ExistsByName(req.Name, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, errors.New("A user with this name already exists. Please choose a different name.")
		}
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
	if req.Password != "" {
		newHash, err := hash.HashPassword(req.Password)
		if err != nil {
			return nil, err
		}
		user.Password = newHash
	}

	if err := s.userRepo.Update(user); err != nil {
		return nil, friendlyUserDBError(err)
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

func (s *UserService) DeleteUser(userID uint, actorID uint, actorName, ip string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	if user.Role.Name == "ADMIN" || user.ID == actorID {
		return errors.New("cannot delete the primary admin user account")
	}

	if err := s.userRepo.Delete(userID); err != nil {
		return err
	}

	s.auditRepo.Create(&domain.AuditLog{
		UserID:     &actorID,
		UserName:   actorName,
		Action:     "DELETE_USER",
		EntityType: "USER",
		EntityID:   user.Email,
		IPAddress:  ip,
	})

	return nil
}

