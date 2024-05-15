package router

import (
	"jditms/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/ginx"
)

type resourceStencilForm struct {
	Name                string                   `json:"name"`
	ResourceStencilInfo []map[string]interface{} `json:"resource_stencil_info"`
	ResourceStencilCate string                   `json:"resource_stencil_cate"`
}

func (rt *Router) resourceStencilAdd(c *gin.Context) {
	var f resourceStencilForm
	ginx.BindJSON(c, &f)

	username := Username(c)
	resource_stencil := &models.ResourceStencil{
		Name:                    f.Name,
		ResourceStencilInfo:     "",
		ResourceStencilInfoJson: f.ResourceStencilInfo,
		ResourceStencilCate:     f.ResourceStencilCate,
		GroupID:                 ginx.UrlParamInt64(c, "id"),
		CreateBy:                username,
		UpdateBy:                username,
	}

	err := resource_stencil.FE2DB()
	if err != nil {
		// logger.Error("json to string error")
		Dangerous(c, err)
		return
	}

	ginx.NewRender(c).Message(resource_stencil.Add(rt.Ctx))
}

func (rt *Router) resourceStencilDel(c *gin.Context) {
	var f idsForm
	ginx.BindJSON(c, &f)
	f.Verify()

	for i := 0; i < len(f.Ids); i++ {
		resourceStencilID := f.Ids[i]

		resourceStencil, err := models.GetResourceStencilGetsById(rt.Ctx, resourceStencilID)
		ginx.Dangerous(err)

		if resourceStencil == nil {
			continue
		}

		me := c.MustGet("user").(*models.User)
		if !me.IsAdmin() {
			// check permission
			rt.bgrwCheck(c, resourceStencil.GroupID)
		}

		ginx.Dangerous(resourceStencil.Del(rt.Ctx))
	}

	ginx.NewRender(c).Message(nil)
}

func (rt *Router) resourceStencilUpsert(c *gin.Context) {
	var req models.ResourceStencil
	ginx.BindJSON(c, &req)

	username := c.MustGet("username").(string)

	req.UpdateBy = username

	var err error
	var count int64

	if req.Id == 0 {
		req.CreateBy = username

		count, err = models.GetResourceStencilCountBy(rt.Ctx, "", req.Name)
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
		req.FE2DB()

		err = req.Update(rt.Ctx, "name", "resource_stencil_cate", "resource_stencil_info", "update_at", "update_by")
	}

	Render(c, nil, err)
}

func (rt *Router) resourceStencilGet(c *gin.Context) {
	// 获取查询参数
	resourceStencilIDStr := c.Param("id")
	// 转换参数为整数
	resourceID, err := strconv.ParseInt(resourceStencilIDStr, 10, 64)
	if err != nil {
		// 处理转换错误
		c.JSON(400, gin.H{"error": "Invalid collect ID"})
		return
	}

	collect, err := models.GetResourceStencilGetsById(rt.Ctx, resourceID)
	Render(c, collect, err)
}

func (rt *Router) resourceStencilGets(c *gin.Context) {
	bgid := ginx.UrlParamStr(c, "id")
	cate := c.Query("resouce_stencil_cate")

	resource_stencil, err := models.GetResourceStencilGetsBy(rt.Ctx, bgid, cate, "")

	ginx.NewRender(c).Data(resource_stencil, err)
}
