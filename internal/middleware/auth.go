package middleware

import (
	"net/http"
	"strings"
	"time"

	"attendancemgmt/backend/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uint            `json:"user_id"`
	Username string          `json:"username"`
	Role     models.UserRole `json:"role"`
	CenterID *uint           `json:"center_id,omitempty"`
	DeptID   *uint           `json:"dept_id,omitempty"`
	jwt.RegisteredClaims
}

func GenerateToken(user *models.User, secret string) (string, error) {
	claims := Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		CenterID: user.CenterID,
		DeptID:   user.DepartmentID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(3 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func AuthMiddleware(secret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authorization header required"})
			c.Abort()
			return
		}
		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			c.Abort()
			return
		}
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", string(claims.Role))
		c.Set("center_id", claims.CenterID)
		c.Set("dept_id", claims.DeptID)
		c.Next()
	}
}

func RequireRole(roles ...models.UserRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := models.UserRole(c.GetString("role"))
		if role == models.RoleSuperAdmin {
			c.Next()
			return
		}
		for _, r := range roles {
			if r == role {
				c.Next()
				return
			}
		}
		c.JSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
		c.Abort()
	}
}

func GetUserID(c *gin.Context) uint {
	val, _ := c.Get("user_id")
	if id, ok := val.(uint); ok {
		return id
	}
	return 0
}

func GetUserRole(c *gin.Context) models.UserRole {
	return models.UserRole(c.GetString("role"))
}

func GetUserDeptID(c *gin.Context) *uint {
	val, _ := c.Get("dept_id")
	if id, ok := val.(*uint); ok {
		return id
	}
	return nil
}

func GetUserCenterID(c *gin.Context) *uint {
	val, _ := c.Get("center_id")
	if id, ok := val.(*uint); ok {
		return id
	}
	return nil
}
