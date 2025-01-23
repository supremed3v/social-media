package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/supremed3v/social-media/internal/store"
)

func (app *application) AuthTokenMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			app.unauthJwtErr(w, r, fmt.Errorf("authorization header is missing"))
			return
		}
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			app.unauthJwtErr(w, r, fmt.Errorf("authorization header is malformed"))
			return
		}
		token := parts[1]

		jwtToken, err := app.authenticator.ValidateToken(token)
		if err != nil {
			app.unAuthorizedErr(w, r, err)
			return
		}

		claims := jwtToken.Claims.(jwt.MapClaims)

		userID, err := strconv.ParseInt(fmt.Sprintf("%.f", claims["subs"]), 10, 64)
		if err != nil {
			app.unAuthorizedErr(w, r, err)
			return
		}

		ctx := r.Context()
		user, err := app.getUser(ctx, userID)
		if err != nil {
			app.unAuthorizedErr(w, r, err)
			return
		}
		ctx = context.WithValue(ctx, userCtx, user)
		next.ServeHTTP(w, r.WithContext(ctx))

	})
}

func (app *application) BasicAuthMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Read auth header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				app.unauthBasicErr(w, r, fmt.Errorf("authorization header is missing"))
				return
			}
			// parse it -> get the base64
			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Basic" {
				app.unauthBasicErr(w, r, fmt.Errorf("authorization header is malformed"))
				return
			}
			// decode it
			decoded, err := base64.StdEncoding.DecodeString(parts[1])
			if err != nil {

				app.unauthBasicErr(w, r, err)
				return
			}

			// check the credentials
			username := app.config.auth.basic.user
			pass := app.config.auth.basic.pass

			creds := strings.SplitN(string(decoded), ":", 2)
			if len(creds) != 2 || creds[0] != username || creds[1] != pass {
				app.unauthBasicErr(w, r, fmt.Errorf("invalid credentials"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (app *application) checkPostOwnership(requiredRole string, next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := getUserFromContext(r)

		fmt.Println("User: ", user)

		post := getPostFromCtx(r)

		// check if it is the user's post
		if post.UserID == user.ID {
			next.ServeHTTP(w, r)
			return
		}
		// role precedence check

		allowed, err := app.checkRolePrecedence(r.Context(), user, requiredRole)

		if err != nil {
			app.internalServerError(w, r, err)
			return
		}

		if !allowed {
			app.forbiddenResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) checkRolePrecedence(ctx context.Context, user *store.User, roleName string) (bool, error) {
	role, err := app.store.Roles.GetByName(ctx, roleName)

	if err != nil {
		return false, err
	}

	return user.Role.Level >= role.Level, nil
}

func (app *application) getUser(ctx context.Context, userID int64) (*store.User, error) {

	if !app.config.redisCfg.enabled {
		return app.store.Users.GetByID(ctx, userID)
	}

	user, err := app.cacheStorage.Users.Get(ctx, userID)
	if err != nil {
		return nil, err
	}

	if user == nil {
		user, err = app.store.Users.GetByID(ctx, userID)

		if err != nil {
			return nil, err
		}
		if err := app.cacheStorage.Users.Set(ctx, user); err != nil {
			return nil, err
		}

	}

	return user, nil

}

func (app *application) RateLimiterMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.rateLimiter.Enabled {
			if allow, retryAfter := app.rateLimiter.Allow(r.RemoteAddr); !allow {
				app.rateLimitExceededResponse(w, r, retryAfter.String())
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

func (app *application) FileUploadMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse the multipart form with a memory limit
		if err := r.ParseMultipartForm(app.config.maxMultipartMem); err != nil {
			app.badRequestError(w, r, fmt.Errorf("failed to parse multipart form: %w", err))
			return
		}

		// Retrieve the file from the form
		file, header, err := r.FormFile("file")
		if err != nil {
			app.badRequestError(w, r, fmt.Errorf("failed to retrieve file: %w", err))
			return
		}
		defer file.Close()

		// Check file size
		if header.Size > app.config.maxMultipartMem {
			app.badRequestError(w, r, fmt.Errorf("file size exceeds the limit of %d bytes", app.config.maxMultipartMem))
			return
		}

		// Read file bytes
		fileBytes, err := io.ReadAll(file)
		if err != nil {
			app.badRequestError(w, r, fmt.Errorf("failed to read file: %w", err))
			return
		}

		// Check file type
		fileType := http.DetectContentType(fileBytes)
		if !strings.HasPrefix(fileType, "image") {
			app.badRequestError(w, r, fmt.Errorf("uploaded file is not an image, detected type: %s", fileType))
			return
		}

		// Use unique context key types
		type contextKey string
		const (
			fileKey     contextKey = "file"
			fileNameKey contextKey = "fileName"
		)

		// Add file data to the request context
		ctx := context.WithValue(r.Context(), fileKey, fileBytes)
		ctx = context.WithValue(ctx, fileNameKey, header.Filename)
		r = r.WithContext(ctx)

		// Pass the request to the next handler
		next.ServeHTTP(w, r)
	})
}
