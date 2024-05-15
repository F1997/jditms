package models

import (
	"encoding/json"
	"jditms/pkg/ctx"
	"log"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
)

type Resource struct {
	Id   int64  `json:"id" gorm:"primaryKey"` // ID 是主键，自动递增
	Name string `json:"name"`                 // Name 是字符串类型，不为空
	// ResourceQueryIndicators string `json:"resource_query_indicators"` // Name 是字符串类型，不为空
	// ResourceIP        string                   `json:"resource_ip"`          // Name 是字符串类型，不为空
	ResourceInfo          string                   `json:"-", gorm:"resource_info"`
	ResourceInfoJson      []map[string]interface{} `json:"resource_info" gorm:"-"`
	ResourceAlertRule     string                   `json:"-", gorm:"resource_alert_rule"`
	ResourceAlertRuleJson []string                 `json:"resource_alert_rule" gorm:"-"`
	ResourceDashboard     string                   `json:"resource_dashboard"`
	ResourceCate          string                   `json:"resource_cate"`
	GroupID               int64                    `json:"group_id"`         // GroupID 是 bigint 类型，不为空
	Ident                 string                   `json:"ident"`            // Ident 是 字符串类型，不为空，
	ResourceStencil       string                   `json:"resource_stencil"` // resource_stencil 是 字符串类型，默认值为 ''，
	CreateAt              int64                    `json:"create_at"`        // CreateAt 是 bigint 类型，不为空，默认为 '0'
	CreateBy              string                   `json:"create_by"`        // CreateBy 是字符串类型，不为空，默认为空字符串
	UpdateAt              int64                    `json:"update_at"`        // UpdateAt 是 bigint 类型，不为空，默认为 '0'
	UpdateBy              string                   `json:"update_by"`        // UpdateBy 是字符串类型，不为空，默认为空字符串
}

func GetResourceGetsBy(ctx *ctx.Context, group_id, resource_cate, name, resource_ip string) ([]*Resource, error) {
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

	if resource_cate != "" {
		session = session.Where("resource_cate = ?", resource_cate)
	}

	if resource_ip != "" {
		session = session.Where("disabled = ?", resource_ip)
	}

	var lst []*Resource
	err := session.Order("id desc").Find(&lst).Error
	if err == nil {
		for i := 0; i < len(lst); i++ {
			lst[i].DB2FE()
		}
	}
	return lst, err
}

func GetResourceGetsById(ctx *ctx.Context, id int64) (*Resource, error) {
	var resource *Resource
	err := DB(ctx).Model(&Resource{}).First(&resource, id).Error
	if err == nil {
		resource.DB2FE()
	}
	return resource, err
}

func GetResourceCountBy(ctx *ctx.Context, resource_cate, name string) (int64, error) {
	session := DB(ctx).Model(&Resource{})

	if name != "" {
		// arr := strings.Fields(name)
		// for i := 0; i < len(arr); i++ {
		// 	qarg := "%" + arr[i] + "%"
		// 	session = session.Where("name = ?", qarg)
		// }
		session = session.Where("name = ?", name)
	}

	if resource_cate != "" {
		session = session.Where("resource_cate = ?", resource_cate)
	}

	return Count(session)
}

func (r *Resource) Update(ctx *ctx.Context, selectField interface{}, selectFields ...interface{}) error {

	r.UpdateAt = time.Now().Unix()

	return DB(ctx).Model(r).Select(selectField, selectFields...).Updates(r).Error
}

func (r *Resource) Add(ctx *ctx.Context) error {

	now := time.Now().Unix()
	r.UpdateAt = now
	r.CreateAt = now
	return Insert(ctx, r)
}

func (r *Resource) Del(ctx *ctx.Context) error {
	return DB(ctx).Where("id=?", r.Id).Delete(&Resource{}).Error
}

func (r *Resource) DB2FE() error {
	if r.ResourceInfo != "" {
		err := json.Unmarshal([]byte(r.ResourceInfo), &r.ResourceInfoJson)
		if err != nil {
			return err
		}
	}
	if r.ResourceAlertRule != "" {
		err := json.Unmarshal([]byte(r.ResourceAlertRule), &r.ResourceAlertRuleJson)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *Resource) FE2DB() error {
	if r.ResourceInfoJson != nil {
		q, err := json.Marshal(r.ResourceInfoJson)
		if err != nil {
			return err
		}
		r.ResourceInfo = string(q)
	}
	if r.ResourceAlertRuleJson != nil {
		q, err := json.Marshal(r.ResourceAlertRuleJson)
		if err != nil {
			return err
		}
		r.ResourceAlertRule = string(q)
	}

	return nil
}

func ResourceUpdateBgid(ctx *ctx.Context, ids []int64, bgid int64) error {
	fields := map[string]interface{}{
		"group_id":  bgid,
		"update_at": time.Now().Unix(),
	}

	return DB(ctx).Model(&Resource{}).Where("id in ?", ids).Updates(fields).Error
}

func ResourceIdsFilter(ctx *ctx.Context, ids []int64, where string, args ...interface{}) ([]int64, error) {
	var arr []int64
	if len(ids) == 0 {
		return arr, nil
	}

	err := DB(ctx).Model(&Resource{}).Where("id in ?", ids).Where(where, args...).Pluck("id", &arr).Error
	return arr, err
}

func IdsFilter(ctx *ctx.Context, ids []int64, where string, args ...interface{}) ([]int64, error) {
	var arr []int64
	if len(ids) == 0 {
		return arr, nil
	}

	err := DB(ctx).Model(&Resource{}).Where("id in ?", ids).Where(where, args...).Pluck("ident", &arr).Error
	return arr, err
}

func ResourceDel(ctx *ctx.Context, ids []int64) error {
	if len(ids) == 0 {
		panic("ids empty")
	}
	return DB(ctx).Where("id in ?", ids).Delete(new(Resource)).Error
}

func buildResourceWhere(ctx *ctx.Context, bgids []int64, cate, query string) *gorm.DB {
	session := DB(ctx).Model(&Resource{})
	if len(bgids) > 0 {
		strSlice := make([]string, len(bgids))
		for i, v := range bgids {
			strSlice[i] = strconv.FormatInt(v, 10)
		}
		group_ids := strings.Join(strSlice, ",")
		log.Println(group_ids)
		session = session.Where("group_id in (?)", group_ids)
	}

	if cate != "" {
		session = session.Where("resource_cate = ?", cate)
	}

	if query != "" {
		arr := strings.Fields(query)
		for i := 0; i < len(arr); i++ {
			q := "%" + arr[i] + "%"
			session = session.Where("ident like ? or name like ?", q, q)
		}
	}
	return session
}

func ResourceTotal(ctx *ctx.Context, bgids []int64, cate, query string) (int64, error) {
	return Count(buildResourceWhere(ctx, bgids, cate, query))
}

func ResourceGets(ctx *ctx.Context, bgids []int64, cate, query string, limit, offset int) ([]*Resource, error) {
	var lst []*Resource
	err := buildResourceWhere(ctx, bgids, cate, query).Order("ident").Limit(limit).Offset(offset).Find(&lst).Error
	if err == nil {
		for i := 0; i < len(lst); i++ {
			lst[i].DB2FE()
		}
	}
	return lst, err
}
