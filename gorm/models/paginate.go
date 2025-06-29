package models

import (
	"math"
	"strconv"

	"gorm.io/gorm"
)

type Paginate struct {
	Page      string `json:"page"`
	PageSize  string `json:"page_size"`
	Offset    int    `json:"-"`
	Total     int64  `json:"total"`
	PageCount int    `json:"page_count"`
}

func (paginate *Paginate) SetPageCount(total int64, pageSize int) {
	paginate.Total = total
	paginate.PageCount = int(math.Ceil(float64(total) / float64(pageSize)))
}

func (paginate *Paginate) SetPagination() {
	page, _ := strconv.Atoi(paginate.Page)
	if page == 0 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(paginate.PageSize)
	switch {
	case pageSize > 100:
		pageSize = 100
	case pageSize <= 0:
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	paginate.PageCount = int(math.Ceil(float64(paginate.Total) / float64(pageSize)))
	paginate.Offset = offset
}

func (paginate *Paginate) Pagination() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		page, _ := strconv.Atoi(paginate.Page)
		if page == 0 {
			page = 1
		}

		pageSize, _ := strconv.Atoi(paginate.PageSize)
		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		paginate.PageCount = int(math.Ceil(float64(paginate.Total) / float64(pageSize)))
		paginate.Offset = offset

		return db.Offset(offset).Limit(pageSize)
	}
}
