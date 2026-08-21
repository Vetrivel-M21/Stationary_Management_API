package repository

import (
	"errors"
	"strings"
	"time"
	"stationery-management/internal/domain"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) FindByMobile(mobile string) (*domain.User, error) {
	if r.db == nil {
		return nil, errors.New("database connection unavailable")
	}

	trimmed := strings.TrimSpace(mobile)
	var digitsOnly strings.Builder
	for _, ch := range trimmed {
		if ch >= '0' && ch <= '9' {
			digitsOnly.WriteRune(ch)
		}
	}
	rawDigits := digitsOnly.String()

	var variations []string
	if trimmed != "" {
		variations = append(variations, trimmed)
	}
	if rawDigits != "" && rawDigits != trimmed {
		variations = append(variations, rawDigits)
	}
	if len(rawDigits) == 10 {
		variations = append(variations, "0"+rawDigits)
	} else if len(rawDigits) == 11 && strings.HasPrefix(rawDigits, "0") {
		variations = append(variations, rawDigits[1:])
	}
	if len(rawDigits) > 10 {
		last10 := rawDigits[len(rawDigits)-10:]
		variations = append(variations, last10, "0"+last10)
	}

	var user domain.User
	err := r.db.Preload("Role").Preload("Branch").Where("mobile IN ?", variations).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindByName(name string) (*domain.User, error) {
	if r.db == nil {
		return nil, errors.New("database connection unavailable")
	}
	trimmed := strings.TrimSpace(name)
	var user domain.User
	err := r.db.Preload("Role").Preload("Branch").Where("LOWER(name) = LOWER(?)", trimmed).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) ExistsByName(name string, excludeUserID uint) (bool, error) {
	if r.db == nil {
		return false, errors.New("database connection unavailable")
	}
	var count int64
	query := r.db.Model(&domain.User{}).Where("LOWER(name) = LOWER(?)", strings.TrimSpace(name))
	if excludeUserID > 0 {
		query = query.Where("id != ?", excludeUserID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

func (r *UserRepository) FindByID(id uint) (*domain.User, error) {
	if r.db == nil {
		return nil, errors.New("database connection unavailable")
	}
	var user domain.User
	err := r.db.Preload("Role").Preload("Branch").Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindAll(search string, page, limit int) ([]domain.User, int64, error) {
	var users []domain.User
	var total int64

	query := r.db.Model(&domain.User{}).Preload("Role").Preload("Branch")
	if search != "" {
		s := "%" + search + "%"
		query = query.Where("name LIKE ? OR email LIKE ? OR mobile LIKE ?", s, s, s)
	}

	query.Count(&total)
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("id desc").Find(&users).Error
	return users, total, err
}

func (r *UserRepository) Create(user *domain.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepository) Update(user *domain.User) error {
	return r.db.Save(user).Error
}
// FindByRoleName returns all active users that have the given role name (e.g. "AGENCY", "APPROVER").
func (r *UserRepository) FindByRoleName(roleName string) ([]domain.User, error) {
	if r.db == nil {
		return nil, errors.New("database connection unavailable")
	}
	var users []domain.User
	err := r.db.
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.name = ? AND users.status = ?", roleName, "ACTIVE").
		Preload("Role").Preload("Branch").
		Find(&users).Error
	return users, err
}

func (r *UserRepository) CountByRoleName(roleName string) (int64, error) {
	if r.db == nil {
		return 0, errors.New("database connection unavailable")
	}
	var count int64
	err := r.db.Model(&domain.User{}).
		Joins("JOIN roles ON roles.id = users.role_id").
		Where("roles.name = ? AND users.status = ?", roleName, "ACTIVE").
		Count(&count).Error
	return count, err
}

func (r *UserRepository) CountByRoleAndDepartment(roleID uint, department string, excludeUserID uint) (int64, error) {
	if r.db == nil {
		return 0, errors.New("database connection unavailable")
	}
	var count int64
	query := r.db.Model(&domain.User{}).
		Where("role_id = ? AND department = ? AND status = ?", roleID, department, "ACTIVE")
	if excludeUserID > 0 {
		query = query.Where("id != ?", excludeUserID)
	}
	err := query.Count(&count).Error
	return count, err
}

func (r *UserRepository) CountByRoleID(roleID uint, excludeUserID uint) (int64, error) {
	if r.db == nil {
		return 0, errors.New("database connection unavailable")
	}
	var count int64
	query := r.db.Model(&domain.User{}).
		Where("role_id = ? AND status = ?", roleID, "ACTIVE")
	if excludeUserID > 0 {
		query = query.Where("id != ?", excludeUserID)
	}
	err := query.Count(&count).Error
	return count, err
}

func (r *UserRepository) Delete(id uint) error {
	if r.db == nil {
		return errors.New("database connection unavailable")
	}
	now := time.Now()
	return r.db.Model(&domain.User{}).Where("id = ?", id).Updates(map[string]interface{}{
		"status":     "INACTIVE",
		"deleted_at": &now,
	}).Error
}


