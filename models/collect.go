package models

import (
	"encoding/json"
	"strings"
	"time"

	"jditms/pkg/ctx"
)

type collectQuery struct {
	Key    string        `json:"key"`
	Op     string        `json:"op"`
	Values []interface{} `json:"values"`
}

// Collect 表示数据库中的 collect 表
type Collect struct {
	Id          int64                    `json:"id" gorm:"primaryKey"` // ID 是主键，自动递增
	Name        string                   `json:"name"`                 // Name 是字符串类型，不为空
	Queries     string                   `json:"-", gorm:"queries"`
	QueriesJson []map[string]interface{} `json:"queries" gorm:"-"`
	GroupID     int64                    `json:"group_id"`  // GroupID 是 bigint 类型，不为空
	Disabled    int64                    `json:"disabled"`  // Disabled 是 bigint 类型，不为空，默认为 '0'
	Cate        string                   `json:"cate"`      // Cate 是字符串类型，不为空
	Version     string                   `json:"version"`   // Version 是字符串类型，不为空，默认为空字符串
	Content     string                   `json:"content"`   // Content 是长文本类型，不为空
	CreateAt    int64                    `json:"create_at"` // CreateAt 是 bigint 类型，不为空，默认为 '0'
	CreateBy    string                   `json:"create_by"` // CreateBy 是字符串类型，不为空，默认为空字符串
	UpdateAt    int64                    `json:"update_at"` // UpdateAt 是 bigint 类型，不为空，默认为 '0'
	UpdateBy    string                   `json:"update_by"` // UpdateBy 是字符串类型，不为空，默认为空字符串
}

func (c *Collect) TableName() string {
	return "collect"
}

func (c *Collect) Add(ctx *ctx.Context) error {

	now := time.Now().Unix()
	c.UpdateAt = now
	c.CreateAt = now
	return Insert(ctx, c)
}

func (c *Collect) Del(ctx *ctx.Context) error {
	return DB(ctx).Where("id=?", c.Id).Delete(&Collect{}).Error
}

func CollectGet(ctx *ctx.Context, where string, args ...interface{}) (*Collect, error) {
	var lst []*Collect
	err := DB(ctx).Where(where, args...).Find(&lst).Error
	if err != nil {
		return nil, err
	}

	if len(lst) == 0 {
		return nil, nil
	}

	return lst[0], nil
}

func (c *Collect) Update(ctx *ctx.Context, selectField interface{}, selectFields ...interface{}) error {

	c.UpdateAt = time.Now().Unix()

	return DB(ctx).Model(c).Select(selectField, selectFields...).Updates(c).Error
}

func GetCollectGetsBy(ctx *ctx.Context, group_id, cate, name, queries, status string) ([]*Collect, error) {
	session := DB(ctx)

	if name != "" {
		// arr := strings.Fields(name)
		// for i := 0; i < len(arr); i++ {
		// 	qarg := "%" + arr[i] + "%"
		// 	session = session.Where("name =  ?", qarg)
		// }
		session = session.Where("name = ?", name)
	}

	if group_id != "" {
		session = session.Where("group_id = ?", group_id)
	}

	if cate != "" {
		session = session.Where("cate = ?", cate)
	}

	if status != "" {
		session = session.Where("disabled = ?", status)
	}

	if queries != "" {
		qarg := "%" + queries + "%"
		session = session.Where("queries like ?", qarg)
	}

	var lst []*Collect
	err := session.Order("id desc").Find(&lst).Error
	if err == nil {
		for i := 0; i < len(lst); i++ {
			lst[i].DB2FE()
		}
	}
	return lst, err
}

func GetCollectCountBy(ctx *ctx.Context, cate, version, name string) (int64, error) {
	session := DB(ctx).Model(&Collect{})

	if name != "" {
		// arr := strings.Fields(name)
		// for i := 0; i < len(arr); i++ {
		// 	qarg := "%" + arr[i] + "%"
		// 	session = session.Where("name = ?", qarg)
		// }
		session = session.Where("name = ?", name)
	}

	if version != "" {
		session = session.Where("version = ?", version)
	}

	if cate != "" {
		session = session.Where("cate = ?", cate)
	}

	return Count(session)
}

func GetCollectGetsById(ctx *ctx.Context, id int64) (*Collect, error) {
	var collect *Collect
	err := DB(ctx).Model(&Collect{}).First(&collect, id).Error
	if err == nil {
		collect.DB2FE()
	}
	return collect, err
}

// CollectGets for list page
func CollectGetsByGroupId(ctx *ctx.Context, groupId int64, query string) ([]Collect, error) {
	session := DB(ctx).Where("group_id=?", groupId).Order("name")

	arr := strings.Fields(query)
	if len(arr) > 0 {
		for i := 0; i < len(arr); i++ {
			if strings.HasPrefix(arr[i], "-") {
				q := "%" + arr[i][1:] + "%"
				session = session.Where("name not like ?", q)
			} else {
				q := "%" + arr[i] + "%"
				session = session.Where("name like ? ", q)
			}
		}
	}

	var objs []Collect
	err := session.Find(&objs).Error
	if err == nil {
		for i := 0; i < len(objs); i++ {
			objs[i].DB2FE()
		}
	}
	return objs, err
}

func (c *Collect) DB2FE() error {
	if c.Queries != "" {
		err := json.Unmarshal([]byte(c.Queries), &c.QueriesJson)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Collect) FE2DB() error {
	if c.QueriesJson != nil {
		q, err := json.Marshal(c.QueriesJson)
		if err != nil {
			return err
		}
		c.Queries = string(q)
	}

	return nil
}
