package router

import (
	"net/http"

	"github.com/stpnv0/WarehouseControl/internal/domain"
	"github.com/stpnv0/WarehouseControl/internal/middleware"
	"github.com/wb-go/wbf/ginext"
)

type AuthHandler interface {
	Login(c *ginext.Context)
	ListUsers(c *ginext.Context)
}

type AuditHandler interface {
	GetByItemID(c *ginext.Context)
	List(c *ginext.Context)
	ExportCSV(c *ginext.Context)
}
type ItemHandler interface {
	Create(c *ginext.Context)
	List(c *ginext.Context)
	GetByID(c *ginext.Context)
	Update(c *ginext.Context)
	Delete(c *ginext.Context)
}

type TokenValidator interface {
	Validate(tokenStr string) (*domain.AuthClaims, error)
}

func InitRouter(
	mode string,
	authHandler AuthHandler,
	auditHandler AuditHandler,
	itemHandler ItemHandler,
	tokenValidator TokenValidator,
	mw ...ginext.HandlerFunc,
) *ginext.Engine {
	router := ginext.New(mode)
	router.Use(ginext.Recovery())
	router.Use(mw...)

	auth := router.Group("/api/auth")
	{
		auth.POST("/login", authHandler.Login)
		auth.GET("/users", authHandler.ListUsers)
	}

	api := router.Group("/api")
	api.Use(middleware.Auth(tokenValidator))
	{
		items := api.Group("/items")
		{
			items.GET("", itemHandler.List)
			items.POST("", itemHandler.Create)
			items.GET("/:id", itemHandler.GetByID)
			items.PUT("/:id", itemHandler.Update)
			items.DELETE("/:id", itemHandler.Delete)

			items.GET("/:id/audit", auditHandler.GetByItemID)
		}

		audit := api.Group("/audit")
		{
			audit.GET("", auditHandler.List)
			audit.GET("/export", auditHandler.ExportCSV)
		}
	}

	router.GET("/health", func(c *ginext.Context) {
		c.JSON(http.StatusOK, ginext.H{"status": "ok"})
	})

	router.LoadHTMLGlob("web/templates/*")
	router.Static("/static", "web/static")

	router.GET("/", func(c *ginext.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	return router
}
