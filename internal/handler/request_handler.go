package handler

import (
	"strconv"
	"stationery-management/internal/domain"
	"stationery-management/internal/service"
	"stationery-management/pkg/response"

	"github.com/gin-gonic/gin"
)

type RequestHandler struct {
	reqSvc  *service.RequestService
	userSvc *service.UserService
}

func NewRequestHandler(reqSvc *service.RequestService, userSvc *service.UserService) *RequestHandler {
	return &RequestHandler{reqSvc: reqSvc, userSvc: userSvc}
}

func (h *RequestHandler) CreateRequest(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	userEmail := c.MustGet("userEmail").(string)

	var req domain.CreateRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", err.Error())
		return
	}

	result, err := h.reqSvc.CreateRequest(userID, &req, userEmail, c.ClientIP())
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSONSuccess(c, 201, "Stationery request submitted successfully", result)
}

func (h *RequestHandler) GetRequests(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	status := c.Query("status")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	user, err := h.userSvc.GetUserByID(userID)
	if err != nil {
		response.Unauthorized(c, "User session invalid")
		return
	}

	requests, total, err := h.reqSvc.GetRequests(user, status, page, limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.JSONSuccess(c, 200, "Requests retrieved", gin.H{
		"requests": requests,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

func (h *RequestHandler) GetRequestByID(c *gin.Context) {
	idParam, _ := strconv.Atoi(c.Param("id"))
	req, err := h.reqSvc.GetRequestByID(uint(idParam))
	if err != nil {
		response.NotFound(c, "Request not found")
		return
	}

	response.JSONSuccess(c, 200, "Request details retrieved", req)
}

func (h *RequestHandler) ProcessApproval(c *gin.Context) {
	idParam, _ := strconv.Atoi(c.Param("id"))
	userID := c.MustGet("userID").(uint)
	userEmail := c.MustGet("userEmail").(string)

	var dto domain.ProcessApprovalDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		response.BadRequest(c, "Validation failed", err.Error())
		return
	}

	req, err := h.reqSvc.ProcessApproval(uint(idParam), userID, &dto, userEmail, c.ClientIP())
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSONSuccess(c, 200, "Request approval processed", req)
}

func (h *RequestHandler) ProcessDelivery(c *gin.Context) {
	idParam, _ := strconv.Atoi(c.Param("id"))
	userID := c.MustGet("userID").(uint)
	userEmail := c.MustGet("userEmail").(string)

	var dto domain.ProcessDeliveryDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		response.BadRequest(c, "Validation failed", err.Error())
		return
	}

	req, err := h.reqSvc.ProcessDelivery(uint(idParam), userID, &dto, userEmail, c.ClientIP())
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSONSuccess(c, 200, "Delivery recorded successfully", req)
}

func (h *RequestHandler) ProcessVerification(c *gin.Context) {
	idParam, _ := strconv.Atoi(c.Param("id"))
	userID := c.MustGet("userID").(uint)
	userEmail := c.MustGet("userEmail").(string)

	var dto domain.ProcessVerificationDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		response.BadRequest(c, "Validation failed", err.Error())
		return
	}

	req, err := h.reqSvc.ProcessVerification(uint(idParam), userID, &dto, userEmail, c.ClientIP())
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSONSuccess(c, 200, "Delivery verification completed", req)
}
