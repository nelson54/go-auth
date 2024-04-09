package user

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"go_auth/config"
	"go_auth/user/userService"
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

type JsonMessage struct {
	Message string `json:"message"`
}

func Msg(msg string) JsonMessage {
	return JsonMessage{msg}
}

func Routes(router *http.ServeMux, database *sql.DB) {
	userService.SetDatabase(database)

	router.HandleFunc("GET /user", AuthMiddleware(currentUserRoute))
	router.HandleFunc("DELETE /user", AuthMiddleware(deleteUserRoute))
	router.HandleFunc("PUT /user", createUserRoute)
	router.HandleFunc("PUT /auth", authenticateRoute)
}

func saltPassword(password string) []byte {
	return []byte(fmt.Sprintf("%s:%s", config.Config().Auth.Salt, password))
}

func currentUserRoute(writer http.ResponseWriter, request *http.Request) {
	user := UserDto{}

	authContext := getAuthContext(request.Context())

	user.UserId = authContext.UserId
	user.Username = authContext.Username
	user.Roles = authContext.Roles

	writeJson(writer, user)
}

func deleteUserRoute(writer http.ResponseWriter, request *http.Request) {
	authContext := getAuthContext(request.Context())

	if ok := userService.Delete(authContext.UserId); ok {
		msg := fmt.Sprintf("Deleted user: %d", authContext.UserId)
		writeJson(writer, Msg(msg))
	} else {
		msg := fmt.Sprintf("Unable to delete user: %d", authContext.UserId)
		slog.Warn(msg)
		writer.WriteHeader(http.StatusInternalServerError)
		writeJson(writer, Msg(msg))
	}
}

func createUserRoute(writer http.ResponseWriter, request *http.Request) {
	user := createUserDto{}
	if err := json.NewDecoder(request.Body).Decode(&user); err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writeJson(writer, Msg("could not parse create user request"))
		return
	}

	if exists, err := userService.Exists(user.Username); exists {
		msg := "user already exists"
		slog.Info(msg)
		writer.WriteHeader(http.StatusInternalServerError)
		writeJson(writer, Msg(msg))
		return
	} else if err != nil {
		slog.Info("failed to check if user exists")
		writer.WriteHeader(http.StatusInternalServerError)
		writeJson(writer, Msg("user could not be created"))
		return
	}
	password := saltPassword(user.Password)
	hash, _ := bcrypt.GenerateFromPassword(password, 10)

	usr := userService.NewUserEntity(user.Username, string(hash))
	insert, err := userService.Insert(usr)

	if err != nil {
		slog.Error("Failed to create user", err)
		writer.WriteHeader(http.StatusInternalServerError)
		writeJson(writer, Msg("Unable to create user."))
		return
	}
	writer.WriteHeader(http.StatusCreated)
	response := fmt.Sprintf("Created User %s", insert.Username)
	slog.Info(response)
	writeJson(writer, Msg(response))

}

func authenticateRoute(writer http.ResponseWriter, request *http.Request) {
	var auth createUserDto
	err := json.NewDecoder(request.Body).Decode(&auth)
	authPassword := saltPassword(auth.Password)

	userFromDb, _ := userService.FindByUsername(auth.Username)
	hashedPassword := []byte(userFromDb.Password)

	err = bcrypt.CompareHashAndPassword(hashedPassword, authPassword)
	if err != nil {
		msg := "Failed to authenticate."
		slog.Info(msg)
		writer.WriteHeader(http.StatusUnauthorized)
		writeJson(writer, Msg(msg))
		return
	}

	if token, err := createToken(userFromDb); err != nil {
		msg := "Failed to authenticate."
		slog.Info(msg, err)
		writer.WriteHeader(http.StatusUnauthorized)
		writeJson(writer, Msg(msg))
		return
	} else {
		writer.WriteHeader(http.StatusCreated)
		write(writer, token)
		return
	}
}

func write(w http.ResponseWriter, msg string) {
	if _, err := w.Write([]byte(msg)); err != nil {
		slog.Warn("Failed to write http response.")
	}
}

func writeJson(w http.ResponseWriter, strct interface{}) {
	jsonBytes, err := json.Marshal(strct)
	if err != nil {
		msg := "Failed to marshal json."

		slog.Warn(msg)
		w.WriteHeader(http.StatusInternalServerError)
		writeJson(w, Msg(msg))

		return
	}

	if _, err := w.Write(jsonBytes); err != nil {
		slog.Warn("Failed to write http response.")
	}

}
