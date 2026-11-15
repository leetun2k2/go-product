package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/leetun2k2/go-product/internal/domain"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

func respondError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	switch {
	case errors.Is(err, domain.ErrProductNotFound),
		errors.Is(err, domain.ErrPlanNotFound),
		errors.Is(err, domain.ErrOfferNotFound),
		errors.Is(err, domain.ErrPriceNotFound),
		errors.Is(err, domain.ErrQuotaNotFound):
		c.JSON(http.StatusNotFound, ErrorResponse{Error: "not_found", Message: err.Error()})

	case errors.Is(err, domain.ErrPlanImmutable):
		c.JSON(http.StatusConflict, ErrorResponse{Error: "plan_immutable", Message: err.Error()})

	case errors.Is(err, domain.ErrConflict):
		c.JSON(http.StatusConflict, ErrorResponse{Error: "conflict", Message: err.Error()})

	case errors.Is(err, domain.ErrPlanActivationInvalid):
		c.JSON(http.StatusUnprocessableEntity, ErrorResponse{Error: "invalid_plan_activation", Message: err.Error()})

	case errors.Is(err, domain.ErrInvalidOfferType),
		errors.Is(err, domain.ErrInvalidBillingCycle),
		errors.Is(err, domain.ErrInvalidAnchorType),
		errors.Is(err, domain.ErrInvalidAmount),
		errors.Is(err, domain.ErrInvalidQuotaValues),
		errors.Is(err, domain.ErrInvalidCurrency),
		errors.Is(err, domain.ErrHierarchyMismatch):
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid_request", Message: err.Error()})

	default:
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: "internal_error", Message: err.Error()})
	}
}
