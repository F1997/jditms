package alert

import (
	"context"
	"fmt"

	"github.com/flashcatcloud/ibex/src/cmd/ibex"
	"jditms/alert/aconf"
	"jditms/alert/astats"
	"jditms/alert/dispatch"
	"jditms/alert/eval"
	"jditms/alert/naming"
	"jditms/alert/process"
	"jditms/alert/queue"
	"jditms/alert/record"
	"jditms/alert/router"
	"jditms/alert/sender"
	"jditms/conf"
	"jditms/dumper"
	"jditms/memsto"
	"jditms/models"
	"jditms/pkg/ctx"
	"jditms/pkg/httpx"
	"jditms/pkg/logx"
	"jditms/prom"
	"jditms/pushgw/pconf"
	"jditms/pushgw/writer"
	"jditms/tdengine"
)

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
	redis, err = storage.NewRedis(config.Redis)
	if err != nil {
		return nil, err
	}

	syncStats := memsto.NewSyncStats()
	alertStats := astats.NewSyncStats()

	configCache := memsto.NewConfigCache(ctx, syncStats, nil, "")
	targetCache := memsto.NewTargetCache(ctx, syncStats, nil)
	busiGroupCache := memsto.NewBusiGroupCache(ctx, syncStats)
	alertMuteCache := memsto.NewAlertMuteCache(ctx, syncStats)
	alertRuleCache := memsto.NewAlertRuleCache(ctx, syncStats)
	notifyConfigCache := memsto.NewNotifyConfigCache(ctx, configCache)
	dsCache := memsto.NewDatasourceCache(ctx, syncStats)
	userCache := memsto.NewUserCache(ctx, syncStats)
	userGroupCache := memsto.NewUserGroupCache(ctx, syncStats)
	taskTplsCache := memsto.NewTaskTplCache(ctx)

	promClients := prom.NewPromClient(ctx)
	tdengineClients := tdengine.NewTdengineClient(ctx, config.Alert.Heartbeat)

	externalProcessors := process.NewExternalProcessors()

	Start(config.Alert, config.Pushgw, syncStats, alertStats, externalProcessors, targetCache, busiGroupCache, alertMuteCache, alertRuleCache, notifyConfigCache, taskTplsCache, dsCache, ctx, promClients, tdengineClients, userCache, userGroupCache)

	r := httpx.GinEngine(config.Global.RunMode, config.HTTP)
	rt := router.New(config.HTTP, config.Alert, alertMuteCache, targetCache, busiGroupCache, alertStats, ctx, externalProcessors)

	if config.Ibex.Enable {
		ibex.ServerStart(false, nil, redis, config.HTTP.APIForService.BasicAuth, config.Alert.Heartbeat, &config.CenterApi, r, nil, config.Ibex, config.HTTP.Port)
	}

	rt.Config(r)
	dumper.ConfigRouter(r)

	httpClean := httpx.Init(config.HTTP, r)

	return func() {
		logxClean()
		httpClean()
	}, nil
}

func Start(alertc aconf.Alert, pushgwc pconf.Pushgw, syncStats *memsto.Stats, alertStats *astats.Stats, externalProcessors *process.ExternalProcessorsType, targetCache *memsto.TargetCacheType, busiGroupCache *memsto.BusiGroupCacheType,
	alertMuteCache *memsto.AlertMuteCacheType, alertRuleCache *memsto.AlertRuleCacheType, notifyConfigCache *memsto.NotifyConfigCacheType, taskTplsCache *memsto.TaskTplCache, datasourceCache *memsto.DatasourceCacheType, ctx *ctx.Context,
	promClients *prom.PromClientMap, tdendgineClients *tdengine.TdengineClientMap, userCache *memsto.UserCacheType, userGroupCache *memsto.UserGroupCacheType) {
	alertSubscribeCache := memsto.NewAlertSubscribeCache(ctx, syncStats)
	recordingRuleCache := memsto.NewRecordingRuleCache(ctx, syncStats)
	targetsOfAlertRulesCache := memsto.NewTargetOfAlertRuleCache(ctx, alertc.Heartbeat.EngineName, syncStats)

	go models.InitNotifyConfig(ctx, alertc.Alerting.TemplatesDir)

	naming := naming.NewNaming(ctx, alertc.Heartbeat, alertStats)

	writers := writer.NewWriters(pushgwc)
	record.NewScheduler(alertc, recordingRuleCache, promClients, writers, alertStats)

	eval.NewScheduler(alertc, externalProcessors, alertRuleCache, targetCache, targetsOfAlertRulesCache,
		busiGroupCache, alertMuteCache, datasourceCache, promClients, tdendgineClients, naming, ctx, alertStats)

	dp := dispatch.NewDispatch(alertRuleCache, userCache, userGroupCache, alertSubscribeCache, targetCache, notifyConfigCache, taskTplsCache, alertc.Alerting, ctx, alertStats)
	consumer := dispatch.NewConsumer(alertc.Alerting, ctx, dp)

	go dp.ReloadTpls()
	go consumer.LoopConsume()

	go queue.ReportQueueSize(alertStats)
	go sender.InitEmailSender(notifyConfigCache)
}
