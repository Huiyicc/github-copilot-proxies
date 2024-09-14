package router

import (
	"github.com/gin-gonic/gin"
	"html/template"
	authApi "hzer/internal/controller/auth"
	"hzer/internal/controller/copilot"
	"hzer/internal/middleware"
	"hzer/static"
)

func NewHTTPRouter(r *gin.Engine) {
	//isDebug := os.Getenv("GIN_MODE") == "debug"
	rootRouter := r.Group("/")
	tmpl := template.Must(template.New("").ParseFS(static.Public, "public/*.html"))
	r.SetHTMLTemplate(tmpl)

	apiRouter := r.Group("/api")

	rootRouter.Use(middleware.Cors())
	apiRouter.Use(middleware.Cors())

	authApi.GinApi(rootRouter)
	copilot.GinApi(rootRouter)

}
