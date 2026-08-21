package domain

import (
	"time"

	"gorm.io/gorm"
)

type Role struct {
	ID          uint   `gorm:"primaryKey" json:"id"`
	Name        string `gorm:"size:50;unique;not null" json:"name"`
	Description string `gorm:"size:255" json:"description"`
}

type Branch struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `gorm:"size:100;not null" json:"name"`
	Code      string         `gorm:"size:20;unique;not null" json:"code"`
	Address   string         `gorm:"type:text" json:"address"`
	Status    string         `gorm:"size:20;default:'ACTIVE'" json:"status"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
}

type User struct {
	ID                 uint           `gorm:"primaryKey" json:"id"`
	Name               string         `gorm:"size:100;unique;not null" json:"name"`
	Email              string         `gorm:"size:100;unique;not null" json:"email"`
	Mobile             string         `gorm:"size:20;unique;not null" json:"mobile"`
	Password           string         `gorm:"size:255;not null" json:"-"`
	RoleID             uint           `gorm:"not null" json:"roleId"`
	Role               Role           `gorm:"foreignKey:RoleID" json:"role"`
	BranchID           *uint          `json:"branchId"`
	Branch             *Branch        `gorm:"foreignKey:BranchID" json:"branch,omitempty"`
	Department         string         `gorm:"size:50" json:"department"`
	ApproverAccessType string         `gorm:"size:20;default:'ALL_BRANCHES'" json:"approverAccessType"`
	Status             string         `gorm:"size:20;default:'ACTIVE'" json:"status"`
	FirstLogin         bool           `gorm:"default:false" json:"firstLogin"`
	DeletedAt          gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedAt          time.Time      `json:"createdAt"`
	UpdatedAt          time.Time      `json:"updatedAt"`
}

type Product struct {
	ID          uint           `gorm:"primaryKey" json:"id"`
	Name        string         `gorm:"size:150;not null" json:"name"`
	Category    string         `gorm:"size:50;not null" json:"category"`
	Unit        string         `gorm:"size:30;not null" json:"unit"`
	UnitPrice   float64        `gorm:"type:decimal(10,2);default:0" json:"unitPrice"`
	Description string         `gorm:"type:text" json:"description"`
	Status      string         `gorm:"size:20;default:'ACTIVE'" json:"status"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
}

type Request struct {
	ID              uint          `gorm:"primaryKey" json:"id"`
	RequestNo       string        `gorm:"size:50;unique;not null" json:"requestNo"`
	BranchID        uint          `gorm:"not null" json:"branchId"`
	Branch          Branch        `gorm:"foreignKey:BranchID" json:"branch"`
	RequesterID     uint          `gorm:"not null" json:"requesterId"`
	Requester       User          `gorm:"foreignKey:RequesterID" json:"requester"`
	ApplicantName   string        `gorm:"size:100" json:"applicantName"`
	ApplicantMobile string        `gorm:"size:20" json:"applicantMobile"`
	ApplicantEmail  string        `gorm:"size:100" json:"applicantEmail"`
	Department      string        `gorm:"size:50;not null" json:"department"`
	Location        string        `gorm:"size:255" json:"location"`
	Status          string        `gorm:"size:30;default:'SUBMITTED'" json:"status"`
	ChatCount       int           `gorm:"-" json:"chatCount"`
	Items           []RequestItem `gorm:"foreignKey:RequestID" json:"items"`
	Deliveries      []Delivery    `gorm:"foreignKey:RequestID" json:"deliveries,omitempty"`
	SubmittedAt     time.Time     `json:"submittedAt"`
	ApprovedAt      *time.Time    `json:"approvedAt,omitempty"`
	CompletedAt     *time.Time    `json:"completedAt,omitempty"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
}

type RequestItem struct {
	ID           uint          `gorm:"primaryKey" json:"id"`
	RequestID    uint          `gorm:"not null" json:"requestId"`
	ProductID    uint          `gorm:"not null" json:"productId"`
	Product      Product       `gorm:"foreignKey:ProductID" json:"product"`
	RequestedQty int           `gorm:"not null;default:1" json:"requestedQty"`
	UnitPrice    float64       `gorm:"type:decimal(10,2);default:0" json:"unitPrice"`
	ApprovalItem *ApprovalItem `gorm:"foreignKey:RequestItemID" json:"approvalItem,omitempty"`
}

type ApprovalItem struct {
	ID            uint   `gorm:"primaryKey" json:"id"`
	RequestItemID uint   `gorm:"not null" json:"requestItemId"`
	ApprovedQty   int    `gorm:"not null;default:0" json:"approvedQty"`
	ApprovedBy    uint   `gorm:"not null" json:"approvedBy"`
	Approver      User   `gorm:"foreignKey:ApprovedBy" json:"approver"`
	Remarks       string `gorm:"type:text" json:"remarks"`
}

type Delivery struct {
	ID            uint           `gorm:"primaryKey" json:"id"`
	RequestID     uint           `gorm:"not null" json:"requestId"`
	Request       Request        `gorm:"foreignKey:RequestID" json:"request"`
	AgencyUser    uint           `gorm:"not null" json:"agencyUser"`
	DeliveryAgent User           `gorm:"foreignKey:AgencyUser" json:"deliveryAgent"`
	DeliveredDate time.Time      `json:"deliveredDate"`
	Status        string         `gorm:"size:30;default:'DELIVERED'" json:"status"`
	BillUrl       string         `gorm:"type:longtext" json:"billUrl"`
	BillNotes     string         `gorm:"type:text" json:"billNotes"`
	Items         []DeliveryItem `gorm:"foreignKey:DeliveryID" json:"items"`
}

type DeliveryItem struct {
	ID               uint              `gorm:"primaryKey" json:"id"`
	DeliveryID       uint              `gorm:"not null" json:"deliveryId"`
	ProductID        uint              `gorm:"not null" json:"productId"`
	Product          Product           `gorm:"foreignKey:ProductID" json:"product"`
	ApprovedQty      int               `gorm:"not null;default:0" json:"approvedQty"`
	DeliveredQty     int               `gorm:"not null;default:0" json:"deliveredQty"`
	UnavailableQty   int               `gorm:"not null;default:0" json:"unavailableQty"`
	UnitPrice        float64           `gorm:"type:decimal(10,2);default:0" json:"unitPrice"`
	PendingQty       int               `gorm:"-" json:"pendingQty"`
	Remarks          string            `gorm:"type:text" json:"remarks"`
	VerificationItem *VerificationItem `gorm:"foreignKey:DeliveryItemID" json:"verificationItem,omitempty"`
}

type VerificationItem struct {
	ID             uint   `gorm:"primaryKey" json:"id"`
	DeliveryItemID uint   `gorm:"not null" json:"deliveryItemId"`
	AcceptedQty    int    `gorm:"not null;default:0" json:"acceptedQty"`
	DamagedQty     int    `gorm:"not null;default:0" json:"damagedQty"`
	NotReceivedQty int    `gorm:"not null;default:0" json:"notReceivedQty"`
	Remarks        string `gorm:"type:text" json:"remarks"`
}

type ChatMessage struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	RequestID  uint      `gorm:"not null;index" json:"requestId"`
	SenderID   uint      `gorm:"not null" json:"senderId"`
	SenderName string    `gorm:"size:100;not null" json:"senderName"`
	SenderRole string    `gorm:"size:50;not null" json:"senderRole"`
	TargetRole string    `gorm:"size:50;not null" json:"targetRole"`
	Message    string    `gorm:"type:text;not null" json:"message"`
	CreatedAt  time.Time `json:"createdAt"`
}

type SlaSettings struct {
	ID              uint      `gorm:"primaryKey" json:"id"`
	MaxApproveDays  int       `gorm:"not null;default:2" json:"maxApproveDays"`
	MaxDeliveryDays int       `gorm:"not null;default:3" json:"maxDeliveryDays"`
	MaxVerifyDays   int       `gorm:"not null;default:2" json:"maxVerifyDays"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

type Notification struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	UserID    uint      `gorm:"not null" json:"userId"`
	Type      string    `gorm:"size:50;not null" json:"type"`
	Subject   string    `gorm:"size:150;not null" json:"subject"`
	Message   string    `gorm:"type:text;not null" json:"message"`
	IsRead    bool      `gorm:"default:false" json:"isRead"`
	CreatedAt time.Time `json:"createdAt"`
}

type AuditLog struct {
	ID         uint      `gorm:"primaryKey" json:"id"`
	UserID     *uint     `json:"userId"`
	UserName   string    `gorm:"size:100" json:"userName"`
	Action     string    `gorm:"size:100;not null" json:"action"`
	EntityType string    `gorm:"size:50" json:"entityType"`
	EntityID   string    `gorm:"size:50" json:"entityId"`
	Details    string    `gorm:"type:text" json:"details"`
	IPAddress  string    `gorm:"size:45" json:"ipAddress"`
	CreatedAt  time.Time `json:"createdAt"`
}

type EmailQueue struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Recipient string    `gorm:"size:100;not null" json:"recipient"`
	Subject   string    `gorm:"size:150;not null" json:"subject"`
	Body      string    `gorm:"type:text;not null" json:"body"`
	Status    string    `gorm:"size:20;default:'PENDING'" json:"status"`
	Attempts  int       `gorm:"default:0" json:"attempts"`
	CreatedAt time.Time `json:"createdAt"`
}
