package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/leetun2k2/go-product/internal/domain"
	"github.com/leetun2k2/go-product/internal/service"
)

type AdminHandler struct {
	adminService *service.AdminService
}

func NewAdminHandler(adminService *service.AdminService) *AdminHandler {
	return &AdminHandler{adminService: adminService}
}

func getIDParam(c *gin.Context, fallbackName string) (uuid.UUID, error) {
	val := strings.TrimSpace(c.Param("id"))
	if val == "" && fallbackName != "" {
		val = strings.TrimSpace(c.Param(fallbackName))
	}
	return uuid.Parse(val)
}

// --- Product Requests & Handlers ---

type CreateProductRequest struct {
	Code        string  `json:"code" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
}

func (h *AdminHandler) CreateProduct(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: err.Error()})
		return
	}

	product, err := h.adminService.CreateProduct(c.Request.Context(), req.Code, req.Name, req.Description)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, product)
}

type UpdateProductRequest struct {
	Name        string                `json:"name"`
	Description *string               `json:"description"`
	Status      *domain.ProductStatus `json:"status"`
}

func (h *AdminHandler) UpdateProduct(c *gin.Context) {
	id, err := getIDParam(c, "product_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "invalid product id"})
		return
	}

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: err.Error()})
		return
	}

	product, err := h.adminService.UpdateProduct(c.Request.Context(), id, req.Name, req.Description, req.Status)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, product)
}

func (h *AdminHandler) GetProduct(c *gin.Context) {
	id, err := getIDParam(c, "product_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "invalid product id"})
		return
	}

	product, err := h.adminService.GetProduct(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, product)
}

func (h *AdminHandler) ListProducts(c *gin.Context) {
	products, err := h.adminService.ListProducts(c.Request.Context())
	if err != nil {
		respondError(c, err)
		return
	}
	if products == nil {
		products = []domain.Product{}
	}
	c.JSON(http.StatusOK, gin.H{"products": products})
}

// --- Plan Requests & Handlers ---

type CreatePlanRequest struct {
	Code        string  `json:"code" binding:"required"`
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description"`
	Version     int     `json:"version"`
}

func (h *AdminHandler) CreatePlan(c *gin.Context) {
	productID, err := getIDParam(c, "product_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "invalid product_id"})
		return
	}

	var req CreatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: err.Error()})
		return
	}

	plan, err := h.adminService.CreatePlan(c.Request.Context(), productID, req.Code, req.Name, req.Description, req.Version)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, plan)
}

type UpdatePlanRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

func (h *AdminHandler) UpdatePlan(c *gin.Context) {
	id, err := getIDParam(c, "plan_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "invalid plan id"})
		return
	}

	var req UpdatePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: err.Error()})
		return
	}

	plan, err := h.adminService.UpdatePlan(c.Request.Context(), id, req.Name, req.Description)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, plan)
}

func (h *AdminHandler) ActivatePlan(c *gin.Context) {
	id, err := getIDParam(c, "plan_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "invalid plan id"})
		return
	}

	plan, err := h.adminService.ActivatePlan(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, plan)
}

func (h *AdminHandler) DeactivatePlan(c *gin.Context) {
	id, err := getIDParam(c, "plan_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "invalid plan id"})
		return
	}

	plan, err := h.adminService.DeactivatePlan(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, plan)
}

func (h *AdminHandler) GetPlan(c *gin.Context) {
	id, err := getIDParam(c, "plan_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "invalid plan id"})
		return
	}

	plan, err := h.adminService.GetPlan(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, plan)
}

func (h *AdminHandler) ListPlansByProduct(c *gin.Context) {
	productID, err := getIDParam(c, "product_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "invalid product_id"})
		return
	}

	plans, err := h.adminService.ListPlansByProductID(c.Request.Context(), productID)
	if err != nil {
		respondError(c, err)
		return
	}
	if plans == nil {
		plans = []domain.Plan{}
	}
	c.JSON(http.StatusOK, gin.H{"plans": plans})
}

// --- Offer Requests & Handlers ---

type CreateOfferRequest struct {
	Code         string              `json:"code" binding:"required"`
	OfferType    domain.OfferType    `json:"offer_type" binding:"required"`
	BillingCycle domain.BillingCycle `json:"billing_cycle" binding:"required"`
	SortOrder    int                 `json:"sort_order"`
}

func (h *AdminHandler) CreateOffer(c *gin.Context) {
	planID, err := getIDParam(c, "plan_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "invalid plan_id"})
		return
	}

	var req CreateOfferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: err.Error()})
		return
	}

	offer, err := h.adminService.CreateOffer(c.Request.Context(), planID, req.Code, req.OfferType, req.BillingCycle, req.SortOrder)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, offer)
}

type UpdateOfferRequest struct {
	Code      string `json:"code"`
	SortOrder int    `json:"sort_order"`
}

func (h *AdminHandler) UpdateOffer(c *gin.Context) {
	id, err := getIDParam(c, "offer_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "invalid offer id"})
		return
	}

	var req UpdateOfferRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: err.Error()})
		return
	}

	offer, err := h.adminService.UpdateOffer(c.Request.Context(), id, req.Code, req.SortOrder)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, offer)
}

func (h *AdminHandler) ActivateOffer(c *gin.Context) {
	id, err := getIDParam(c, "offer_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "invalid offer id"})
		return
	}

	offer, err := h.adminService.ActivateOffer(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, offer)
}

func (h *AdminHandler) DeactivateOffer(c *gin.Context) {
	id, err := getIDParam(c, "offer_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "invalid offer id"})
		return
	}

	offer, err := h.adminService.DeactivateOffer(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, offer)
}

func (h *AdminHandler) GetOffer(c *gin.Context) {
	id, err := getIDParam(c, "offer_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "invalid offer id"})
		return
	}

	offer, err := h.adminService.GetOffer(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, offer)
}

func (h *AdminHandler) ListOffersByPlan(c *gin.Context) {
	planID, err := getIDParam(c, "plan_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "invalid plan_id"})
		return
	}

	offers, err := h.adminService.ListOffersByPlanID(c.Request.Context(), planID)
	if err != nil {
		respondError(c, err)
		return
	}
	if offers == nil {
		offers = []domain.Offer{}
	}
	c.JSON(http.StatusOK, gin.H{"offers": offers})
}

// --- Price Requests & Handlers ---

type UpsertPriceRequest struct {
	Currency string             `json:"currency" binding:"required"`
	Amount   int64              `json:"amount"`
	Status   domain.PriceStatus `json:"status"`
}

func (h *AdminHandler) UpsertPrice(c *gin.Context) {
	offerID, err := getIDParam(c, "offer_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "invalid offer_id"})
		return
	}

	var req UpsertPriceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: err.Error()})
		return
	}

	price, err := h.adminService.UpsertPrice(c.Request.Context(), offerID, req.Currency, req.Amount, req.Status)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, price)
}

type UpdatePriceStatusRequest struct {
	Status domain.PriceStatus `json:"status" binding:"required"`
}

func (h *AdminHandler) UpdatePriceStatus(c *gin.Context) {
	id, err := getIDParam(c, "price_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "invalid price id"})
		return
	}

	var req UpdatePriceStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: err.Error()})
		return
	}

	price, err := h.adminService.UpdatePriceStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, price)
}

func (h *AdminHandler) GetPrice(c *gin.Context) {
	id, err := getIDParam(c, "price_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "invalid price id"})
		return
	}

	price, err := h.adminService.GetPrice(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, price)
}

func (h *AdminHandler) ListPricesByOffer(c *gin.Context) {
	offerID, err := getIDParam(c, "offer_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "invalid offer_id"})
		return
	}

	prices, err := h.adminService.ListPricesByOfferID(c.Request.Context(), offerID)
	if err != nil {
		respondError(c, err)
		return
	}
	if prices == nil {
		prices = []domain.Price{}
	}
	c.JSON(http.StatusOK, gin.H{"prices": prices})
}

// --- Quota Rule Requests & Handlers ---

type CreateQuotaRuleRequest struct {
	RuleKey               string            `json:"rule_key" binding:"required"`
	LimitAmount           int64             `json:"limit_amount" binding:"required"`
	WindowDurationSeconds int               `json:"window_duration_seconds" binding:"required"`
	AnchorType            domain.AnchorType `json:"anchor_type" binding:"required"`
}

func (h *AdminHandler) CreateQuotaRule(c *gin.Context) {
	planID, err := getIDParam(c, "plan_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "invalid plan_id"})
		return
	}

	var req CreateQuotaRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: err.Error()})
		return
	}

	quota, err := h.adminService.CreateQuotaRule(c.Request.Context(), planID, req.RuleKey, req.LimitAmount, req.WindowDurationSeconds, req.AnchorType)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusCreated, quota)
}

type UpdateQuotaRuleRequest struct {
	LimitAmount           int64 `json:"limit_amount" binding:"required"`
	WindowDurationSeconds int   `json:"window_duration_seconds" binding:"required"`
}

func (h *AdminHandler) UpdateQuotaRule(c *gin.Context) {
	id, err := getIDParam(c, "quota_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "invalid quota rule id"})
		return
	}

	var req UpdateQuotaRuleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: err.Error()})
		return
	}

	quota, err := h.adminService.UpdateQuotaRule(c.Request.Context(), id, req.LimitAmount, req.WindowDurationSeconds)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, quota)
}

func (h *AdminHandler) DeleteQuotaRule(c *gin.Context) {
	id, err := getIDParam(c, "quota_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "invalid quota rule id"})
		return
	}

	if err := h.adminService.DeleteQuotaRule(c.Request.Context(), id); err != nil {
		respondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *AdminHandler) GetQuotaRule(c *gin.Context) {
	id, err := getIDParam(c, "quota_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "invalid quota rule id"})
		return
	}

	quota, err := h.adminService.GetQuotaRule(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}
	c.JSON(http.StatusOK, quota)
}

func (h *AdminHandler) ListQuotaRulesByPlan(c *gin.Context) {
	planID, err := getIDParam(c, "plan_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: "invalid plan_id"})
		return
	}

	quotas, err := h.adminService.ListQuotaRulesByPlanID(c.Request.Context(), planID)
	if err != nil {
		respondError(c, err)
		return
	}
	if quotas == nil {
		quotas = []domain.QuotaRule{}
	}
	c.JSON(http.StatusOK, gin.H{"quotas": quotas})
}
