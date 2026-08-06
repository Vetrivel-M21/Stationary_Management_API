package repository

import (
	"stationery-management/internal/domain"
	"time"

	"gorm.io/gorm"
)

type RequestRepository struct {
	db *gorm.DB
}

func NewRequestRepository(db *gorm.DB) *RequestRepository {
	return &RequestRepository{db: db}
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
		Preload("Deliveries.Items.Product")

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
