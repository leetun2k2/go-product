package transport

import (
	"github.com/gin-gonic/gin"
	"github.com/leetun2k2/go-product/internal/middleware"
)

func (app *Application) RegisterRestfulApi(engine *gin.Engine) error {
	engine.Use(middleware.CorrelationID)

	engine.GET("/health", app.HealthHandler.Health)
	engine.GET("/ready", app.HealthHandler.Ready)

	registerCatalog := func(group *gin.RouterGroup) {
		group.GET("/products/:id", app.CatalogHandler.GetCatalog)
		group.GET("/catalog", app.CatalogHandler.GetCatalog)
		group.GET("/payments/price", app.CatalogHandler.GetPaymentPrice)
		group.GET("/subscriptions/plans/:id", app.CatalogHandler.GetSubscriptionPlan)
		group.GET("/usage/plans/:id/quotas", app.CatalogHandler.GetUsageQuotas)
	}

	registerAdmin := func(group *gin.RouterGroup) {
		// Products
		group.POST("/products", app.AdminHandler.CreateProduct)
		group.GET("/products", app.AdminHandler.ListProducts)
		group.GET("/products/:id", app.AdminHandler.GetProduct)
		group.PATCH("/products/:id", app.AdminHandler.UpdateProduct)

		// Plans under Product
		group.POST("/products/:id/plans", app.AdminHandler.CreatePlan)
		group.GET("/products/:id/plans", app.AdminHandler.ListPlansByProduct)

		// Plans direct
		group.GET("/plans/:id", app.AdminHandler.GetPlan)
		group.PATCH("/plans/:id", app.AdminHandler.UpdatePlan)
		group.POST("/plans/:id/activate", app.AdminHandler.ActivatePlan)
		group.POST("/plans/:id/deactivate", app.AdminHandler.DeactivatePlan)

		// Offers under Plan
		group.POST("/plans/:id/offers", app.AdminHandler.CreateOffer)
		group.GET("/plans/:id/offers", app.AdminHandler.ListOffersByPlan)

		// Quotas under Plan
		group.POST("/plans/:id/quotas", app.AdminHandler.CreateQuotaRule)
		group.GET("/plans/:id/quotas", app.AdminHandler.ListQuotaRulesByPlan)

		// Offers direct
		group.GET("/offers/:id", app.AdminHandler.GetOffer)
		group.PATCH("/offers/:id", app.AdminHandler.UpdateOffer)
		group.POST("/offers/:id/activate", app.AdminHandler.ActivateOffer)
		group.POST("/offers/:id/deactivate", app.AdminHandler.DeactivateOffer)

		// Prices under Offer
		group.POST("/offers/:id/prices", app.AdminHandler.UpsertPrice)
		group.GET("/offers/:id/prices", app.AdminHandler.ListPricesByOffer)

		// Prices direct
		group.GET("/prices/:id", app.AdminHandler.GetPrice)
		group.PATCH("/prices/:id", app.AdminHandler.UpdatePriceStatus)

		// Quotas direct
		group.GET("/quotas/:id", app.AdminHandler.GetQuotaRule)
		group.PATCH("/quotas/:id", app.AdminHandler.UpdateQuotaRule)
		group.DELETE("/quotas/:id", app.AdminHandler.DeleteQuotaRule)
	}

	registerCatalog(engine.Group("/api/catalog/v1"))
	registerCatalog(engine.Group("/api/product/v1"))

	registerAdmin(engine.Group("/api/admin/v1"))
	registerAdmin(engine.Group("/api/product/v1/admin"))
	registerAdmin(engine.Group("/api/catalog/v1/admin"))

	return nil
}
