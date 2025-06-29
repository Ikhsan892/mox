package gorm_utls

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func WithSort(request map[string]string, keys map[string]string, tx *gorm.DB) {
	for key, value := range keys {
		_, isSet := request[key]
		if isSet {
			direction := request[key]
			desc := false
			if direction == "descend" {
				desc = true
			}

			tx.Order(clause.OrderByColumn{
				Column: clause.Column{
					Name: value,
				},
				Desc: desc,
			})
		}
	}
}

type SortHandler map[string]func(value string, desc bool, tx *gorm.DB)

func WithDefaultSort(column string) func(value string, desc bool, tx *gorm.DB) {
	return func(value string, desc bool, tx *gorm.DB) {
		tx.Order(clause.OrderByColumn{
			Column: clause.Column{
				Name: column,
			},
			Desc: desc,
		})
	}
}

type WithSortV2Param struct {
	Column    string
	Direction string
}

func WithSortV2(param WithSortV2Param, handlers SortHandler, tx *gorm.DB) {
	value, isSet := handlers[param.Column]
	if isSet {
		direction := param.Direction
		desc := false
		if direction == "descend" || direction == "desc" {
			desc = true
		}
		value(param.Column, desc, tx)
	}
}

func WithSortAllColumn(request map[string]string, handlers SortHandler, tx *gorm.DB) {
	index := 0
	for key, value := range handlers {
		_, isSet := request[key]
		if isSet {
			direction := request[key]
			desc := false
			if direction == "descend" {
				desc = true
			}
			value(request[key], desc, tx)
			index++
		}
	}
}
