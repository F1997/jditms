package router

import (
	"strconv"
	"time"

	"jditms/models"

	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/ginx"
	"github.com/toolkits/pkg/logger"
	"github.com/toolkits/pkg/str"
)

type collectForm struct {
	Name     string                   `json:"name"`
	Queries  []map[string]interface{} `json:"queries"`
	GroupID  int64                    `json:"group_id"`
	Disabled int64                    `json:"disabled"`
	Cate     string                   `json:"cate"`
	Content  string                   `json:"content"`
}

type collectListReq struct {
	Name    string `json:"name"`
	GroupID string `json:"group_id"`
	Cate    string `json:"cate"`
}

func (rt *Router) collectAddPost(c *gin.Context) {
	// 绑定参数
	var f collectForm
	ginx.BindJSON(c, &f)

	// 当前账户名
	username := Username(c)

	collect := models.Collect{
		Name:        f.Name,
		Queries:     "",
		QueriesJson: f.Queries,
		GroupID:     f.GroupID,
		Disabled:    f.Disabled,
		Cate:        f.Cate,
		Version:     str.MD5(f.Content),
		Content:     f.Content,
		CreateBy:    username,
		UpdateBy:    username,
	}

	err := collect.FE2DB()
	if err != nil {
		logger.Error("json to string error")
		Dangerous(c, err)
		return
	}
	ginx.NewRender(c).Message(collect.Add(rt.Ctx))
}

func (rt *Router) collectDel(c *gin.Context) {
	var f idsForm
	ginx.BindJSON(c, &f)
	f.Verify()

	for i := 0; i < len(f.Ids); i++ {
		cid := f.Ids[i]

		collect, err := models.GetCollectGetsById(rt.Ctx, cid)
		ginx.Dangerous(err)

		if collect == nil {
			continue
		}

		me := c.MustGet("user").(*models.User)
		if !me.IsAdmin() {
			// check permission
			rt.bgrwCheck(c, collect.GroupID)
		}

		ginx.Dangerous(collect.Del(rt.Ctx))
	}

	ginx.NewRender(c).Message(nil)
}

func (rt *Router) collectUpsert(c *gin.Context) {

	var req models.Collect
	ginx.BindJSON(c, &req)

	username := c.MustGet("username").(string)

	req.UpdateBy = username

	var err error
	var count int64

	if req.Id == 0 {
		req.CreateBy = username

		count, err = models.GetCollectCountBy(rt.Ctx, "", "", req.Name)
		if err != nil {
			Render(c, nil, err)
			return
		}

		if count > 0 {
			Render(c, nil, "name already exists")
			return
		}
		err = req.Add(rt.Ctx)
	} else {
		req.Version = str.MD5(req.Content)
		req.FE2DB()

		err = req.Update(rt.Ctx, "name", "queries", "group_id", "disabled", "cate", "version", "content", "update_at", "update_by")
	}

	Render(c, nil, err)
}

func (rt *Router) collectUpdataStatus(c *gin.Context) {
	var f collectForm
	ginx.BindJSON(c, &f)

	me := c.MustGet("user").(*models.User)

	// 获取查询参数
	collectIDStr := c.Param("id")
	// 转换参数为整数
	collectID, err := strconv.ParseInt(collectIDStr, 10, 64)
	if err != nil {
		// 处理转换错误
		c.JSON(400, gin.H{"error": "Invalid collect ID"})
		return
	}

	collect, err := models.GetCollectGetsById(rt.Ctx, collectID)
	if err != nil {
		// Render(c, collect, err)
		c.JSON(400, gin.H{"error": "Invalid collect ID"})
	}

	// check permission
	if !me.IsAdmin() {
		rt.bgrwCheck(c, collect.GroupID)
	}

	collect.Disabled = f.Disabled
	collect.UpdateBy = me.Username
	collect.UpdateAt = time.Now().Unix()

	err = collect.Update(rt.Ctx, "disabled", "updated_by", "updated_at")
	ginx.NewRender(c).Data(collect, err)
}

func (rt *Router) collectList(c *gin.Context) {

	var req collectListReq
	ginx.BindJSON(c, &req)

	group_id := req.GroupID
	cate := req.Cate
	name := req.Name

	list, err := models.GetCollectGetsBy(rt.Ctx, group_id, cate, name, "", "")
	Render(c, list, err)
}

func (rt *Router) collectGet(c *gin.Context) {
	// 获取查询参数
	collectIDStr := c.Param("id")
	// 转换参数为整数
	collectID, err := strconv.ParseInt(collectIDStr, 10, 64)
	if err != nil {
		// 处理转换错误
		c.JSON(400, gin.H{"error": "Invalid collect ID"})
		return
	}

	collect, err := models.GetCollectGetsById(rt.Ctx, collectID)
	Render(c, collect, err)
}

func (rt *Router) collectGets(c *gin.Context) {
	bgid := ginx.UrlParamInt64(c, "id")
	query := ginx.QueryStr(c, "query", "")

	collects, err := models.CollectGetsByGroupId(rt.Ctx, bgid, query)

	ginx.NewRender(c).Data(collects, err)
}

type ConfigFormat string

const (
	YamlFormat ConfigFormat = "yaml"
	TomlFormat ConfigFormat = "toml"
	JsonFormat ConfigFormat = "json"
)

type ConfigWithFormat struct {
	Config   string       `json:"config"`
	Format   ConfigFormat `json:"format"`
	checkSum string       `json:"-"`
}

type httpProviderResponse struct {
	// version is signature/md5 of current Config, server side should deal with the Version calculate
	Version string `json:"version"`

	// ConfigMap (InputName -> Config), if version is identical, server side can set Config to nil
	Configs map[string]map[string]*ConfigWithFormat `json:"configs"`
}

func (rt *Router) collectGetByAgentHostname(c *gin.Context) {
	hostname := c.Query("agent_hostname")
	lst, err := models.GetCollectGetsBy(rt.Ctx, "", "", "", hostname, "")
	if err != nil {
		logger.Error(" E! failed to exec mysql query:", err)
	}

	// Convert the query result to the desired response structure
	resp := httpProviderResponse{
		Version: "some_version",
		Configs: make(map[string]map[string]*ConfigWithFormat),
	}

	version := ""
	for _, result := range lst {

		// 创建数据库分类（例如：mysql、nginx）的映射
		categoryMap, exists := resp.Configs[result.Cate]
		if !exists {
			categoryMap = make(map[string]*ConfigWithFormat)
			resp.Configs[result.Cate] = categoryMap
		}

		// 将配置项添加到分类映射中
		categoryMap[result.Version] = &ConfigWithFormat{
			Config: result.Content,
			Format: TomlFormat,
		}
		version += result.Version
	}

	resp.Version = str.MD5(version)

	// Render(c, resp, err)
	c.JSON(200, resp)

}

func (rt *Router) collectAdd(c *gin.Context) {
	var f collectForm
	ginx.BindJSON(c, &f)

	username := Username(c)
	collect := &models.Collect{
		Name:        f.Name,
		Queries:     "",
		QueriesJson: f.Queries,
		GroupID:     ginx.UrlParamInt64(c, "id"),
		Disabled:    f.Disabled,
		Cate:        f.Cate,
		Version:     str.MD5(f.Content),
		Content:     f.Content,
		CreateBy:    username,
		UpdateBy:    username,
	}

	err := collect.FE2DB()
	if err != nil {
		logger.Error("json to string error")
		Dangerous(c, err)
		return
	}

	err = collect.Add(rt.Ctx)
	ginx.Dangerous(err)

	// if f.Content != "" {
	// 	ginx.Dangerous(models.BoardPayloadSave(rt.Ctx, collect.Id, f.Content))
	// }

	ginx.NewRender(c).Data(collect, nil)
}
