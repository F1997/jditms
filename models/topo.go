package models

import (
	"encoding/json"
	"jditms/pkg/ctx"
	"log"
	"time"
)

// Topo 表示数据库中的 Topo 表
type Topo struct {
	Id           int64           `json:"id" gorm:"primaryKey"` // ID 是主键，自动递增
	Name         string          `json:"name"`                 // Name 是字符串类型，不为空
	Relation     json.RawMessage `json:"relation"`
	DatasourceID int64           `json:"datasource_id"`
	GroupID      int64           `json:"group_id"`  // GroupID 是 bigint 类型，不为空
	CreateAt     int64           `json:"create_at"` // CreateAt 是 bigint 类型，不为空，默认为 '0'
	CreateBy     string          `json:"create_by"` // CreateBy 是字符串类型，不为空，默认为空字符串
	UpdateAt     int64           `json:"update_at"` // UpdateAt 是 bigint 类型，不为空，默认为 '0'
	UpdateBy     string          `json:"update_by"` // UpdateBy 是字符串类型，不为空，默认为空字符串
}

func (t *Topo) TableName() string {
	return "topo"
}

func (t *Topo) Add(ctx *ctx.Context) error {
	now := time.Now().Unix()
	t.UpdateAt = now
	t.CreateAt = now
	return Insert(ctx, t)
}

func GetTopoGetsBy(ctx *ctx.Context, group_id, name string) ([]*Topo, error) {
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

	var lst []*Topo
	err := session.Order("id desc").Find(&lst).Error
	if err != nil {
		log.Println(err)
	}
	return lst, err
}

func GetTopoGetsById(ctx *ctx.Context, id int64) (*Topo, error) {
	var topo *Topo
	err := DB(ctx).Model(&Topo{}).First(&topo, id).Error
	if err != nil {
		log.Println(err)
	}
	return topo, err
}

func GetTopoCountBy(ctx *ctx.Context, name string) (int64, error) {
	session := DB(ctx).Model(&Topo{})

	if name != "" {
		// arr := strings.Fields(name)
		// for i := 0; i < len(arr); i++ {
		// 	qarg := "%" + arr[i] + "%"
		// 	session = session.Where("name = ?", qarg)
		// }
		session = session.Where("name = ?", name)
	}

	return Count(session)
}

func (t *Topo) Update(ctx *ctx.Context, selectField interface{}, selectFields ...interface{}) error {

	t.UpdateAt = time.Now().Unix()

	return DB(ctx).Model(t).Select(selectField, selectFields...).Updates(t).Error
}

func (c *Topo) Del(ctx *ctx.Context) error {
	return DB(ctx).Where("id=?", c.Id).Delete(&Topo{}).Error
}
