package gorm_utls

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func WithValue(column, value string, tx *gorm.DB) {
	tx.Clauses(clause.OrderBy{
		Expression: clause.Expr{
			SQL:  "(? = ?) DESC",
			Vars: []interface{}{clause.Column{Name: column}, value},
		},
	})
}

type WithValueV2Param struct {
	Col string
	Val string
}

func WithValueV2(param WithValueV2Param, handlers SortHandler, tx *gorm.DB) {
	_, isSet := handlers[param.Col]
	if isSet {
		WithValue(param.Col, param.Val, tx)
	}
}
