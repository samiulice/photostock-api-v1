package api

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/samiulice/photostock/internal/models"

	"github.com/golang-jwt/jwt/v5"
)

// userContextKey is the key used to store user claims in the request context
type contextKey string

// AuthUser is a middleware that checks if the user is authenticated
// It expects the Authorization header to be present in the request
// If the header is missing or invalid, it returns a 401 Unauthorized response
// If the token is valid, it adds the user claims to the request context
// and proceeds to the next handler
func (app *application) AuthUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.errorLog.Println("No authorization header")
			app.writeJSON(w, http.StatusUnauthorized, models.Response{
				Error:   true,
				Message: "Unauthorized",
			})
			return
		}

		// Check if the Authorization header has the correct format
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			app.errorLog.Println("Invalid authorization header format")
			app.writeJSON(w, http.StatusUnauthorized, models.Response{
				Error:   true,
				Message: "Access Denied: Invalid Authorization Header",
			})
			return
		}

		// Get the token
		tokenString := parts[1]

		// Parse and validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(app.config.jwt.secretKey), nil // Replace with your actual secret key
		})

		if err != nil {
			app.errorLog.Printf("Error parsing token: %v", err)
			app.writeJSON(w, http.StatusUnauthorized, models.Response{
				Error:   true,
				Message: "Access Denied: Invalid Token",
			})
			return
		}

		// Check if the token is valid
		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			// Check if the token is expired
			exp, ok := claims["exp"].(float64)
			if !ok || float64(time.Now().Unix()) > exp {
				app.errorLog.Println("Token expired")
				app.writeJSON(w, http.StatusUnauthorized, models.Response{
					Error:   true,
					Message: "Token expired",
				})
				return
			}

			// Safely extract user fields from claims
			tokenUser := &models.JWT{}
			if id, ok := claims["id"].(int); ok {
				tokenUser.ID = id
			}
			if name, ok := claims["name"].(string); ok {
				tokenUser.Name = name
			}
			if username, ok := claims["username"].(string); ok {
				tokenUser.Username = username
			}
			if role, ok := claims["role"].(string); ok {
				tokenUser.Role = role
			}
			if iss, ok := claims["iss"].(string); ok {
				tokenUser.Issuer = iss
			}
			if aud, ok := claims["aud"].(string); ok {
				tokenUser.Audience = aud
			}
			if exp, ok := claims["exp"].(float64); ok {
				tokenUser.ExpiresAt = int64(exp)
			}
			if iat, ok := claims["iat"].(float64); ok {
				tokenUser.IssuedAt = int64(iat)
			}
			// No userStruct needed; user is already a *models.User

			// Add user struct to the request context
			ctx := context.WithValue(r.Context(), contextKey("user"), tokenUser)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			app.errorLog.Println("Invalid token")
			app.writeJSON(w, http.StatusUnauthorized, models.Response{
				Error:   true,
				Message: "Unauthorized: Invalid Token",
			})
			return
		}
	})
}

// GetUserTokenFromContext retrieves the user claims from the request context
// It returns the user struct and a boolean indicating if the user was found
// If the user is not found, it logs an error and returns nil
// This function is used in the AuthAdmin middleware to check if the user is an admin
func (app *application) GetUserTokenFromContext(ctx context.Context) (*models.JWT, bool) {
	user, ok := ctx.Value(contextKey("user")).(*models.JWT)
	if !ok || user == nil {
		app.errorLog.Println("No user found in context")
		return nil, false
	}
	return user, true
}

// AuthAdmin is a middleware that checks if the user is an admin
// It expects the user claims to be present in the request context
// If the user is not an admin, it returns a 403 Forbidden response
// If the user is an admin, it proceeds to the next handler
func (app *application) AuthAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, ok := app.GetUserTokenFromContext(r.Context())
		if !ok {
			app.writeJSON(w, http.StatusUnauthorized, models.Response{
				Error:   true,
				Message: "Unauthorized: No user found in context",
			})
			return
		}
		if token.Role != "admin" {
			app.writeJSON(w, http.StatusForbidden, models.Response{
				Error:   true,
				Message: "Forbidden: You do not have permission to access this resource",
			})
			return
		}
		// If the user is an admin, proceed to the next handler
		next.ServeHTTP(w, r)
	})
}

// Logger is a middleware that logs the details of each HTTP request.
func (app *application) Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		app.infoLog.Println("Received request:", r.Method, r.URL.Path, "from", r.RemoteAddr)
		// Log the request details
		// Using log.Printf for formatted output
		// log.Printf("Received request: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}
