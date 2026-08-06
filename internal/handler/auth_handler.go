package handler

import (
	"stationery-management/internal/domain"
	"stationery-management/internal/service"
	"stationery-management/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req domain.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload", err.Error())
		return
	}

	resp, err := h.authSvc.Login(req.Mobile, req.Password, c.ClientIP())
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSONSuccess(c, 200, "Login successful", resp)
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	var req domain.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request payload", err.Error())
		return
	}

	if err := h.authSvc.ChangePassword(userID, req.OldPassword, req.NewPassword, c.ClientIP()); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSONSuccess(c, 200, "Password changed successfully", nil)
}

func (h *AuthHandler) GetMe(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	response.JSONSuccess(c, 200, "User profile retrieved", gin.H{
		"userID":             userID,
		"email":              c.MustGet("userEmail"),
		"role":               c.MustGet("userRole"),
		"branchId":           c.MustGet("userBranchID"),
		"approverAccessType": c.MustGet("approverAccessType"),
		"firstLogin":         c.MustGet("firstLogin"),
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	response.JSONSuccess(c, 200, "Logout successful", nil)
}
