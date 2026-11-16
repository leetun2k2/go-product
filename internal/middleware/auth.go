package middleware

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/leetun2k2/go-bedrock/logx"
	"github.com/leetun2k2/go-product/config"
)

type accessResponse struct {
	Data struct {
		UserID  string `json:"user_id"`
		IsAdmin bool   `json:"is_admin"`
	} `json:"data"`
}

// AdminAuth validates that the request carries an authenticated admin session by querying
// go-auth GET /api/admin/v1/access.
func AdminAuth(cfg *config.AuthConfig, logger *logx.Logger) gin.HandlerFunc {
	timeout := 5 * time.Second
	if cfg != nil && cfg.Timeout > 0 {
		timeout = cfg.Timeout
	}
	httpClient := &http.Client{
		Timeout: timeout,
	}

	return func(c *gin.Context) {
		ctx := c.Request.Context()

		if cfg == nil || strings.TrimSpace(cfg.URL) == "" {
			logger.Error(ctx, "admin auth rejected: auth url not configured")
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error":   "service_unavailable",
				"message": "Authentication service unavailable",
			})
			return
		}

		var token string
		if authHeader := c.GetHeader("Authorization"); authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.EqualFold(parts[0], "bearer") {
				token = strings.TrimSpace(parts[1])
			} else {
				token = strings.TrimSpace(authHeader)
			}
		} else if cookieVal, err := c.Cookie("go_auth_access"); err == nil && cookieVal != "" {
			token = cookieVal
		} else if cookieHeader := c.GetHeader("Cookie"); cookieHeader != "" {
			for _, part := range strings.Split(cookieHeader, ";") {
				part = strings.TrimSpace(part)
				if strings.HasPrefix(part, "go_auth_access=") {
					token = strings.TrimPrefix(part, "go_auth_access=")
					break
				}
			}
		}

		if token == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authentication required",
			})
			return
		}

		authURL := strings.TrimRight(cfg.URL, "/") + "/api/admin/v1/access"
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, authURL, nil)
		if err != nil {
			logger.Error(ctx, "failed to build auth access request", "error", err)
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error":   "service_unavailable",
				"message": "Authentication service unavailable",
			})
			return
		}

		req.Header.Set("Authorization", "Bearer "+token)

		if cfg.ClientID != "" {
			req.Header.Set("X-Client-Id", cfg.ClientID)
		}
		if cfg.ClientSecret != "" {
			req.Header.Set("X-Client-Secret", cfg.ClientSecret)
		}
		provider := cfg.Provider
		if provider == "" {
			provider = "google"
		}
		req.Header.Set("X-Client-Provider", provider)

		// Forward correlation id
		if correlationID := c.GetHeader("X-Correlation-Id"); correlationID != "" {
			req.Header.Set("X-Correlation-Id", correlationID)
		}

		resp, err := httpClient.Do(req)
		if err != nil {
			logger.Error(ctx, "auth service access check failed", "error", err, "url", authURL)
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error":   "service_unavailable",
				"message": "Authentication service unavailable",
			})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusUnauthorized {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authentication required",
			})
			return
		}

		if resp.StatusCode == http.StatusForbidden {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Admin role required",
			})
			return
		}

		if resp.StatusCode != http.StatusOK {
			logger.Error(ctx, "auth service returned non-200 status", "status", resp.StatusCode)
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error":   "service_unavailable",
				"message": "Authentication service unavailable",
			})
			return
		}

		var body accessResponse
		if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
			logger.Error(ctx, "failed to parse auth access response", "error", err)
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"error":   "service_unavailable",
				"message": "Authentication service unavailable",
			})
			return
		}

		if !body.Data.IsAdmin {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Admin role required",
			})
			return
		}

		c.Set("admin_user_id", body.Data.UserID)
		c.Next()
	}
}
