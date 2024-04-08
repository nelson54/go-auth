package user

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"go_auth/config"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"net/http"
	"strings"
)

func Routes(router *http.ServeMux, db *sql.DB) {

	type userDto struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	router.HandleFunc("DELETE /user", func(writer http.ResponseWriter, request *http.Request) {
		authToken := request.Header.Get("Authorization")
		if authToken == "" {
			writer.Write([]byte("not authorized."))
			return
		}

		authorization := strings.Split(authToken, " ")
		if len(authorization) > 2 || authorization[0] == "Bearer:" {
			writer.Write([]byte("badly formatted authorization."))
			return
		}

		token, err := verifyToken(authorization[1])
		if err != nil {
			slog.Warn("Unable to authorize", err)
			writer.Write([]byte("Invalid authorization."))
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			Delete(db, claims["user_id"].(string))
		} else {
			writer.Write([]byte("Invalid JWT Token"))

		}

	})

	router.HandleFunc("PUT /user", func(writer http.ResponseWriter, request *http.Request) {
		user := userDto{}
		if err := json.NewDecoder(request.Body).Decode(&user); err != nil {
			writer.Write([]byte("could not parse create user request"))
		}

		if exists, err := Exists(db, user.Username); exists {
			msg := "user already exists"
			slog.Info(msg)
			writer.Write([]byte(msg))
		} else if err != nil {
			slog.Info("failed to check if user exists")
			writer.Write([]byte("user could not be created"))
		} else {
			password := saltPassword(user.Password)
			hash, _ := bcrypt.GenerateFromPassword(password, 10)

			usr := Create(user.Username, string(hash))
			insert, err := Insert(db, usr)

			if err != nil {
				slog.Error("Failed to create user", err)
				writer.Write([]byte("Unable to create user."))
			}
			response := fmt.Sprintf("Welcome %s", insert.Username)
			writer.Write([]byte(response))
		}
	})

	router.HandleFunc("PUT /auth", func(writer http.ResponseWriter, request *http.Request) {
		var auth userDto
		err := json.NewDecoder(request.Body).Decode(&auth)

		authPassword := saltPassword(auth.Password)

		userFromDb, _ := FindByUsername(db, auth.Username)
		hashedPassword := []byte(userFromDb.Password)

		err = bcrypt.CompareHashAndPassword(hashedPassword, authPassword)
		if err != nil {
			slog.Info("Failed to authenticate")
			writer.Write([]byte("Failed to authenticate"))
			return
		}

		if token, err := createToken(userFromDb); err != nil {
			slog.Info("Failed to authenticate.", err)
			writer.Write([]byte("Failed to authenticate."))
			return
		} else {
			writer.Write([]byte(token))
			return
		}
	})
}

func saltPassword(password string) []byte {
	return []byte(fmt.Sprintf("%s:%s", config.Config().Auth.Salt, password))
}
