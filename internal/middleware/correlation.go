package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/leetun2k2/go-bedrock/contextx"
)

const (
	headerCorrelationID = "X-Correlation-Id"
	headerRequestID     = "X-Request-Id"
)

// CorrelationID resolves a per-request correlation ID from the inbound
// X-Correlation-Id or X-Request-Id header, generating one when absent, and
// attaches it to the request context so structured auth event logs can
// carry it (spec 11).
func CorrelationID(c *gin.Context) {
	id := strings.TrimSpace(c.GetHeader(headerCorrelationID))
	if id == "" {
		id = strings.TrimSpace(c.GetHeader(headerRequestID))
	}
	if id == "" {
		if generated, err := uuid.NewV7(); err == nil {
			id = generated.String()
		}
	}

	c.Request = c.Request.WithContext(contextx.SetCorrelationID(c.Request.Context(), id))
	c.Writer.Header().Set(headerCorrelationID, id)
	c.Next()
}
