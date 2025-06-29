package gorm_utls

import (
	"goodin/gorm/models"

	"gorm.io/gorm"
)

func WithPagination(pagination *models.Paginate, tx *gorm.DB) {
	tx.Count(&pagination.Total).Scopes(pagination.Pagination())
}
