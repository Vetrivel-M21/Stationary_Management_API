package service

import (
	"errors"
	"stationery-management/internal/config"
	"stationery-management/internal/domain"
	"stationery-management/internal/repository"
	"stationery-management/pkg/hash"
	"stationery-management/pkg/jwt"
)

type AuthService struct {
	userRepo  *repository.UserRepository
	auditRepo *repository.AuditRepository
	cfg       *config.Config
}

func NewAuthService(userRepo *repository.UserRepository, auditRepo *repository.AuditRepository, cfg *config.Config) *AuthService {
	return &AuthService{userRepo: userRepo, auditRepo: auditRepo, cfg: cfg}
}

func (s *AuthService) Login(mobile, password, ip string) (*domain.LoginResponse, error) {
	user, err := s.userRepo.FindByMobile(mobile)
	if err != nil {
		return nil, errors.New("invalid mobile number or password")
	}

	if user.Status != "ACTIVE" {
		return nil, errors.New("account is disabled")
	}

	if !hash.CheckPasswordHash(password, user.Password) {
		return nil, errors.New("invalid mobile number or password")
	}

	token, err := jwt.GenerateToken(
		user.ID, user.Email, user.Role.Name, user.BranchID, user.ApproverAccessType, user.FirstLogin,
		s.cfg.JWTSecret, s.cfg.JWTExpirationHours,
	)
	if err != nil {
		return nil, err
	}

	refreshToken, err := jwt.GenerateToken(
		user.ID, user.Email, user.Role.Name, user.BranchID, user.ApproverAccessType, user.FirstLogin,
		s.cfg.JWTSecret, s.cfg.JWTRefreshExpirationHour,
	)
	if err != nil {
		return nil, err
	}

	// Audit Log
	s.auditRepo.Create(&domain.AuditLog{
		UserID:     &user.ID,
		UserName:   user.Name,
		Action:     "USER_LOGIN",
		EntityType: "USER",
		EntityID:   user.Email,
		IPAddress:  ip,
	})

	return &domain.LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User:         *user,
	}, nil
}

func (s *AuthService) ChangePassword(userID uint, oldPassword, newPassword, ip string) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return errors.New("user not found")
	}

	if !hash.CheckPasswordHash(oldPassword, user.Password) {
		return errors.New("current password is incorrect")
	}

	newHash, err := hash.HashPassword(newPassword)
	if err != nil {
		return err
	}

	user.Password = newHash
	user.FirstLogin = false

	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	s.auditRepo.Create(&domain.AuditLog{
		UserID:     &user.ID,
		UserName:   user.Name,
		Action:     "CHANGE_PASSWORD",
		EntityType: "USER",
		EntityID:   user.Email,
		IPAddress:  ip,
	})

	return nil
}
