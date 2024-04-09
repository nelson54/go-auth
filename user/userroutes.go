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

type createUserDto struct {
	Username string   `json:"username"`
	Password string   `json:"password"`
	Roles    []string `json:"roles"`
}

type UserDto struct {
	UserId   int64    `json:"id"`
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
}

var db *sql.DB

func Routes(router *http.ServeMux, database *sql.DB) {
	db = database
	router.HandleFunc("GET /user", AuthMiddleware(getUser))

	router.HandleFunc("DELETE /user", AuthMiddleware(deleteUser))

	router.HandleFunc("PUT /user", createUser)

	router.HandleFunc("PUT /auth", authenticate)
}

func saltPassword(password string) []byte {
	return []byte(fmt.Sprintf("%s:%s", config.Config().Auth.Salt, password))
}

func getUser(writer http.ResponseWriter, request *http.Request) {
	user := UserDto{}

	authContext := getAuthContext(request.Context())

	user.UserId = authContext.UserId
	user.Username = authContext.Username
	user.Roles = authContext.Roles

	userBytes, err := json.Marshal(user)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Unable to retrieve user"))
		return
	}

	writer.Write(userBytes)

}

func deleteUser(writer http.ResponseWriter, request *http.Request) {
	authContext := getAuthContext(request.Context())

	if ok := Delete(db, authContext.UserId); ok {
		msg := "Deleted user"
		writer.Write([]byte(msg))
	} else {
		msg := "Unable to delete user"
		slog.Warn(msg, authContext.UserId)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(msg))
	}

}

func createUser(writer http.ResponseWriter, request *http.Request) {
	user := createUserDto{}
	if err := json.NewDecoder(request.Body).Decode(&user); err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte("could not parse create user request"))
		return
	}

	if exists, err := Exists(db, user.Username); exists {
		msg := "user already exists"
		slog.Info(msg)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte(msg))
		return
	} else if err != nil {
		slog.Info("failed to check if user exists")
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("user could not be created"))
		return
	}
	password := saltPassword(user.Password)
	hash, _ := bcrypt.GenerateFromPassword(password, 10)

	usr := Create(user.Username, string(hash))
	insert, err := Insert(db, usr)

	if err != nil {
		slog.Error("Failed to create user", err)
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Unable to create user."))
		return
	}
	writer.WriteHeader(http.StatusCreated)
	response := fmt.Sprintf("Created User %s", insert.Username)
	writer.Write([]byte(response))
}

func authenticate(writer http.ResponseWriter, request *http.Request) {
	var auth createUserDto
	err := json.NewDecoder(request.Body).Decode(&auth)
	authPassword := saltPassword(auth.Password)

	userFromDb, _ := FindByUsername(db, auth.Username)
	hashedPassword := []byte(userFromDb.Password)

	err = bcrypt.CompareHashAndPassword(hashedPassword, authPassword)
	if err != nil {
		slog.Info("Failed to authenticate")
		writer.WriteHeader(http.StatusUnauthorized)
		writer.Write([]byte("Failed to authenticate"))
		return
	}

	if token, err := createToken(userFromDb); err != nil {
		slog.Info("Failed to authenticate.", err)
		writer.WriteHeader(http.StatusUnauthorized)
		writer.Write([]byte("Failed to authenticate."))
		return
	} else {
		writer.WriteHeader(http.StatusCreated)
		writer.Write([]byte(token))
		return
	}
}
