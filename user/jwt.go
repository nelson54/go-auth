package user

import (
	"context"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"go_auth/config"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

type ctxKey string

type authenticationContext struct {
	UserId   int64
	Username string
	Roles    []string
}

const ctxAuthKey = ctxKey("auth")

func setAuthContext(ctx context.Context, authContext *authenticationContext) context.Context {
	return context.WithValue(ctx, ctxAuthKey, authContext)
}

func getAuthContext(ctx context.Context) *authenticationContext {
	return ctx.Value(ctxAuthKey).(*authenticationContext)
}

type additionalClaims struct {
	UserId   int64    `json:"userId"`
	UserName string   `json:"username"`
	Roles    []string `json:"roles"`
	jwt.RegisteredClaims
}

func AuthMiddleware(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if ok, authContext := AuthHandler(w, r); ok {
			requestContext := r.WithContext(
				setAuthContext(r.Context(), authContext),
			)

			handler(w, requestContext)
		}
	}
}

func AuthHandler(writer http.ResponseWriter, request *http.Request) (bool, *authenticationContext) {
	var authContext authenticationContext

	authToken := request.Header.Get("Authorization")
	if authToken == "" {
		writer.WriteHeader(http.StatusUnauthorized)
		writer.Write([]byte("not authorized."))
		return false, &authContext
	}

	authorization := strings.Split(authToken, " ")
	if len(authorization) > 2 || authorization[0] == "Bearer:" {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte("badly formatted authorization."))
		return false, &authContext
	}

	token, err := verifyToken(authorization[1])
	if err != nil {
		slog.Warn("Unable to authorize", err)
		writer.WriteHeader(http.StatusUnauthorized)
		writer.Write([]byte("Invalid authorization."))
		return false, &authContext
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		var roles []string
		for _, p := range claims["roles"].([]interface{}) {
			roles = append(roles, p.(string))
		}

		authContext.UserId = int64(claims["userId"].(float64))
		authContext.Username = claims["username"].(string)
		authContext.Roles = roles

		return true, &authContext
	} else {
		writer.WriteHeader(http.StatusUnauthorized)
		writer.Write([]byte("Invalid authorization"))
		return false, &authContext
	}
}

func createToken(entity UserEntity) (string, error) {

	claims := additionalClaims{
		entity.userId,
		entity.Username,
		entity.Roles,
		jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(config.Config().Auth.Secret))

	return tokenString, err
}

func verifyToken(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.Config().Auth.Secret), nil
	})

	if !token.Valid {
		return token, fmt.Errorf("invalid token")
	}

	return token, err
}
