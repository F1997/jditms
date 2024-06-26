package models

import (
	"errors"
	"strings"
	"time"

	"jditms/pkg/ctx"
)

// BuiltinComponent represents a builtin component along with its metadata.
type BuiltinComponent struct {
	ID        uint64 `json:"id" gorm:"primaryKey;type:bigint;autoIncrement;comment:'unique identifier'"`
	Ident     string `json:"ident" gorm:"type:varchar(191);not null;index:idx_ident,sort:asc;comment:'identifier of component'"`
	Logo      string `json:"logo" gorm:"type:varchar(191);not null;comment:'logo of component'"`
	Readme    string `json:"readme" gorm:"type:text;not null;comment:'readme of component'"`
	CreatedAt int64  `json:"created_at" gorm:"type:bigint;not null;default:0;comment:'create time'"`
	CreatedBy string `json:"created_by" gorm:"type:varchar(191);not null;default:'';comment:'creator'"`
	UpdatedAt int64  `json:"updated_at" gorm:"type:bigint;not null;default:0;comment:'update time'"`
	UpdatedBy string `json:"updated_by" gorm:"type:varchar(191);not null;default:'';comment:'updater'"`
}

func (bc *BuiltinComponent) TableName() string {
	return "builtin_components"
}

func (bc *BuiltinComponent) Verify() error {
	bc.Ident = strings.TrimSpace(bc.Ident)
	if bc.Ident == "" {
		return errors.New("ident is blank")
	}

	return nil
}

func BuiltinComponentExists(ctx *ctx.Context, bc *BuiltinComponent) (bool, error) {
	var count int64
	err := DB(ctx).Model(bc).Where("ident = ?", bc.Ident).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (bc *BuiltinComponent) Add(ctx *ctx.Context, username string) error {
	if err := bc.Verify(); err != nil {
		return err
	}
	exists, err := BuiltinComponentExists(ctx, bc)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("builtin component already exists")
	}
	now := time.Now().Unix()
	bc.CreatedAt = now
	bc.UpdatedAt = now
	bc.CreatedBy = username
	return Insert(ctx, bc)
}

func (bc *BuiltinComponent) Update(ctx *ctx.Context, req BuiltinComponent) error {
	if err := req.Verify(); err != nil {
		return err
	}

	if bc.Ident != req.Ident {
		exists, err := BuiltinComponentExists(ctx, &req)
		if err != nil {
			return err
		}
		if exists {
			return errors.New("builtin component already exists")
		}
	}
	req.UpdatedAt = time.Now().Unix()

	return DB(ctx).Model(bc).Select("*").Updates(req).Error
}

func BuiltinComponentDels(ctx *ctx.Context, ids []int64) error {
	if len(ids) == 0 {
		return nil
	}
	return DB(ctx).Where("id in ?", ids).Delete(new(BuiltinComponent)).Error
}

func BuiltinComponentGets(ctx *ctx.Context, query string, limit, offset int) ([]*BuiltinComponent, error) {
	session := DB(ctx)
	if query != "" {
		queryPattern := "%" + query + "%"
		session = session.Where("ident LIKE ?", queryPattern)
	}

	var lst []*BuiltinComponent

	err := session.Limit(limit).Offset(offset).Find(&lst).Error

	return lst, err
}

func BuiltinComponentCount(ctx *ctx.Context, query string) (int64, error) {
	session := DB(ctx).Model(&BuiltinComponent{})
	if query != "" {
		queryPattern := "%" + query + "%"
		session = session.Where("ident LIKE ?", queryPattern)
	}

	var cnt int64
	err := session.Count(&cnt).Error

	return cnt, err
}

func BuiltinComponentGet(ctx *ctx.Context, where string, args ...interface{}) (*BuiltinComponent, error) {
	var lst []*BuiltinComponent
	err := DB(ctx).Where(where, args...).Find(&lst).Error
	if err != nil {
		return nil, err
	}

	if len(lst) == 0 {
		return nil, nil
	}

	return lst[0], nil
}
