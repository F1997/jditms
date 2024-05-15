package router

import (
	"jditms/models"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/ginx"
	"github.com/toolkits/pkg/str"
)

type resourceForm struct {
	Name  string `json:"name" binding:"required"`  // Name 是字符串类型，不为空
	Ident string `json:"ident" binding:"required"` // Ident 是字符串类型，不为空
	// ResourceQueryIndicators string `json:"resource_query_indicators"` // ResourceQueryIndicators 是字符串类型
	// ResourceIP        string                   `json:"resource_ip" binding:"required"` // Name 是字符串类型，不为空
	ResourceAlertRule []string                 `json:"resource_alert_rule"`
	ResourceInfo      []map[string]interface{} `json:"resource_info"`
	ResourceDashboard string                   `json:"resource_dashboard"`
	ResourceCate      string                   `json:"resource_cate" binding:"required"`
	ResourceStencil   string                   `json:"resource_stencil"`
	// GroupID      int64           `json:"group_id"`  // GroupID 是 bigint 类型，不为空
}

func (rt *Router) resourceAdd(c *gin.Context) {
	var f resourceForm
	ginx.BindJSON(c, &f)

	username := Username(c)
	resource := &models.Resource{
		Name:                  f.Name,
		Ident:                 f.Ident,
		ResourceAlertRule:     "",
		ResourceAlertRuleJson: f.ResourceAlertRule,
		ResourceDashboard:     f.ResourceDashboard,
		ResourceInfo:          "",
		ResourceInfoJson:      f.ResourceInfo,
		ResourceCate:          f.ResourceCate,
		ResourceStencil:       f.ResourceStencil,
		GroupID:               ginx.UrlParamInt64(c, "id"),
		CreateBy:              username,
		UpdateBy:              username,
	}

	err := resource.FE2DB()
	if err != nil {
		// logger.Error("json to string error")
		Dangerous(c, err)
		return
	}

	err = resource.Add(rt.Ctx)
	ginx.Dangerous(err)

	ginx.NewRender(c).Data(resource, nil)
}

func (rt *Router) resourceUpsert(c *gin.Context) {
	var req models.Resource
	ginx.BindJSON(c, &req)

	username := c.MustGet("username").(string)

	req.UpdateBy = username

	var err error
	var count int64

	if req.Id == 0 {
		req.CreateBy = username

		count, err = models.GetResourceCountBy(rt.Ctx, "", req.Name)
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

		err = req.Update(rt.Ctx, "name", "resource_cate", "resource_stencil", "ident", "resource_alert_rule", "resource_info", "resource_dashboard", "update_at", "update_by")
	}

	Render(c, nil, err)
}

func (rt *Router) resourceGet(c *gin.Context) {
	// 获取查询参数
	resourceIDStr := c.Param("id")
	// 转换参数为整数
	resourceID, err := strconv.ParseInt(resourceIDStr, 10, 64)
	if err != nil {
		// 处理转换错误
		c.JSON(400, gin.H{"error": "Invalid collect ID"})
		return
	}

	collect, err := models.GetResourceGetsById(rt.Ctx, resourceID)
	Render(c, collect, err)
}

// func (rt *Router) resourceGets(c *gin.Context) {
// 	bgid := ginx.UrlParamStr(c, "id")

// 	resources, err := models.GetResourceGetsBy(rt.Ctx, bgid, "", "", "")
// 	// log.Println(resources)
// 	ginx.NewRender(c).Data(resources, err)
// }

func (rt *Router) resourceGets(c *gin.Context) {
	bgids := str.IdsInt64(ginx.QueryStr(c, "gids", ""), ",")
	cate := ginx.QueryStr(c, "cate", "")
	query := ginx.QueryStr(c, "query", "")
	limit := ginx.QueryInt(c, "limit", 30)

	user := c.MustGet("user").(*models.User)
	if !user.IsAdmin() {
		// 如果是非 admin 用户，全部对象的情况，找到用户有权限的业务组
		var err error
		bgids, err = models.MyBusiGroupIds(rt.Ctx, user.Id)
		ginx.Dangerous(err)

		// 将未分配业务组的对象也加入到列表中
		bgids = append(bgids, 0)
	}

	var err error
	log.Println(bgids)

	total, err := models.ResourceTotal(rt.Ctx, bgids, cate, query)
	ginx.Dangerous(err)

	list, err := models.ResourceGets(rt.Ctx, bgids, cate, query, limit, ginx.Offset(c, limit))
	ginx.Dangerous(err)

	ginx.NewRender(c).Data(gin.H{
		"list":  list,
		"total": total,
	}, nil)
}

// func (rt *Router) resourceGetByCates(c *gin.Context) {
// 	bgid := ginx.UrlParamStr(c, "id")
// 	resource_cate := ginx.UrlParamStr(c, "cate")
// 	// log.Println(resource_cate)
// 	resources, err := models.GetResourceGetsBy(rt.Ctx, bgid, resource_cate, "", "")
// 	// log.Println(resources)
// 	ginx.NewRender(c).Data(resources, err)
// }

func (rt *Router) checkResourcePerm(c *gin.Context, ids []int64) {
	user := c.MustGet("user").(*models.User)
	nopri, err := user.NopriResourceID(rt.Ctx, ids)
	ginx.Dangerous(err)

	if len(nopri) > 0 {
		var result string
		for i, num := range nopri {
			if i > 0 {
				result += ","
			}
			result += strconv.FormatInt(num, 10)
		}
		ginx.Bomb(http.StatusForbidden, "No permission to operate the resources: %s", result)
	}
}

type resourceBgidForm struct {
	Ids  []int64 `json:"ids" binding:"required"`
	Bgid int64   `json:"bgid"`
}

func (rt *Router) resourceUpdateBgid(c *gin.Context) {
	var f resourceBgidForm
	ginx.BindJSON(c, &f)

	if len(f.Ids) == 0 {
		ginx.Bomb(http.StatusBadRequest, "ids empty")
	}

	user := c.MustGet("user").(*models.User)
	if user.IsAdmin() {
		ginx.NewRender(c).Message(models.ResourceUpdateBgid(rt.Ctx, f.Ids, f.Bgid))
		return
	}

	if f.Bgid > 0 {
		// 把要操作的机器分成两部分，一部分是bgid为0，需要管理员分配，另一部分bgid>0，说明是业务组内部想调整
		// 比如原来分配给didiyun的机器，didiyun的管理员想把部分机器调整到didiyun-ceph下
		// 对于调整的这种情况，当前登录用户要对这批机器有操作权限，同时还要对目标BG有操作权限
		orphans, err := models.ResourceIdsFilter(rt.Ctx, f.Ids, "group_id = ?", 0)
		ginx.Dangerous(err)

		// 机器里边存在未归组的，登录用户就需要是admin
		if len(orphans) > 0 && !user.IsAdmin() {
			can, err := user.CheckPerm(rt.Ctx, "/resource/bind")
			ginx.Dangerous(err)
			if !can {
				ginx.Bomb(http.StatusForbidden, "No permission. Only admin can assign BG")
			}
		}

		reBelongs, err := models.IdsFilter(rt.Ctx, f.Ids, "group_id > ?", 0)
		ginx.Dangerous(err)

		if len(reBelongs) > 0 {
			// 对于这些要重新分配的机器，操作者要对这些机器本身有权限，同时要对目标bgid有权限
			rt.checkResourcePerm(c, f.Ids)

			bg := BusiGroup(rt.Ctx, f.Bgid)
			can, err := user.CanDoBusiGroup(rt.Ctx, bg, "rw")
			ginx.Dangerous(err)

			if !can {
				ginx.Bomb(http.StatusForbidden, "No permission. You are not admin of BG(%s)", bg.Name)
			}
		}
	} else if f.Bgid == 0 {
		// 退还机器
		rt.checkResourcePerm(c, f.Ids)
	} else {
		ginx.Bomb(http.StatusBadRequest, "invalid bgid")
	}

	ginx.NewRender(c).Message(models.ResourceUpdateBgid(rt.Ctx, f.Ids, f.Bgid))
}

type resourceIdsForm struct {
	Ids []int64 `json:"ids" binding:"required"`
}

func (rt *Router) resourceDel(c *gin.Context) {
	var f resourceIdsForm
	ginx.BindJSON(c, &f)

	if len(f.Ids) == 0 {
		ginx.Bomb(http.StatusBadRequest, "idents empty")
	}

	rt.checkResourcePerm(c, f.Ids)

	ginx.NewRender(c).Message(models.ResourceDel(rt.Ctx, f.Ids))
}
