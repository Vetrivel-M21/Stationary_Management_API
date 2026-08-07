package repository

import (
	"fmt"
	"log"
	"stationery-management/internal/config"
	"stationery-management/internal/domain"
	"stationery-management/pkg/hash"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	// First connect without DB name to ensure database exists
	baseDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort)
	
	baseDB, err := gorm.Open(mysql.Open(baseDSN), &gorm.Config{})
	if err == nil {
		createDBSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;", cfg.DBName)
		baseDB.Exec(createDBSQL)
		sqlDB, _ := baseDB.DB()
		if sqlDB != nil {
			sqlDB.Close()
		}
	}

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	// Safely clean up legacy FK constraints only if present in MySQL
	var count1, count2 int64
	db.Raw("SELECT COUNT(*) FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'users' AND constraint_name = 'users_ibfk_1'").Scan(&count1)
	if count1 > 0 {
		_ = db.Exec("ALTER TABLE `users` DROP FOREIGN KEY `users_ibfk_1`").Error
	}
	db.Raw("SELECT COUNT(*) FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'users' AND constraint_name = 'users_ibfk_2'").Scan(&count2)
	if count2 > 0 {
		_ = db.Exec("ALTER TABLE `users` DROP FOREIGN KEY `users_ibfk_2`").Error
	}

	_ = db.Exec("ALTER TABLE `roles` MODIFY `id` BIGINT UNSIGNED AUTO_INCREMENT").Error
	_ = db.Exec("ALTER TABLE `branches` MODIFY `id` BIGINT UNSIGNED AUTO_INCREMENT").Error
	_ = db.Exec("ALTER TABLE `users` MODIFY `role_id` BIGINT UNSIGNED NOT NULL").Error
	_ = db.Exec("ALTER TABLE `users` MODIFY `branch_id` BIGINT UNSIGNED").Error

	// Auto-migrate tables
	err = db.AutoMigrate(
		&domain.Role{},
		&domain.Branch{},
		&domain.User{},
		&domain.Product{},
		&domain.Request{},
		&domain.RequestItem{},
		&domain.ApprovalItem{},
		&domain.Delivery{},
		&domain.DeliveryItem{},
		&domain.VerificationItem{},
		&domain.ChatMessage{},
		&domain.SlaSettings{},
		&domain.Notification{},
		&domain.AuditLog{},
	)
	if err != nil {
		log.Printf("AutoMigrate warning: %v\n", err)
	}

	// Explicitly create tables if not existing
	if !db.Migrator().HasTable(&domain.ChatMessage{}) {
		_ = db.Migrator().CreateTable(&domain.ChatMessage{})
	}
	if !db.Migrator().HasTable(&domain.SlaSettings{}) {
		_ = db.Migrator().CreateTable(&domain.SlaSettings{})
	}

	// Explicitly migrate missing columns for existing MySQL tables
	if !db.Migrator().HasColumn(&domain.User{}, "department") {
		_ = db.Migrator().AddColumn(&domain.User{}, "department")
	}
	if !db.Migrator().HasColumn(&domain.Request{}, "applicant_name") {
		_ = db.Migrator().AddColumn(&domain.Request{}, "applicant_name")
	}
	if !db.Migrator().HasColumn(&domain.Request{}, "applicant_mobile") {
		_ = db.Migrator().AddColumn(&domain.Request{}, "applicant_mobile")
	}
	if !db.Migrator().HasColumn(&domain.Request{}, "applicant_email") {
		_ = db.Migrator().AddColumn(&domain.Request{}, "applicant_email")
	}
	if !db.Migrator().HasColumn(&domain.Request{}, "department") {
		_ = db.Migrator().AddColumn(&domain.Request{}, "department")
	}
	if !db.Migrator().HasColumn(&domain.Request{}, "location") {
		_ = db.Migrator().AddColumn(&domain.Request{}, "location")
	}
	if !db.Migrator().HasColumn(&domain.Product{}, "unit_price") {
		_ = db.Migrator().AddColumn(&domain.Product{}, "unit_price")
	}
	if !db.Migrator().HasColumn(&domain.RequestItem{}, "unit_price") {
		_ = db.Migrator().AddColumn(&domain.RequestItem{}, "unit_price")
	}
	if !db.Migrator().HasColumn(&domain.Delivery{}, "bill_url") {
		_ = db.Migrator().AddColumn(&domain.Delivery{}, "bill_url")
	}
	if !db.Migrator().HasColumn(&domain.Delivery{}, "bill_notes") {
		_ = db.Migrator().AddColumn(&domain.Delivery{}, "bill_notes")
	}
	if !db.Migrator().HasColumn(&domain.User{}, "deleted_at") {
		_ = db.Migrator().AddColumn(&domain.User{}, "deleted_at")
	}
	if !db.Migrator().HasColumn(&domain.Branch{}, "deleted_at") {
		_ = db.Migrator().AddColumn(&domain.Branch{}, "deleted_at")
	}
	if !db.Migrator().HasColumn(&domain.Product{}, "deleted_at") {
		_ = db.Migrator().AddColumn(&domain.Product{}, "deleted_at")
	}
	if !db.Migrator().HasColumn(&domain.Request{}, "deleted_at") {
		_ = db.Migrator().AddColumn(&domain.Request{}, "deleted_at")
	}

	SeedInitialData(db)
	return db, nil
}

func SeedInitialData(db *gorm.DB) {
	var roleCount int64
	db.Model(&domain.Role{}).Count(&roleCount)
	if roleCount == 0 {
		roles := []domain.Role{
			{ID: 1, Name: "ADMIN", Description: "System Administrator"},
			{ID: 2, Name: "BRANCH_REQUESTER", Description: "Branch Requester"},
			{ID: 3, Name: "APPROVER", Description: "Branch Approver"},
			{ID: 4, Name: "AGENCY", Description: "Delivery Agency"},
			{ID: 5, Name: "MONITOR", Description: "Read-only Monitor"},
		}
		db.Create(&roles)
	}

	var branchCount int64
	db.Model(&domain.Branch{}).Count(&branchCount)
	if branchCount == 0 {
		branches := []domain.Branch{
			{ID: 1, Name: "Headquarters", Code: "HQ-001", Address: "100 Enterprise Tower, Financial District", Status: "ACTIVE"},
			{ID: 2, Name: "North Region Branch", Code: "BR-101", Address: "45 North Commercial Boulevard", Status: "ACTIVE"},
			{ID: 3, Name: "South Region Branch", Code: "BR-102", Address: "88 South Industrial Park", Status: "ACTIVE"},
			{ID: 4, Name: "East Coast Office", Code: "BR-103", Address: "12 Harbour View Road", Status: "ACTIVE"},
		}
		db.Create(&branches)
	}

	hashedPassword, _ := hash.HashPassword("Admin@123")
	branchHQ := uint(1)

	// Seed or update SLA Settings
	var slaCount int64
	db.Model(&domain.SlaSettings{}).Count(&slaCount)
	if slaCount == 0 {
		sla := domain.SlaSettings{
			ID:              1,
			MaxApproveDays:  2,
			MaxDeliveryDays: 3,
			MaxVerifyDays:   2,
		}
		db.Create(&sla)
	}

	// Define Seed Users - Admin Only
	seedUsers := []domain.User{
		{ID: 1, Name: "System Administrator", Email: "admin@stationery.com", Mobile: "09999999999", Password: hashedPassword, RoleID: 1, BranchID: &branchHQ, ApproverAccessType: "ALL_BRANCHES", Status: "ACTIVE", FirstLogin: false},
	}

	for _, su := range seedUsers {
		var existing domain.User
		if err := db.Where("mobile = ? OR email = ?", su.Mobile, su.Email).First(&existing).Error; err != nil {
			db.Create(&su)
		} else {
			// Ensure password and department are up to date
			updates := map[string]interface{}{
				"name":       su.Name,
				"department": su.Department,
				"password":   hashedPassword,
			}
			db.Model(&existing).Updates(updates)
		}
	}

	var productCount int64
	db.Model(&domain.Product{}).Count(&productCount)
	if productCount == 0 {
		products := []domain.Product{
			{ID: 1, Name: "Ballpoint Pen - Blue (Box of 10)", Category: "Writing Instruments", Unit: "Box", UnitPrice: 120.00, Description: "High-quality blue ink ballpoint pens 0.7mm", Status: "ACTIVE"},
			{ID: 2, Name: "A4 Printing Paper (80gsm - 500 Sheets)", Category: "Paper Products", Unit: "Ream", UnitPrice: 280.00, Description: "Premium white multipurpose copy paper", Status: "ACTIVE"},
			{ID: 3, Name: "Permanent Marker - Black", Category: "Writing Instruments", Unit: "Piece", UnitPrice: 35.00, Description: "Chisel tip waterproof black permanent marker", Status: "ACTIVE"},
			{ID: 4, Name: "Heavy Duty Stapler No. 10", Category: "Desk Supplies", Unit: "Piece", UnitPrice: 180.00, Description: "Durable metal body desk stapler", Status: "ACTIVE"},
			{ID: 5, Name: "Sticky Notes 3x3 Yellow (100 Sheets)", Category: "Paper Products", Unit: "Pad", UnitPrice: 45.00, Description: "Standard self-adhesive memo pads", Status: "ACTIVE"},
			{ID: 6, Name: "Expandable File Folder A4", Category: "Filing & Storage", Unit: "Piece", UnitPrice: 65.00, Description: "Heavy-duty poly expandable document organizer", Status: "ACTIVE"},
			{ID: 7, Name: "12-Digit Desk Calculator", Category: "Electronics", Unit: "Piece", UnitPrice: 450.00, Description: "Dual-power solar and battery desktop calculator", Status: "ACTIVE"},
			{ID: 8, Name: "Paper Clips (100 pcs/box)", Category: "Desk Supplies", Unit: "Box", UnitPrice: 25.00, Description: "Vinyl coated rust-resistant paper clips", Status: "ACTIVE"},
		}
		db.Create(&products)
	}
}
