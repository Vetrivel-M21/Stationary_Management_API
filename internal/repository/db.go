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
	baseDSN := fmt.Sprintf("%s:%s@tcp(%s:%s)/?charset=utf8mb4&parseTime=True&loc=Local&timeout=3s",
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

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local&timeout=3s",
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
	var count3 int64
	db.Raw("SELECT COUNT(*) FROM information_schema.table_constraints WHERE table_schema = DATABASE() AND table_name = 'request_items' AND constraint_name = 'request_items_ibfk_2'").Scan(&count3)
	if count3 > 0 {
		_ = db.Exec("ALTER TABLE `request_items` DROP FOREIGN KEY `request_items_ibfk_2`").Error
	}

	_ = db.Exec("ALTER TABLE `roles` MODIFY `id` BIGINT UNSIGNED AUTO_INCREMENT").Error
	_ = db.Exec("ALTER TABLE `branches` MODIFY `id` BIGINT UNSIGNED AUTO_INCREMENT").Error
	_ = db.Exec("ALTER TABLE `users` MODIFY `role_id` BIGINT UNSIGNED NOT NULL").Error
	_ = db.Exec("ALTER TABLE `users` MODIFY `branch_id` BIGINT UNSIGNED").Error
	_ = db.Exec("ALTER TABLE `request_items` MODIFY `product_id` BIGINT UNSIGNED NULL").Error

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
	if !db.Migrator().HasColumn(&domain.Request{}, "payment_proof_url") {
		_ = db.Migrator().AddColumn(&domain.Request{}, "payment_proof_url")
	}
	if !db.Migrator().HasColumn(&domain.RequestItem{}, "item_kind") {
		_ = db.Migrator().AddColumn(&domain.RequestItem{}, "item_kind")
		_ = db.Exec("UPDATE `request_items` SET `item_kind` = 'CATALOG' WHERE `item_kind` = '' OR `item_kind` IS NULL").Error
	}
	if !db.Migrator().HasColumn(&domain.RequestItem{}, "item_name") {
		_ = db.Migrator().AddColumn(&domain.RequestItem{}, "item_name")
	}
	if !db.Migrator().HasColumn(&domain.RequestItem{}, "product_name") {
		_ = db.Migrator().AddColumn(&domain.RequestItem{}, "product_name")
		_ = db.Migrator().AddColumn(&domain.RequestItem{}, "product_category")
		_ = db.Migrator().AddColumn(&domain.RequestItem{}, "product_unit")
		_ = db.Exec(`UPDATE request_items ri JOIN products p ON p.id = ri.product_id
			SET ri.product_name = p.name, ri.product_category = p.category, ri.product_unit = p.unit
			WHERE ri.product_name = '' OR ri.product_name IS NULL`).Error
	}
	if !db.Migrator().HasColumn(&domain.DeliveryItem{}, "product_name") {
		_ = db.Migrator().AddColumn(&domain.DeliveryItem{}, "product_name")
		_ = db.Migrator().AddColumn(&domain.DeliveryItem{}, "product_category")
		_ = db.Migrator().AddColumn(&domain.DeliveryItem{}, "product_unit")
		_ = db.Exec(`UPDATE delivery_items di JOIN products p ON p.id = di.product_id
			SET di.product_name = p.name, di.product_category = p.category, di.product_unit = p.unit
			WHERE di.product_name = '' OR di.product_name IS NULL`).Error
	}
	if !db.Migrator().HasColumn(&domain.Request{}, "shop_items_verified_at") {
		_ = db.Migrator().AddColumn(&domain.Request{}, "shop_items_verified_at")
	}
	if !db.Migrator().HasColumn(&domain.Request{}, "shop_bill_urls") {
		_ = db.Migrator().AddColumn(&domain.Request{}, "shop_bill_urls")
	}
	if !db.Migrator().HasColumn(&domain.Request{}, "shop_payment_proof_urls") {
		_ = db.Migrator().AddColumn(&domain.Request{}, "shop_payment_proof_urls")
	}
	if !db.Migrator().HasColumn(&domain.User{}, "deleted_at") {
		_ = db.Migrator().AddColumn(&domain.User{}, "deleted_at")
	}
	if !db.Migrator().HasColumn(&domain.Branch{}, "deleted_at") {
		_ = db.Migrator().AddColumn(&domain.Branch{}, "deleted_at")
	}
	if !db.Migrator().HasColumn(&domain.Branch{}, "department") {
		_ = db.Migrator().AddColumn(&domain.Branch{}, "department")
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
		{ID: 1, Name: "Admin", Email: "admin@stationery.com", Mobile: "9999999999", Password: hashedPassword, RoleID: 1, BranchID: &branchHQ, ApproverAccessType: "ALL_BRANCHES", Status: "ACTIVE", FirstLogin: false},
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

}
