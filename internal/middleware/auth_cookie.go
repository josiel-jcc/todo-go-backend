package middleware

import (
	"net/http"
	"time"
	"todo-go-backend/internal/config"
	"todo-go-backend/internal/database"
	"todo-go-backend/internal/models"
	"todo-go-backend/pkg/utils"

	"github.com/gin-gonic/gin"
)

// SetAuthCookie sets the HttpOnly session cookie.
func SetAuthCookie(c *gin.Context, cfg *config.Config, token string) {
	maxAge := int(utils.TokenMaxAge.Seconds())
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(authCookieName, token, maxAge, "/", cfg.CookieDomain, cfg.CookieSecure, true)
}

// ClearAuthCookie removes the session cookie.
func ClearAuthCookie(c *gin.Context, cfg *config.Config) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(authCookieName, "", -1, "/", cfg.CookieDomain, cfg.CookieSecure, true)
}

// RevokeTokenJTI adds a JWT ID to the denylist until its expiration.
func RevokeTokenJTI(jti string, expiresAt time.Time) error {
	if jti == "" {
		return nil
	}
	entry := models.TokenDenylist{JTI: jti, ExpiresAt: expiresAt}
	return database.DB.Create(&entry).Error
}
