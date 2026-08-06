package handler

import (
	"strconv"
	"stationery-management/internal/repository"
	"stationery-management/pkg/response"

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

func (h *DashboardHandler) GetAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	logs, total, err := h.auditRepo.FindAll(page, limit)
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
