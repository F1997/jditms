package idents

import (
	"context"
	"fmt"
	"sync"
	"time"

	"jditms/models"
	"jditms/pkg/ctx"
	"jditms/pkg/poster"
	"jditms/storage"

	"github.com/toolkits/pkg/logger"
	"github.com/toolkits/pkg/slice"
)

type Set struct {
	sync.Mutex
	items map[string]struct{}
	redis storage.Redis
	ctx   *ctx.Context
}

func New(ctx *ctx.Context, redis storage.Redis) *Set {
	set := &Set{
		items: make(map[string]struct{}),
		redis: redis,
		ctx:   ctx,
	}

	set.Init()
	return set
}

func (s *Set) Init() {
	go s.LoopPersist()
}

func (s *Set) MSet(items map[string]struct{}) {
	s.Lock()
	defer s.Unlock()
	for ident := range items {
		s.items[ident] = struct{}{}
	}
}

func (s *Set) LoopPersist() {
	for {
		time.Sleep(time.Second)
		s.persist()
	}
}

func (s *Set) persist() {
	var items map[string]struct{}

	s.Lock()
	if len(s.items) == 0 {
		s.Unlock()
		return
	}

	items = s.items
	s.items = make(map[string]struct{})
	s.Unlock()

	s.updateTimestamp(items)
}

func (s *Set) updateTimestamp(items map[string]struct{}) {
	lst := make([]string, 0, 100)
	now := time.Now().Unix()
	num := 0
	for ident := range items {
		lst = append(lst, ident)
		num++
		if num == 100 {
			if err := s.UpdateTargets(lst, now); err != nil {
				logger.Errorf("failed to update targets: %v", err)
			}
			lst = lst[:0]
			num = 0
		}
	}

	if err := s.UpdateTargets(lst, now); err != nil {
		logger.Errorf("failed to update targets: %v", err)
	}
}

type TargetUpdate struct {
	Lst []string `json:"lst"`
	Now int64    `json:"now"`
}

func (s *Set) UpdateTargets(lst []string, now int64) error {
	err := updateTargetsUpdateTs(lst, now, s.redis)
	if err != nil {
		logger.Errorf("failed to update targets:%v update_ts: %v", lst, err)
	}

	if !s.ctx.IsCenter {
		t := TargetUpdate{
			Lst: lst,
			Now: now,
		}
		err := poster.PostByUrls(s.ctx, "/v1/jditms/target-update", t)
		return err
	}

	count := int64(len(lst))
	if count == 0 {
		return nil
	}

	ret := s.ctx.DB.Table("target").Where("ident in ?", lst).Update("update_at", now)
	if ret.Error != nil {
		return ret.Error
	}

	// if ret.RowsAffected == count {
	// 	return nil
	// }

	// there are some idents not found in db, so insert them
	var exists []string
	err = s.ctx.DB.Table("target").Where("ident in ?", lst).Pluck("ident", &exists).Error
	if err != nil {
		return err
	}

	news := slice.SubString(lst, exists)
	for i := 0; i < len(news); i++ {
		err = s.ctx.DB.Exec("INSERT INTO target(ident, update_at) VALUES(?, ?)", news[i], now).Error
		if err != nil {
			logger.Error("failed to insert target:", news[i], "error:", err)
		}
	}

	ret_resource := s.ctx.DB.Table("resource").Where("resource_cate in ?", []string{"host", "switch", "router", "friewall"}).Where("ident in ?", lst).Update("update_at", now)

	if ret_resource.Error != nil {
		return ret.Error
	}

	if ret_resource.RowsAffected == count {
		return nil
	}

	var exists_resource []string
	err = s.ctx.DB.Table("resource").Where("ident in ?", lst).Pluck("ident", &exists_resource).Error
	if err != nil {
		return err
	}
	new_resources := slice.SubString(lst, exists_resource)
	for i := 0; i < len(new_resources); i++ {
		fmt.Println(new_resources[i])
		err = s.ctx.DB.Exec("INSERT INTO resource(ident, name,group_id,resource_cate,resource_info,create_at, update_at) VALUES(?,?,?,?,?,?,?)", new_resources[i], new_resources[i], 0, "host", "", now, now).Error
		if err != nil {
			logger.Error("failed to insert resource:", new_resources[i], "error:", err)
		}
	}

	return nil
}

func updateTargetsUpdateTs(lst []string, now int64, redis storage.Redis) error {
	if redis == nil {
		return fmt.Errorf("redis is nil")
	}

	count := int64(len(lst))
	if count == 0 {
		return nil
	}

	newMap := make(map[string]interface{}, count)
	for _, ident := range lst {
		hostUpdateTime := models.HostUpdteTime{
			UpdateTime: now,
			Ident:      ident,
		}
		newMap[models.WrapIdentUpdateTime(ident)] = hostUpdateTime
	}

	err := storage.MSet(context.Background(), redis, newMap)
	return err
}
