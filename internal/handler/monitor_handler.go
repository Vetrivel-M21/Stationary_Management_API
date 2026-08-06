package handler

import (
	"stationery-management/internal/domain"
	"stationery-management/internal/service"
	"stationery-management/pkg/response"

	"github.com/gin-gonic/gin"
)

type MonitorHandler struct {
	monitorSvc *service.MonitorService
}

func NewMonitorHandler(monitorSvc *service.MonitorService) *MonitorHandler {
	return &MonitorHandler{monitorSvc: monitorSvc}
}

func (h *MonitorHandler) SendReminder(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	userEmail := c.MustGet("userEmail").(string)

	var dto domain.SendReminderDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		response.BadRequest(c, "Validation failed", err.Error())
		return
	}

	if err := h.monitorSvc.SendReminder(&dto, userID, userEmail, c.ClientIP()); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSONSuccess(c, 200, "Reminder email dispatched successfully", nil)
}
