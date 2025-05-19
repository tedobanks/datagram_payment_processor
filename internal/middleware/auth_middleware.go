// paystack_supabase_backend/internal/middleware/auth_middleware.go
package middleware

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	// Ensure your module name is correct for importing utils
	"github.com/tedobanks/datagram_payment_processor/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// SupabaseClaims defines the structure of the claims within a Supabase JWT.
// You might need to adjust this based on the exact claims Supabase includes.
// Standard claims like 'exp' (expiration), 'sub' (subject/user_id) are common.
// Supabase also includes 'aud' (audience), 'iss' (issuer), 'email', 'role', etc.
type SupabaseClaims struct {
	Audience     string                 `json:"aud"`      // Audience (e.g., "authenticated")
	ExpiresAt    *jwt.NumericDate       `json:"exp"`      // Expiration time
	Subject      string                 `json:"sub"`      // User ID (this is what you need)
	Email        string                 `json:"email"`    // User's email
	Phone        string                 `json:"phone"`    // User's phone
	AppMetadata  map[string]interface{} `json:"app_metadata"`
	UserMetadata map[string]interface{} `json:"user_metadata"`
	Role         string                 `json:"role"`     // User's role (e.g., "authenticated")
	SessionId    string                 `json:"session_id"`
	// Add any other custom claims you expect from Supabase
	jwt.RegisteredClaims // Embeds standard claims like Issuer, Subject, Audience, ExpiresAt, NotBefore, IssuedAt, ID
}

// AuthMiddleware verifies Supabase JWT tokens.
func AuthMiddleware() gin.HandlerFunc {
	// Load the Supabase JWT secret from environment variables ONCE when the middleware is initialized.
	// It's better to inject this secret if your app structure allows, rather than reading env var on every call.
	// For simplicity here, we read it once.
	supabaseJWTSecret := os.Getenv("SUPABASE_JWT_SECRET")
	if supabaseJWTSecret == "" {
		// Log a fatal error because the application cannot securely operate without the JWT secret.
		log.Fatal("CRITICAL SECURITY: SUPABASE_JWT_SECRET environment variable not set. AuthMiddleware cannot function.")
	}
	jwtSecretKey := []byte(supabaseJWTSecret)

	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.RespondWithError(c, http.StatusUnauthorized, "Authorization header is required")
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.RespondWithError(c, http.StatusUnauthorized, "Authorization header format must be Bearer {token}")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Parse and validate the token
		token, err := jwt.ParseWithClaims(tokenString, &SupabaseClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Check the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			// Return the secret key for validation
			return jwtSecretKey, nil
		})

		if err != nil {
			log.Printf("Token validation error: %v", err)
			// Differentiate between different types of errors if needed
			// For example, jwt.ErrTokenExpired, jwt.ErrTokenNotValidYet
			if err == jwt.ErrTokenExpired {
				utils.RespondWithError(c, http.StatusUnauthorized, "Token has expired")
			} else if err == jwt.ErrTokenSignatureInvalid {
				utils.RespondWithError(c, http.StatusUnauthorized, "Invalid token signature")
			} else {
				utils.RespondWithError(c, http.StatusUnauthorized, "Invalid token: "+err.Error())
			}
			c.Abort()
			return
		}

		if !token.Valid {
			utils.RespondWithError(c, http.StatusUnauthorized, "Token is not valid")
			c.Abort()
			return
		}

		claims, ok := token.Claims.(*SupabaseClaims)
		if !ok {
			utils.RespondWithError(c, http.StatusInternalServerError, "Failed to parse token claims")
			c.Abort()
			return
		}

		// Token is valid, and claims are parsed.
		// Set user information in the context for subsequent handlers.
		// The 'sub' claim in a Supabase JWT typically holds the User ID.
		if claims.Subject == "" {
			utils.RespondWithError(c, http.StatusInternalServerError, "UserID (sub) not found in token claims")
			c.Abort()
			return
		}

		c.Set("userID", claims.Subject)   // This is the Supabase User ID (UUID)
		c.Set("userEmail", claims.Email) // Optional: make email available
		c.Set("userRole", claims.Role)   // Optional: make role available

		log.Printf("AuthMiddleware: UserID '%s' (Email: '%s', Role: '%s') authenticated successfully.", claims.Subject, claims.Email, claims.Role)
		c.Next() // Proceed to the next handler
	}
}
