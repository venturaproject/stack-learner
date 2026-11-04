package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	appauth "boilerplate-learner/internal/application/auth"
	"boilerplate-learner/internal/config"
	domainauth "boilerplate-learner/internal/domain/auth"
	"boilerplate-learner/internal/infrastructure/http/httpx"
	"boilerplate-learner/internal/infrastructure/http/middleware"
)

type AuthHandler struct {
	svc     *appauth.AuthService
	jwtSvc  *appauth.JWTService
	revoked domainauth.RevokedTokenRepository
	cfg     *config.Config
}

func NewAuthHandler(svc *appauth.AuthService, jwtSvc *appauth.JWTService, revoked domainauth.RevokedTokenRepository, cfg *config.Config) *AuthHandler {
	return &AuthHandler{svc: svc, jwtSvc: jwtSvc, revoked: revoked, cfg: cfg}
}

// Login godoc
// @Summary Login
// @Tags auth
// @Accept json
// @Produce json
// @Param body body appauth.LoginRequest true "Credenciales"
// @Success 200 {object} map[string]interface{}
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req appauth.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	details, tokens, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	csrf := newCSRFToken()
	h.setAccessCookie(c, tokens.AccessToken)
	h.setRefreshCookie(c, tokens.RefreshJTI, tokens.RefreshExpiresAt)
	h.setCsrfCookie(c, csrf)

	c.JSON(http.StatusOK, gin.H{"user": formatUser(details)})
}

// Logout godoc
// @Summary Logout
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /auth/logout [post]
func (h *AuthHandler) Logout(c *gin.Context) {
	if accessToken, err := c.Cookie("access_token"); err == nil && accessToken != "" {
		if claims, err := h.jwtSvc.ParseAccessToken(accessToken); err == nil {
			_ = h.revoked.Revoke(c.Request.Context(), claims.ID, claims.ExpiresAt.Time)
		}
	}

	refreshJTI, _ := c.Cookie("refresh_jti")
	if refreshJTI != "" {
		_ = h.svc.Logout(c.Request.Context(), refreshJTI)
	}

	h.clearCookies(c)
	c.JSON(http.StatusOK, gin.H{"message": "Sesión cerrada"})
}

// Refresh godoc
// @Summary Refresh tokens
// @Tags auth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /auth/refresh [post]
func (h *AuthHandler) Refresh(c *gin.Context) {
	refreshJTI, err := c.Cookie("refresh_jti")
	if err != nil || refreshJTI == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No refresh token"})
		return
	}

	tokens, err := h.svc.Refresh(c.Request.Context(), refreshJTI)
	if err != nil {
		h.clearCookies(c)
		httpx.Respond(c, err)
		return
	}

	csrf := newCSRFToken()
	h.setAccessCookie(c, tokens.AccessToken)
	h.setRefreshCookie(c, tokens.RefreshJTI, tokens.RefreshExpiresAt)
	h.setCsrfCookie(c, csrf)

	c.JSON(http.StatusOK, gin.H{"message": "Token refrescado"})
}

// GetMe godoc
// @Summary Get current user
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /auth/me [get]
func (h *AuthHandler) GetMe(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No autenticado"})
		return
	}

	details, err := h.svc.GetMe(c.Request.Context(), userID)
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": formatUser(details)})
}

// UpdateMe godoc
// @Summary Update current user profile
// @Tags auth
// @Security BearerAuth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /auth/me [patch]
func (h *AuthHandler) UpdateMe(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No autenticado"})
		return
	}

	var req appauth.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	details, err := h.svc.UpdateMe(c.Request.Context(), userID, req)
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": formatUser(details)})
}

// UploadAvatar godoc
// @Summary Upload avatar
// @Tags auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /auth/me/avatar [post]
func (h *AuthHandler) UploadAvatar(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No autenticado"})
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No se recibió archivo"})
		return
	}

	src, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No se pudo leer el archivo"})
		return
	}
	defer src.Close()

	head := make([]byte, 512)
	n, _ := io.ReadFull(src, head)
	contentType := http.DetectContentType(head[:n])
	if !isAllowedAvatarContentType(contentType) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Formato de avatar no permitido"})
		return
	}

	avatarsDir := filepath.Join(h.cfg.Media.PrivateDir, "avatars")
	if err := os.MkdirAll(avatarsDir, 0755); err != nil { //nolint:gosec // 0755 needed for nginx traversal
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creando directorio"})
		return
	}
	_ = os.Chmod(avatarsDir, 0755) //nolint:gosec // 0755 needed for service traversal

	filename, err := newAvatarFilename(userID, file.Filename)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error generando nombre de archivo"})
		return
	}

	if err := c.SaveUploadedFile(file, filepath.Join(avatarsDir, filename)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error guardando archivo"})
		return
	}

	details, err := h.svc.UploadAvatar(c.Request.Context(), userID, filename)
	if err != nil {
		httpx.Respond(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": formatUser(details)})
}

func (h *AuthHandler) GetMyAvatar(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No autenticado"})
		return
	}

	h.serveAvatarForUser(c, userID)
}

func (h *AuthHandler) GetUserAvatar(c *gin.Context) {
	userID, err := pathInt32(c, "id")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	h.serveAvatarForUser(c, userID)
}

func (h *AuthHandler) serveAvatarForUser(c *gin.Context, userID int32) {
	user, err := h.svc.GetMe(c.Request.Context(), userID)
	if err != nil {
		httpx.Respond(c, err)
		return
	}
	if user.Avatar == nil || strings.TrimSpace(*user.Avatar) == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Avatar no encontrado"})
		return
	}

	avatarPath := h.resolveAvatarPath(*user.Avatar)
	if avatarPath == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "Avatar no encontrado"})
		return
	}

	if _, err := os.Stat(avatarPath); err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Avatar no encontrado"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No se pudo acceder al avatar"})
		return
	}

	c.Header("Cache-Control", "private, no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.File(avatarPath)
}

// ── Cookies ───────────────────────────────────────────────────────────────────

func (h *AuthHandler) setAccessCookie(c *gin.Context, token string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "access_token",
		Value:    token,
		MaxAge:   h.cfg.JWT.AccessTTLSecs,
		Path:     "/",
		Domain:   h.cfg.Cookie.Domain,
		Secure:   h.cfg.Cookie.Secure,
		HttpOnly: true,
		SameSite: parseSameSite(h.cfg.Cookie.SameSite),
	})
}

func (h *AuthHandler) setRefreshCookie(c *gin.Context, jti string, exp time.Time) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_jti",
		Value:    jti,
		MaxAge:   int(time.Until(exp).Seconds()),
		Path:     "/api/auth",
		Domain:   h.cfg.Cookie.Domain,
		Secure:   h.cfg.Cookie.Secure,
		HttpOnly: true,
		SameSite: parseSameSite(h.cfg.Cookie.SameSite),
	})
}

func (h *AuthHandler) setCsrfCookie(c *gin.Context, token string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "csrftoken",
		Value:    token,
		MaxAge:   h.cfg.JWT.RefreshTTLSecs,
		Path:     "/",
		Domain:   h.cfg.Cookie.Domain,
		Secure:   h.cfg.Cookie.Secure,
		HttpOnly: false, // Must be readable by JS for double-submit pattern
		SameSite: parseSameSite(h.cfg.Cookie.SameSite),
	})
}

func (h *AuthHandler) clearCookies(c *gin.Context) {
	ss := parseSameSite(h.cfg.Cookie.SameSite)
	for _, name := range []string{"access_token", "csrftoken"} {
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     name,
			Value:    "",
			MaxAge:   -1,
			Path:     "/",
			Domain:   h.cfg.Cookie.Domain,
			Secure:   h.cfg.Cookie.Secure,
			HttpOnly: name == "access_token",
			SameSite: ss,
		})
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_jti",
		Value:    "",
		MaxAge:   -1,
		Path:     "/api/auth",
		Domain:   h.cfg.Cookie.Domain,
		Secure:   h.cfg.Cookie.Secure,
		HttpOnly: true,
		SameSite: ss,
	})
}

func newCSRFToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func parseSameSite(s string) http.SameSite {
	switch strings.ToLower(s) {
	case "strict":
		return http.SameSiteStrictMode
	case "none":
		return http.SameSiteNoneMode
	default:
		return http.SameSiteLaxMode
	}
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func formatUser(d interface{}) interface{} {
	switch v := d.(type) {
	case *appauth.UserWithDetails:
		if v == nil {
			return nil
		}

		out := *v
		if v.Avatar != nil && strings.TrimSpace(*v.Avatar) != "" {
			avatarURL := fmt.Sprintf("/api/users/%d/avatar", v.ID)
			out.Avatar = &avatarURL
		}

		return &out
	case appauth.UserWithDetails:
		out := v
		if v.Avatar != nil && strings.TrimSpace(*v.Avatar) != "" {
			avatarURL := fmt.Sprintf("/api/users/%d/avatar", v.ID)
			out.Avatar = &avatarURL
		}
		return out
	default:
		return d
	}
}

func newAvatarFilename(userID int32, originalName string) (string, error) {
	randomBytes := make([]byte, 8)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}

	ext := strings.ToLower(filepath.Ext(originalName))
	return fmt.Sprintf("%d-%d-%s%s", userID, time.Now().UnixNano(), hex.EncodeToString(randomBytes), ext), nil
}

func isAllowedAvatarContentType(contentType string) bool {
	switch contentType {
	case "image/jpeg", "image/png", "image/gif", "image/webp":
		return true
	default:
		return false
	}
}

func (h *AuthHandler) resolveAvatarPath(raw string) string {
	filename := filepath.Base(strings.TrimSpace(raw))
	if filename == "." || filename == "" || filename == "/" {
		return ""
	}

	if strings.HasPrefix(raw, "/media/avatars/") {
		return filepath.Join(h.cfg.Media.Dir, "avatars", filename)
	}

	return filepath.Join(h.cfg.Media.PrivateDir, "avatars", filename)
}
