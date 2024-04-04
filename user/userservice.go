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

type GetUser struct {
	userId   int64  `field:"user_id"`
	Username string `field:"username"`
	Password string `field:"password"`
}

func FindByUsername(db *sql.DB, username string) (User, error) {
	var u = GetUser{}
	stmt, err := db.Prepare("select user_id, password from users where username = $1")
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	err = stmt.QueryRow(username).Scan(&u.userId, &u.Password)
	user := User{userId: u.userId, Username: username, password: u.Password}
	return user, err
}
