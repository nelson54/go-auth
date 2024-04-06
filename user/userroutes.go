package user

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"go_auth/config"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"net/http"
)

func Routes(cfg config.Config, router *http.ServeMux, db *sql.DB) {

	type userDto struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

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
			password := []byte(fmt.Sprintf("%s:%s", cfg.User.Salt, user.Password))
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

		authPassword := []byte(fmt.Sprintf("%s:%s", cfg.User.Salt, auth.Password))

		userFromDb, _ := FindByUsername(db, auth.Username)
		hashedPassword := []byte(userFromDb.Password)

		err = bcrypt.CompareHashAndPassword(hashedPassword, authPassword)
		if err != nil {
			slog.Info("Failed to authenticate")
			writer.Write([]byte("Failed to authenticate"))
		} else {
			response := fmt.Sprintf("Authenticated as %s", auth.Username)
			writer.Write([]byte(response))
		}
	})
}
