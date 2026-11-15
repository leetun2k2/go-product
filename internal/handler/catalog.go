package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/leetun2k2/go-product/internal/domain"
	"github.com/leetun2k2/go-product/internal/service"
)

type CatalogHandler struct {
	catalogService *service.CatalogService
	adminService   *service.AdminService
}

func NewCatalogHandler(catalogService *service.CatalogService, adminService *service.AdminService) *CatalogHandler {
	return &CatalogHandler{
		catalogService: catalogService,
		adminService:   adminService,
	}
}

// GetCatalog handles GET /api/catalog/v1/products/:id?currency=USD
// or GET /api/catalog/v1/catalog?product_id=...&currency=USD
func (h *CatalogHandler) GetCatalog(c *gin.Context) {
	ctx := c.Request.Context()

	productParam := strings.TrimSpace(c.Param("id"))
	if productParam == "" {
		productParam = strings.TrimSpace(c.Param("product_id"))
	}
	if productParam == "" {
		productParam = strings.TrimSpace(c.Query("product_id"))
	}
	currency := strings.TrimSpace(c.Query("currency"))

	if productParam == "" || currency == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "product_id and currency query params are required"})
		return
	}

	productID, err := uuid.Parse(productParam)
	if err != nil {
		// Attempt resolution by product code
		prod, prodErr := h.adminService.GetProductByCode(ctx, productParam)
		if prodErr != nil {
			respondError(c, prodErr)
			return
		}
		productID = prod.ID
	}

	items, err := h.catalogService.GetCatalog(ctx, productID, currency)
	if err != nil {
		respondError(c, err)
		return
	}

	if items == nil {
		items = []domain.CatalogItem{}
	}

	c.JSON(http.StatusOK, gin.H{
		"product_id": productID,
		"currency":   strings.ToUpper(currency),
		"items":      items,
	})
}

// GetPaymentPrice handles GET /api/catalog/v1/payments/price?offer_id=...&currency=...
func (h *CatalogHandler) GetPaymentPrice(c *gin.Context) {
	ctx := c.Request.Context()

	offerIDStr := strings.TrimSpace(c.Query("offer_id"))
	currency := strings.TrimSpace(c.Query("currency"))

	if offerIDStr == "" || currency == "" {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "offer_id and currency query params are required"})
		return
	}

	offerID, err := uuid.Parse(offerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "invalid offer_id format"})
		return
	}

	price, err := h.catalogService.GetPaymentPrice(ctx, offerID, currency)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"price_id":   price.ID,
		"offer_id":   price.OfferID,
		"product_id": price.ProductID,
		"plan_id":    price.PlanID,
		"currency":   price.Currency,
		"amount":     price.Amount,
		"status":     price.Status,
	})
}

// GetSubscriptionPlan handles GET /api/catalog/v1/subscriptions/plans/:id
func (h *CatalogHandler) GetSubscriptionPlan(c *gin.Context) {
	ctx := c.Request.Context()

	planIDStr := strings.TrimSpace(c.Param("id"))
	if planIDStr == "" {
		planIDStr = strings.TrimSpace(c.Param("plan_id"))
	}
	planID, err := uuid.Parse(planIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "invalid plan_id format"})
		return
	}

	plan, quotas, err := h.catalogService.GetSubscriptionPlan(ctx, planID)
	if err != nil {
		respondError(c, err)
		return
	}

	if quotas == nil {
		quotas = []domain.QuotaRule{}
	}

	c.JSON(http.StatusOK, gin.H{
		"plan":   plan,
		"quotas": quotas,
	})
}

// GetUsageQuotas handles GET /api/catalog/v1/usage/plans/:id/quotas
func (h *CatalogHandler) GetUsageQuotas(c *gin.Context) {
	ctx := c.Request.Context()

	planIDStr := strings.TrimSpace(c.Param("id"))
	if planIDStr == "" {
		planIDStr = strings.TrimSpace(c.Param("plan_id"))
	}
	planID, err := uuid.Parse(planIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "invalid plan_id format"})
		return
	}

	quotas, err := h.catalogService.GetUsageQuotas(ctx, planID)
	if err != nil {
		respondError(c, err)
		return
	}

	if quotas == nil {
		quotas = []domain.QuotaRule{}
	}

	c.JSON(http.StatusOK, gin.H{
		"plan_id": planID,
		"quotas":  quotas,
	})
}
