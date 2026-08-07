package repository

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"stationery-management/internal/domain"

	"gorm.io/gorm"
)

type RequestRepository struct {
	db *gorm.DB
}

func NewRequestRepository(db *gorm.DB) *RequestRepository {
	return &RequestRepository{db: db}
}

func (r *RequestRepository) FindOrCreateBranchByName(name string) (*domain.Branch, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		var defaultBranch domain.Branch
		if err := r.db.First(&defaultBranch).Error; err == nil {
			return &defaultBranch, nil
		}
		return nil, errors.New("branch name required")
	}

	var branch domain.Branch
	if err := r.db.Where("LOWER(name) = ?", strings.ToLower(name)).First(&branch).Error; err == nil {
		return &branch, nil
	}

	code := fmt.Sprintf("BR-%d", time.Now().UnixNano()%10000)
	newBranch := domain.Branch{
		Name:   name,
		Code:   code,
		Status: "ACTIVE",
	}
	if err := r.db.Create(&newBranch).Error; err != nil {
		var defaultBranch domain.Branch
		if err := r.db.First(&defaultBranch).Error; err == nil {
			return &defaultBranch, nil
		}
		return nil, err
	}
	return &newBranch, nil
}

func (r *RequestRepository) GenerateUniqueRequestNo() string {
	dateStr := time.Now().Format("20060102")
	for i := 0; i < 20; i++ {
		randVal := (time.Now().UnixNano()/1000 + int64(i*37))%9000 + 1000
		candidate := fmt.Sprintf("REQ-%s-%d", dateStr, randVal)

		var count int64
		r.db.Model(&domain.Request{}).Where("request_no = ?", candidate).Count(&count)
		if count == 0 {
			return candidate
		}
	}
	return fmt.Sprintf("REQ-%s-%d", time.Now().Format("20060102150405"), (time.Now().UnixNano()/1000)%1000)
}

func (r *RequestRepository) Create(req *domain.Request) error {
	return r.db.Create(req).Error
}

func (r *RequestRepository) FindByID(id uint) (*domain.Request, error) {
	var req domain.Request
	err := r.db.Preload("Branch").
		Preload("Requester").
		Preload("Items.Product").
		Preload("Items.ApprovalItem.Approver").
		Preload("Deliveries.DeliveryAgent").
		Preload("Deliveries.Items.Product").
		Preload("Deliveries.Items.VerificationItem").
		Where("id = ?", id).
		First(&req).Error
	if err != nil {
		return nil, err
	}

	var chatCount int64
	r.db.Model(&domain.ChatMessage{}).Where("request_id = ?", req.ID).Count(&chatCount)
	req.ChatCount = int(chatCount)

	return &req, nil
}

func (r *RequestRepository) FindAll(branchID *uint, requesterID *uint, department string, status string, page, limit int) ([]domain.Request, int64, error) {
	var requests []domain.Request
	var total int64

	query := r.db.Model(&domain.Request{}).
		Preload("Branch").
		Preload("Requester").
		Preload("Items.Product").
		Preload("Items.ApprovalItem.Approver").
		Preload("Deliveries.DeliveryAgent").
		Preload("Deliveries.Items.Product").
		Preload("Deliveries.Items.VerificationItem")

	if branchID != nil {
		query = query.Where("branch_id = ?", *branchID)
	}
	if requesterID != nil {
		query = query.Where("requester_id = ?", *requesterID)
	}
	if department != "" {
		query = query.Where("department = ?", department)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query.Count(&total)
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("id desc").Find(&requests).Error
	if err != nil {
		return nil, 0, err
	}

	if len(requests) > 0 {
		var reqIDs []uint
		for _, req := range requests {
			reqIDs = append(reqIDs, req.ID)
		}

		type ChatCountResult struct {
			RequestID uint `gorm:"column:request_id"`
			Count     int  `gorm:"column:count"`
		}
		var counts []ChatCountResult
		r.db.Model(&domain.ChatMessage{}).
			Select("request_id, count(*) as count").
			Where("request_id IN ?", reqIDs).
			Group("request_id").
			Scan(&counts)

		countMap := make(map[uint]int)
		for _, c := range counts {
			countMap[c.RequestID] = c.Count
		}

		for i := range requests {
			requests[i].ChatCount = countMap[requests[i].ID]
		}
	}

	return requests, total, nil
}

func (r *RequestRepository) Update(req *domain.Request) error {
	return r.db.Save(req).Error
}

func (r *RequestRepository) CreateApprovalItems(items []domain.ApprovalItem) error {
	return r.db.Create(&items).Error
}

func (r *RequestRepository) CreateDelivery(delivery *domain.Delivery) error {
	return r.db.Create(delivery).Error
}

func (r *RequestRepository) FindDeliveriesByRequestID(requestID uint) ([]domain.Delivery, error) {
	var deliveries []domain.Delivery
	err := r.db.Preload("DeliveryAgent").
		Preload("Items.Product").
		Preload("Items.VerificationItem").
		Where("request_id = ?", requestID).
		Find(&deliveries).Error
	return deliveries, err
}

func (r *RequestRepository) CreateVerifications(items []domain.VerificationItem) error {
	return r.db.Create(&items).Error
}

func (r *RequestRepository) GetDashboardMetrics() (*domain.DashboardMetrics, error) {
	var metrics domain.DashboardMetrics

	r.db.Model(&domain.Product{}).Count(&metrics.TotalProducts)
	r.db.Model(&domain.Request{}).Count(&metrics.TotalRequests)
	r.db.Model(&domain.Request{}).Where("status = ?", "SUBMITTED").Count(&metrics.PendingApprovals)
	r.db.Model(&domain.Request{}).Where("status IN (?, ?)", "APPROVED", "PARTIALLY_DELIVERED").Count(&metrics.PendingDeliveries)
	r.db.Model(&domain.Request{}).Where("status = ?", "COMPLETED").Count(&metrics.Completed)
	r.db.Model(&domain.Request{}).Where("status = ?", "REJECTED").Count(&metrics.Rejected)

	// Delayed requests (> 3 days pending approval or delivery)
	threeDaysAgo := time.Now().AddDate(0, 0, -3)
	r.db.Model(&domain.Request{}).Where("status IN (?, ?) AND created_at < ?", "SUBMITTED", "APPROVED", threeDaysAgo).Count(&metrics.Delayed)

	var damagedSum struct{ Sum int64 }
	r.db.Model(&domain.VerificationItem{}).Select("COALESCE(SUM(damaged_qty), 0) as sum").Scan(&damagedSum)
	metrics.DamagedItems = damagedSum.Sum

	var unavailSum struct{ Sum int64 }
	r.db.Model(&domain.DeliveryItem{}).Select("COALESCE(SUM(unavailable_qty), 0) as sum").Scan(&unavailSum)
	metrics.UnavailableItems = unavailSum.Sum

	return &metrics, nil
}

func (r *RequestRepository) UpdateItemUnitPrice(requestID, productID uint, unitPrice float64) error {
	return r.db.Model(&domain.RequestItem{}).
		Where("request_id = ? AND product_id = ?", requestID, productID).
		Update("unit_price", unitPrice).Error
}

func (r *RequestRepository) UpdateProductUnitPrice(productID uint, unitPrice float64) error {
	return r.db.Model(&domain.Product{}).
		Where("id = ?", productID).
		Update("unit_price", unitPrice).Error
}
