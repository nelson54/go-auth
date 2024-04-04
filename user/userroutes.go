package user

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"go_auth/config"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
)

func Routes(cfg config.Config, router *chi.Mux, db *sql.DB) {

	type updateUserDto struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	router.Put("/user", func(writer http.ResponseWriter, request *http.Request) {
		var user updateUserDto
		err := json.NewDecoder(request.Body).Decode(&user)

		password := []byte(fmt.Sprintf("%s:%s", cfg.User.Salt, user.Password))
		hash, _ := bcrypt.GenerateFromPassword(password, 10)

		usr := Create(user.Username, string(hash))
		insert, err := Insert(db, usr)

		if err != nil {
			log.Println(err)
		}
		response := fmt.Sprintf("Welcome %s", insert.Username)
		writer.Write([]byte(response))
	})

	router.Put("/auth", func(writer http.ResponseWriter, request *http.Request) {
		var auth updateUserDto
		err := json.NewDecoder(request.Body).Decode(&auth)

		authPassword := []byte(fmt.Sprintf("%s:%s", cfg.User.Salt, auth.Password))

		userFromDb, _ := FindByUsername(db, auth.Username)
		hashedPassword := []byte(userFromDb.password)

		err = bcrypt.CompareHashAndPassword(hashedPassword, authPassword)
		if err != nil {
			writer.Write([]byte("Failed to authenticate"))
		} else {
			response := fmt.Sprintf("Authenticated as %s", auth.Username)
			writer.Write([]byte(response))
		}

	})

}
