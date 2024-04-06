package user

import (
	"database/sql"
	"log"
	"log/slog"
)

type UserEntity struct {
	userId   int64  `field:"user_id"`
	Username string `field:"username"`
	Password string `field:"password"`
}

func Create(username, password string) UserEntity {
	return UserEntity{Username: username, Password: password}
}

func Exists(db *sql.DB, username string) (bool, error) {
	stmt, err := db.Prepare("select count(user_id) from users where username = $1")
	if err != nil {
		slog.Error("Unable to prepare select user exist statement.", err)
		log.Fatal(err)
	}

	defer stmt.Close()
	count := 0

	if err = stmt.QueryRow(username).Scan(&count); err != nil {
		slog.Error("Unable to scan user exist row.", err)
	}

	return count > 0, err
}

func Insert(db *sql.DB, user UserEntity) (UserEntity, error) {
	sqlStatement := `
		INSERT INTO users (username, password)
		VALUES ($1, $2)
		RETURNING user_id`

	stmt, err := db.Prepare(sqlStatement)
	if err != nil {
		slog.Error("Unable to prepare insert user statement.", err)
		log.Fatal(err)
	}

	err = stmt.QueryRow(user.Username, user.Password).Scan(&user.userId)
	return user, err
}

func FindByUsername(db *sql.DB, username string) (UserEntity, error) {
	var user = UserEntity{}
	stmt, err := db.Prepare("select user_id, password from users where username = $1")
	if err != nil {
		slog.Error("Unable to prepare find user by username statement.", err)
		log.Fatal(err)

	}
	defer stmt.Close()

	err = stmt.QueryRow(username).Scan(&user.userId, &user.Password)
	user = UserEntity{userId: user.userId, Username: username, Password: user.Password}
	return user, err
}
