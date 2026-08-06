package handler

import (
	"strconv"
	"stationery-management/internal/domain"
	"stationery-management/internal/service"
	"stationery-management/pkg/response"

	"github.com/gin-gonic/gin"
)

type BranchHandler struct {
	branchSvc *service.BranchService
}

func NewBranchHandler(branchSvc *service.BranchService) *BranchHandler {
	return &BranchHandler{branchSvc: branchSvc}
}

func (h *BranchHandler) GetAllBranches(c *gin.Context) {
	search := c.Query("search")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	branches, total, err := h.branchSvc.GetAllBranches(search, page, limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.JSONSuccess(c, 200, "Branches retrieved", gin.H{
		"branches": branches,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

func (h *BranchHandler) CreateBranch(c *gin.Context) {
	actorID := c.MustGet("userID").(uint)
	actorEmail := c.MustGet("userEmail").(string)

	var req domain.CreateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", err.Error())
		return
	}

	branch, err := h.branchSvc.CreateBranch(&req, actorID, actorEmail, c.ClientIP())
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSONSuccess(c, 201, "Branch created successfully", branch)
}

func (h *BranchHandler) UpdateBranch(c *gin.Context) {
	idParam, _ := strconv.Atoi(c.Param("id"))
	actorID := c.MustGet("userID").(uint)
	actorEmail := c.MustGet("userEmail").(string)

	var req domain.UpdateBranchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", err.Error())
		return
	}

	branch, err := h.branchSvc.UpdateBranch(uint(idParam), &req, actorID, actorEmail, c.ClientIP())
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSONSuccess(c, 200, "Branch updated successfully", branch)
}
