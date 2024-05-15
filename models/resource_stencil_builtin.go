package models

import (
	"encoding/json"
	"jditms/pkg/ctx"
)

type ResourceStencilBuiltIn struct {
	Id                      int64                    `json:"id" gorm:"primaryKey"` // ID 是主键，自动递增
	Name                    string                   `json:"name"`                 // Name 是字符串类型，不为空
	IconUrl                 string                   `json:"icon_url"`
	ResourceStencilInfo     string                   `json:"-", gorm:"resource_stencil_info"`
	ResourceStencilInfoJson []map[string]interface{} `json:"resource_stencil_info" gorm:"-"`
}

func (rs *ResourceStencilBuiltIn) TableName() string {
	return "resource_stencil_builtin"
}

func GetResourceStencilBuiltInGetsBy(ctx *ctx.Context, name string) ([]*ResourceStencilBuiltIn, error) {
	session := DB(ctx)

	if name != "" {
		session = session.Where("name = ?", name)
	}

	var lst []*ResourceStencilBuiltIn
	err := session.Order("id asc").Find(&lst).Error
	if err == nil {
		for i := 0; i < len(lst); i++ {
			lst[i].DB2FE()
		}
	}
	return lst, err
}

func (r *ResourceStencilBuiltIn) DB2FE() error {
	if r.ResourceStencilInfo != "" {
		err := json.Unmarshal([]byte(r.ResourceStencilInfo), &r.ResourceStencilInfoJson)
		if err != nil {
			return err
		}
	}

	return nil
}
