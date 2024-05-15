package models

import (
	"encoding/json"
	"jditms/pkg/ctx"
	"time"
)

type ResourceStencil struct {
	Id                      int64                    `json:"id" gorm:"primaryKey"` // ID 是主键，自动递增
	Name                    string                   `json:"name"`                 // Name 是字符串类型，不为空
	ResourceStencilInfo     string                   `json:"-", gorm:"resource_stencil_info"`
	ResourceStencilInfoJson []map[string]interface{} `json:"resource_stencil_info" gorm:"-"`
	ResourceStencilCate     string                   `json:"resource_stencil_cate"`
	GroupID                 int64                    `json:"group_id"`  // GroupID 是 bigint 类型，不为空
	CreateAt                int64                    `json:"create_at"` // CreateAt 是 bigint 类型，不为空，默认为 '0'
	CreateBy                string                   `json:"create_by"` // CreateBy 是字符串类型，不为空，默认为空字符串
	UpdateAt                int64                    `json:"update_at"` // UpdateAt 是 bigint 类型，不为空，默认为 '0'
	UpdateBy                string                   `json:"update_by"` // UpdateBy 是字符串类型，不为空，默认为空字符串
}

func (rs *ResourceStencil) TableName() string {
	return "resource_stencil"
}

func (rs *ResourceStencil) Add(ctx *ctx.Context) error {

	now := time.Now().Unix()
	rs.UpdateAt = now
	rs.CreateAt = now
	return Insert(ctx, rs)
}

func (r *ResourceStencil) Del(ctx *ctx.Context) error {
	return DB(ctx).Where("id=?", r.Id).Delete(&ResourceStencil{}).Error
}

func (r *ResourceStencil) Update(ctx *ctx.Context, selectField interface{}, selectFields ...interface{}) error {

	r.UpdateAt = time.Now().Unix()

	return DB(ctx).Model(r).Select(selectField, selectFields...).Updates(r).Error
}

func GetResourceStencilGetsBy(ctx *ctx.Context, group_id, resource_stencil_cate, name string) ([]*ResourceStencil, error) {
	session := DB(ctx)

	if name != "" {
		session = session.Where("name = ?", name)
	}

	if group_id != "" {
		session = session.Where("group_id = ?", group_id)
	}

	if resource_stencil_cate != "" {
		session = session.Where("resource_stencil_cate = ?", resource_stencil_cate)
	}

	var lst []*ResourceStencil
	err := session.Order("id desc").Find(&lst).Error
	if err == nil {
		for i := 0; i < len(lst); i++ {
			lst[i].DB2FE()
		}
	}
	return lst, err
}

func GetResourceStencilCountBy(ctx *ctx.Context, resource_stencil_cate, name string) (int64, error) {
	session := DB(ctx).Model(&ResourceStencil{})

	if name != "" {
		// arr := strings.Fields(name)
		// for i := 0; i < len(arr); i++ {
		// 	qarg := "%" + arr[i] + "%"
		// 	session = session.Where("name = ?", qarg)
		// }
		session = session.Where("name = ?", name)
	}

	if resource_stencil_cate != "" {
		session = session.Where("resource_stencil_cate = ?", resource_stencil_cate)
	}

	return Count(session)
}

func GetResourceStencilGetsById(ctx *ctx.Context, id int64) (*ResourceStencil, error) {
	var resource *ResourceStencil
	err := DB(ctx).Model(&ResourceStencil{}).First(&resource, id).Error
	if err == nil {
		resource.DB2FE()
	}
	return resource, err
}

func (r *ResourceStencil) FE2DB() error {
	if r.ResourceStencilInfoJson != nil {
		q, err := json.Marshal(r.ResourceStencilInfoJson)
		if err != nil {
			return err
		}
		r.ResourceStencilInfo = string(q)
	}

	return nil
}

func (r *ResourceStencil) DB2FE() error {
	if r.ResourceStencilInfo != "" {
		err := json.Unmarshal([]byte(r.ResourceStencilInfo), &r.ResourceStencilInfoJson)
		if err != nil {
			return err
		}
	}

	return nil
}
