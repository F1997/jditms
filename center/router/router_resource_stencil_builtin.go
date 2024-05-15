package router

import (
	"jditms/models"

	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/ginx"
)

func (rt *Router) resourceStencilBuiltInGets(c *gin.Context) {

	name := c.Query("resouce_stencil_cate")

	// if name != "" {
	// 	resources, err := models.GetResourceStencilBuiltInGetsBy(rt.Ctx, "")
	// 	ginx.NewRender(c).Data(resources, err)
	// } else {
	// 	resources, err := models.GetResourceStencilBuiltInGetsBy(rt.Ctx, "")
	// 	ginx.NewRender(c).Data(resources, err)
	// }
	resources, err := models.GetResourceStencilBuiltInGetsBy(rt.Ctx, name)
	// log.Println(resources)
	ginx.NewRender(c).Data(resources, err)
}
