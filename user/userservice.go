package user

import (
	"database/sql"
	"log"
)

type User struct {
	userId   int64
	Username string
	password string
}

func Create(username, password string) User {
	return User{Username: username, password: password}
}
func Insert(db *sql.DB, user User) (User, error) {
	sqlStatement := `
		INSERT INTO users (username, password)
		VALUES ($1, $2)
		RETURNING user_id`

	exec, err := db.Exec(sqlStatement, user.Username, user.password)
	if err != nil {
		log.Println(err)
	}

	user.userId, err = exec.LastInsertId()

	return user, err
}
