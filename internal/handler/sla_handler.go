package handler

import (
	"net/http"
	"stationery-management/internal/domain"
	"stationery-management/internal/service"
	"stationery-management/pkg/response"

	"github.com/gin-gonic/gin"
)

type SlaHandler struct {
	slaSvc service.SlaService
}

func NewSlaHandler(slaSvc service.SlaService) *SlaHandler {
	return &SlaHandler{slaSvc: slaSvc}
}

func (h *SlaHandler) GetSlaSettings(c *gin.Context) {
	settings, err := h.slaSvc.GetSlaSettings()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}
	response.JSONSuccess(c, http.StatusOK, "SLA settings retrieved successfully", settings)
}

func (h *SlaHandler) UpdateSlaSettings(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		response.Unauthorized(c, "Unauthorized access")
		return
	}
	userID := userIDVal.(uint)
	userEmail := c.GetString("userEmail")

	var dto domain.UpdateSlaSettingsDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		response.BadRequest(c, "Invalid SLA settings payload", err.Error())
		return
	}

	settings, err := h.slaSvc.UpdateSlaSettings(userID, userEmail, c.ClientIP(), dto)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.JSONSuccess(c, http.StatusOK, "SLA thresholds updated successfully", settings)
}

func (h *SlaHandler) GetDelayedOrders(c *gin.Context) {
	userRole := c.GetString("userRole")
	deptFilter := ""
	if userRole == "MONITOR" || userRole == "APPROVER" {
		deptFilter = c.GetString("department")
	}
	if deptFilter == "" && c.Query("department") != "" {
		deptFilter = c.Query("department")
	}

	delayedOrders, err := h.slaSvc.GetDelayedOrders(deptFilter)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.JSONSuccess(c, http.StatusOK, "Delayed orders retrieved successfully", delayedOrders)
}
