package handler

import (
	"stationery-management/internal/repository"
	"stationery-management/pkg/response"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	reqRepo   *repository.RequestRepository
	auditRepo *repository.AuditRepository
}

func NewDashboardHandler(reqRepo *repository.RequestRepository, auditRepo *repository.AuditRepository) *DashboardHandler {
	return &DashboardHandler{reqRepo: reqRepo, auditRepo: auditRepo}
}

func (h *DashboardHandler) GetMetrics(c *gin.Context) {
	metrics, err := h.reqRepo.GetDashboardMetrics()
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.JSONSuccess(c, 200, "Dashboard metrics retrieved", metrics)
}

func (h *DashboardHandler) GetProductTotals(c *gin.Context) {
	status := c.Query("status")
	totals, err := h.reqRepo.GetProductTotals(status)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.JSONSuccess(c, 200, "Product totals retrieved", totals)
}

func (h *DashboardHandler) GetAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	entityType := c.Query("entityType")
	action := c.Query("action")
	startDate := c.Query("startDate")
	endDate := c.Query("endDate")

	logs, total, err := h.auditRepo.FindAll(page, limit, entityType, action, startDate, endDate)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.JSONSuccess(c, 200, "Audit logs retrieved", gin.H{
		"logs":  logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}
