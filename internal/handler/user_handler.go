package handler

import (
	"strconv"
	"stationery-management/internal/domain"
	"stationery-management/internal/service"
	"stationery-management/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userSvc *service.UserService
}

func NewUserHandler(userSvc *service.UserService) *UserHandler {
	return &UserHandler{userSvc: userSvc}
}

func (h *UserHandler) GetAllUsers(c *gin.Context) {
	search := c.Query("search")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	users, total, err := h.userSvc.GetAllUsers(search, page, limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.JSONSuccess(c, 200, "Users retrieved", gin.H{
		"users": users,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *UserHandler) CreateUser(c *gin.Context) {
	actorID := c.MustGet("userID").(uint)
	actorEmail := c.MustGet("userEmail").(string)

	var req domain.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", err.Error())
		return
	}

	user, err := h.userSvc.CreateUser(&req, actorID, actorEmail, c.ClientIP())
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSONSuccess(c, 201, "User created successfully", user)
}

func (h *UserHandler) UpdateUser(c *gin.Context) {
	idParam, _ := strconv.Atoi(c.Param("id"))
	actorID := c.MustGet("userID").(uint)
	actorEmail := c.MustGet("userEmail").(string)

	var req domain.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", err.Error())
		return
	}

	user, err := h.userSvc.UpdateUser(uint(idParam), &req, actorID, actorEmail, c.ClientIP())
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSONSuccess(c, 200, "User updated successfully", user)
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	actorID := c.MustGet("userID").(uint)
	actorEmail := c.MustGet("userEmail").(string)

	var req domain.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", err.Error())
		return
	}

	if err := h.userSvc.ResetPassword(req.UserID, req.NewPassword, actorID, actorEmail, c.ClientIP()); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSONSuccess(c, 200, "User password reset successfully", nil)
}
