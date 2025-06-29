package gorm_utls

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

func WithSearch(request map[string]string, keys map[string]string, tx *gorm.DB) {
	index := 0
	for key, value := range keys {
		_, isSet := request[key]
		if isSet {
			if index > 1 {
				tx.Or(fmt.Sprintf("lower(%s) like ? ", value), "%"+strings.ToLower(request[key])+"%")
			} else {
				tx.Where(fmt.Sprintf("lower(%s) like ? ", value), "%"+strings.ToLower(request[key])+"%")
			}
			index++
		}
	}
}

type SearchHandler map[string]func(value string, tx *gorm.DB)

type WithSearchParam struct {
	Column string
	Value  string
}

func WithSearchV2(param WithSearchParam, handlers SearchHandler, tx *gorm.DB) {
	value, isSet := handlers[param.Column]
	if isSet {
		value(param.Value, tx)
	}
}

func WithSearchAllColumn(request map[string]string, handlers SearchHandler, tx *gorm.DB) {
	index := 0
	for key, value := range handlers {
		_, isSet := request[key]
		if isSet {
			value(request[key], tx)
			index++
		}
	}
}
