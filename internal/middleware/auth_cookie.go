package middleware

import (
	"net/http"
	"strings"
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
	sameSite := parseSameSite(cfg.CookieSameSite)
	secure := cfg.CookieSecure || sameSite == http.SameSiteNoneMode
	c.SetSameSite(sameSite)
	c.SetCookie(authCookieName, token, maxAge, "/", cfg.CookieDomain, secure, true)
}

// ClearAuthCookie removes the session cookie.
func ClearAuthCookie(c *gin.Context, cfg *config.Config) {
	sameSite := parseSameSite(cfg.CookieSameSite)
	secure := cfg.CookieSecure || sameSite == http.SameSiteNoneMode
	c.SetSameSite(sameSite)
	c.SetCookie(authCookieName, "", -1, "/", cfg.CookieDomain, secure, true)
}

func parseSameSite(value string) http.SameSite {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "none":
		return http.SameSiteNoneMode
	case "strict":
		return http.SameSiteStrictMode
	default:
		return http.SameSiteLaxMode
	}
}

// RevokeTokenJTI adds a JWT ID to the denylist until its expiration.
func RevokeTokenJTI(jti string, expiresAt time.Time) error {
	if jti == "" {
		return nil
	}
	entry := models.TokenDenylist{JTI: jti, ExpiresAt: expiresAt}
	return database.DB.Create(&entry).Error
}
