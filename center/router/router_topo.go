package router

import (
	"encoding/json"
	"jditms/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/ginx"
)

type topoAddForm struct {
	Name         string          `json:"name" binding:"required"`
	DatasourceID int64           `json:"datasource_id" binding:"required"`
	Relation     json.RawMessage `json:"relation" binding:"required"`
	// GroupID  int64           `json:"group_id" binding:"required"`
}

type topoListReq struct {
	Name    string `json:"name"`
	GroupID string `json:"group_id"`
}

func (rt *Router) topoAddPost(c *gin.Context) {
	// 绑定参数
	var f topoAddForm
	ginx.BindJSON(c, &f)

	// 当前账户名
	username := Username(c)

	topo := models.Topo{
		Name:         f.Name,
		DatasourceID: f.DatasourceID,
		Relation:     f.Relation,
		GroupID:      ginx.UrlParamInt64(c, "bgid"),
		CreateBy:     username,
		UpdateBy:     username,
	}

	ginx.NewRender(c).Message(topo.Add(rt.Ctx))
}

func (rt *Router) topoAdd(c *gin.Context) {
	// 绑定参数
	var f topoAddForm
	ginx.BindJSON(c, &f)

	// 当前账户名
	username := Username(c)

	topo := models.Topo{
		Name:         f.Name,
		Relation:     f.Relation,
		DatasourceID: f.DatasourceID,
		GroupID:      ginx.UrlParamInt64(c, "id"),
		CreateBy:     username,
		UpdateBy:     username,
	}

	err := topo.Add(rt.Ctx)
	ginx.Dangerous(err)

	// if f.Content != "" {
	// 	ginx.Dangerous(models.BoardPayloadSave(rt.Ctx, collect.Id, f.Content))
	// }

	ginx.NewRender(c).Data(topo, nil)
	// ginx.NewRender(c).Message(topo.Add(rt.Ctx))
}

// get Topo List
func (rt *Router) topoList(c *gin.Context) {
	var req topoListReq
	ginx.BindJSON(c, &req)

	group_id := req.GroupID
	name := req.Name

	list, err := models.GetTopoGetsBy(rt.Ctx, group_id, name)
	Render(c, list, err)
}

func (rt *Router) topoGets(c *gin.Context) {
	bgid := ginx.UrlParamStr(c, "id")
	// query := ginx.QueryStr(c, "query", "")

	topos, err := models.GetTopoGetsBy(rt.Ctx, bgid, "")

	ginx.NewRender(c).Data(topos, err)
}

func (rt *Router) topoGet(c *gin.Context) {
	// 获取查询参数
	topoIDStr := c.Param("id")
	// 转换参数为整数
	topotID, err := strconv.ParseInt(topoIDStr, 10, 64)
	if err != nil {
		// 处理转换错误
		c.JSON(400, gin.H{"error": "Invalid topo ID"})
		return
	}

	topo, err := models.GetTopoGetsById(rt.Ctx, topotID)

	Render(c, topo, err)
}

func (rt *Router) topoGetStatus(c *gin.Context) {
	// 获取查询参数
	topoIDStr := c.Param("id")
	// 转换参数为整数
	topotID, err := strconv.ParseInt(topoIDStr, 10, 64)
	if err != nil {
		// 处理转换错误
		c.JSON(400, gin.H{"error": "Invalid topo ID"})
		return
	}

	topo, err := models.GetTopoGetsById(rt.Ctx, topotID)

	Render(c, topo, err)
}

func (rt *Router) topoUpsert(c *gin.Context) {

	var req models.Topo
	ginx.BindJSON(c, &req)

	username := c.MustGet("username").(string)

	req.UpdateBy = username

	req.Id = ginx.UrlParamInt64(c, "id")
	var err error
	var count int64

	if req.Id == 0 {
		req.CreateBy = username

		count, err = models.GetTopoCountBy(rt.Ctx, req.Name)
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
		err = req.Update(rt.Ctx, "name", "relation", "datasource_id", "group_id", "update_at", "update_by")
	}
	// log.Println(req.Relation)
	Render(c, nil, err)
}

func (rt *Router) topoDel(c *gin.Context) {
	var f idsForm
	ginx.BindJSON(c, &f)
	f.Verify()

	for i := 0; i < len(f.Ids); i++ {
		topoID := f.Ids[i]

		topo, err := models.GetTopoGetsById(rt.Ctx, topoID)
		ginx.Dangerous(err)

		if topo == nil {
			continue
		}

		me := c.MustGet("user").(*models.User)
		if !me.IsAdmin() {
			// check permission
			rt.bgrwCheck(c, topo.GroupID)
		}

		ginx.Dangerous(topo.Del(rt.Ctx))
	}

	ginx.NewRender(c).Message(nil)
}
