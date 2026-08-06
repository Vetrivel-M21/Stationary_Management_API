package domain

type LoginRequest struct {
	Mobile   string `json:"mobile" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token        string `json:"token"`
	RefreshToken string `json:"refreshToken"`
	User         User   `json:"user"`
}

type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

type ResetPasswordRequest struct {
	UserID      uint   `json:"userId" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=6"`
}

type CreateUserRequest struct {
	Name               string `json:"name" binding:"required"`
	Email              string `json:"email" binding:"required,email"`
	Mobile             string `json:"mobile" binding:"required"`
	DefaultPassword    string `json:"defaultPassword" binding:"required,min=6"`
	RoleID             uint   `json:"roleId" binding:"required"`
	BranchID           *uint  `json:"branchId"`
	Department         string `json:"department"`
	ApproverAccessType string `json:"approverAccessType"`
}

type UpdateUserRequest struct {
	Name               string `json:"name"`
	Email              string `json:"email"`
	Mobile             string `json:"mobile"`
	RoleID             uint   `json:"roleId"`
	BranchID           *uint  `json:"branchId"`
	Department         string `json:"department"`
	ApproverAccessType string `json:"approverAccessType"`
	Status             string `json:"status"`
}

type CreateBranchRequest struct {
	Name    string `json:"name" binding:"required"`
	Code    string `json:"code" binding:"required"`
	Address string `json:"address"`
}

type UpdateBranchRequest struct {
	Name    string `json:"name"`
	Code    string `json:"code"`
	Address string `json:"address"`
	Status  string `json:"status"`
}

type CreateProductRequest struct {
	Name        string  `json:"name" binding:"required"`
	Category    string  `json:"category" binding:"required"`
	Unit        string  `json:"unit" binding:"required"`
	UnitPrice   float64 `json:"unitPrice"`
	Description string  `json:"description"`
}

type UpdateProductRequest struct {
	Name        string  `json:"name"`
	Category    string  `json:"category"`
	Unit        string  `json:"unit"`
	UnitPrice   float64 `json:"unitPrice"`
	Description string  `json:"description"`
	Status      string  `json:"status"`
}

type CreateRequestItemInput struct {
	ProductID    uint    `json:"productId" binding:"required"`
	RequestedQty int     `json:"requestedQty" binding:"required,gt=0"`
	UnitPrice    float64 `json:"unitPrice"`
}

type CreateRequestDTO struct {
	BranchID        uint                     `json:"branchId" binding:"required"`
	ApplicantName   string                   `json:"applicantName" binding:"required"`
	ApplicantMobile string                   `json:"applicantMobile" binding:"required"`
	ApplicantEmail  string                   `json:"applicantEmail" binding:"required"`
	Department      string                   `json:"department" binding:"required"`
	Location        string                   `json:"location"`
	Items           []CreateRequestItemInput `json:"items" binding:"required,gt=0"`
}

type ApprovalItemInput struct {
	RequestItemID uint   `json:"requestItemId" binding:"required"`
	ApprovedQty   int    `json:"approvedQty"`
	Remove        bool   `json:"remove"`
	Remarks       string `json:"remarks"`
}

type ProcessApprovalDTO struct {
	Action  string              `json:"action" binding:"required"` // APPROVE or REJECT
	Remarks string              `json:"remarks"`
	Items   []ApprovalItemInput `json:"items"`
}

type DeliveryItemInput struct {
	ProductID      uint    `json:"productId" binding:"required"`
	ApprovedQty    int     `json:"approvedQty"`
	DeliveredQty   int     `json:"deliveredQty"`
	UnavailableQty int     `json:"unavailableQty"`
	UnitPrice      float64 `json:"unitPrice"`
	Remarks        string  `json:"remarks"`
}

type ProcessDeliveryDTO struct {
	DeliveryNotes string              `json:"deliveryNotes"`
	BillUrl       string              `json:"billUrl"`
	BillNotes     string              `json:"billNotes"`
	Items         []DeliveryItemInput `json:"items" binding:"required,gt=0"`
}

type VerificationItemInput struct {
	DeliveryItemID uint   `json:"deliveryItemId" binding:"required"`
	AcceptedQty    int    `json:"acceptedQty"`
	DamagedQty     int    `json:"damagedQty"`
	NotReceivedQty int    `json:"notReceivedQty"`
	Remarks        string `json:"remarks"`
}

type ProcessVerificationDTO struct {
	VerificationNotes string                  `json:"verificationNotes"`
	Items             []VerificationItemInput `json:"items" binding:"required,gt=0"`
}

type SendReminderDTO struct {
	RequestID uint   `json:"requestId" binding:"required"`
	Target    string `json:"target" binding:"required"` // REQUESTER, APPROVER, AGENCY
	Message   string `json:"message"`
}

type SendChatMessageDTO struct {
	Message string `json:"message" binding:"required"`
}

type UpdateSlaSettingsDTO struct {
	MaxApproveDays  int `json:"maxApproveDays" binding:"required,min=1"`
	MaxDeliveryDays int `json:"maxDeliveryDays" binding:"required,min=1"`
	MaxVerifyDays   int `json:"maxVerifyDays" binding:"required,min=1"`
}

type DelayedOrderDTO struct {
	Request       Request `json:"request"`
	DelayDays     int     `json:"delayDays"`
	DelayedStage  string  `json:"delayedStage"`
	TargetRole    string  `json:"targetRole"`
	MaxAllowedDays int    `json:"maxAllowedDays"`
}

type DashboardMetrics struct {
	TotalProducts     int64 `json:"totalProducts"`
	TotalRequests     int64 `json:"totalRequests"`
	PendingApprovals  int64 `json:"pendingApprovals"`
	PendingDeliveries int64 `json:"pendingDeliveries"`
	Completed         int64 `json:"completed"`
	Rejected          int64 `json:"rejected"`
	Delayed           int64 `json:"delayed"`
	DamagedItems      int64 `json:"damagedItems"`
	UnavailableItems  int64 `json:"unavailableItems"`
}
