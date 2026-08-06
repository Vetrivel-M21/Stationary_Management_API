package handler

import (
	"strconv"
	"stationery-management/internal/domain"
	"stationery-management/internal/service"
	"stationery-management/pkg/response"

	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	productSvc *service.ProductService
}

func NewProductHandler(productSvc *service.ProductService) *ProductHandler {
	return &ProductHandler{productSvc: productSvc}
}

func (h *ProductHandler) GetAllProducts(c *gin.Context) {
	search := c.Query("search")
	category := c.Query("category")
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))

	products, total, err := h.productSvc.GetAllProducts(search, category, page, limit)
	if err != nil {
		response.InternalError(c, err.Error())
		return
	}

	response.JSONSuccess(c, 200, "Products retrieved", gin.H{
		"products": products,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

func (h *ProductHandler) CreateProduct(c *gin.Context) {
	actorID := c.MustGet("userID").(uint)
	actorEmail := c.MustGet("userEmail").(string)

	var req domain.CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", err.Error())
		return
	}

	product, err := h.productSvc.CreateProduct(&req, actorID, actorEmail, c.ClientIP())
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSONSuccess(c, 201, "Product created successfully", product)
}

func (h *ProductHandler) UpdateProduct(c *gin.Context) {
	idParam, _ := strconv.Atoi(c.Param("id"))
	actorID := c.MustGet("userID").(uint)
	actorEmail := c.MustGet("userEmail").(string)

	var req domain.UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Validation failed", err.Error())
		return
	}

	product, err := h.productSvc.UpdateProduct(uint(idParam), &req, actorID, actorEmail, c.ClientIP())
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSONSuccess(c, 200, "Product updated successfully", product)
}

func (h *ProductHandler) DeleteProduct(c *gin.Context) {
	idParam, _ := strconv.Atoi(c.Param("id"))
	actorID := c.MustGet("userID").(uint)
	actorEmail := c.MustGet("userEmail").(string)

	if err := h.productSvc.DeleteProduct(uint(idParam), actorID, actorEmail, c.ClientIP()); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	response.JSONSuccess(c, 200, "Product disabled/deleted successfully", nil)
}
