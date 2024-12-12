package utils

import "gorm.io/gorm"

// IsDeleteSoft 软删除
func IsDeleteSoft(db *gorm.DB) *gorm.DB {
	return db.Where("is_deleted = ?", 0)
}
