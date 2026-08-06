package handler

import (
	"net/http"
	"stationery-management/internal/domain"
	"stationery-management/internal/service"

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
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    settings,
	})
}

func (h *SlaHandler) UpdateSlaSettings(c *gin.Context) {
	currentUserVal, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUserVal.(*domain.User)

	var dto domain.UpdateSlaSettingsDTO
	if err := c.ShouldBindJSON(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	settings, err := h.slaSvc.UpdateSlaSettings(user, dto)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    settings,
	})
}

func (h *SlaHandler) GetDelayedOrders(c *gin.Context) {
	currentUserVal, exists := c.Get("currentUser")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	user := currentUserVal.(*domain.User)

	deptFilter := ""
	if user.Role.Name == "MONITOR" || user.Role.Name == "APPROVER" {
		deptFilter = user.Department
	} else if c.Query("department") != "" {
		deptFilter = c.Query("department")
	}

	delayedOrders, err := h.slaSvc.GetDelayedOrders(deptFilter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    delayedOrders,
	})
}
