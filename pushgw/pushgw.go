package pushgw

import (
	"context"
	"fmt"

	"jditms/center/metas"
	"jditms/conf"
	"jditms/memsto"
	"jditms/pkg/ctx"
	"jditms/pkg/httpx"
	"jditms/pkg/logx"
	"jditms/pushgw/idents"
	"jditms/pushgw/router"
	"jditms/pushgw/writer"
	"jditms/storage"
)

type PushgwProvider struct {
	Ident  *idents.Set
	Router *router.Router
}

func Initialize(configDir string, cryptoKey string) (func(), error) {
	config, err := conf.InitConfig(configDir, cryptoKey)
	if err != nil {
		return nil, fmt.Errorf("failed to init config: %v", err)
	}

	logxClean, err := logx.Init(config.Log)
	if err != nil {
		return nil, err
	}

	ctx := ctx.NewContext(context.Background(), nil, false, config.CenterApi)

	var redis storage.Redis
	if config.Redis.Address != "" {
		redis, err = storage.NewRedis(config.Redis)
		if err != nil {
			return nil, err
		}
	}
	idents := idents.New(ctx, redis)
	metas := metas.New(redis)

	stats := memsto.NewSyncStats()

	busiGroupCache := memsto.NewBusiGroupCache(ctx, stats)
	targetCache := memsto.NewTargetCache(ctx, stats, nil)

	writers := writer.NewWriters(config.Pushgw)

	r := httpx.GinEngine(config.Global.RunMode, config.HTTP)
	rt := router.New(config.HTTP, config.Pushgw, config.Alert, targetCache, busiGroupCache, idents, metas, writers, ctx)
	rt.Config(r)

	httpClean := httpx.Init(config.HTTP, r)

	return func() {
		logxClean()
		httpClean()
	}, nil
}
