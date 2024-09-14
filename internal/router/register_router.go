package router

import (
	"github.com/gin-gonic/gin"
	"html/template"
	authApi "ripper/internal/controller/auth"
	"ripper/internal/controller/copilot"
	"ripper/internal/middleware"
	"ripper/static"
)

func NewHTTPRouter(r *gin.Engine) {
	rootRouter := r.Group("/")
	tmpl := template.Must(template.New("").ParseFS(static.Public, "public/*.html"))
	r.SetHTMLTemplate(tmpl)

	apiRouter := r.Group("/api")

	rootRouter.Use(middleware.Cors())
	apiRouter.Use(middleware.Cors())

	authApi.GinApi(rootRouter)
	copilot.GinApi(rootRouter)

}
