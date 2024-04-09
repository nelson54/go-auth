package userService

import (
	"database/sql"
	"log"
	"log/slog"
	"slices"
)

const (
	AuthorityUser   = "ROLE_USER"
	AuthorityAdmin  = "ROLE_ADMIN"
	AuthoritySystem = "ROLE_SYSTEM"
)

type UserEntity struct {
	UserId   int64  `field:"user_id"`
	Username string `field:"username"`
	Password string `field:"password"`
	Roles    []string
}

func NewUserEntity(username, password string) UserEntity {
	return UserEntity{
		UserId:   -1,
		Username: username,
		Password: password,
		Roles:    []string{AuthorityUser},
	}
}

func FindByUsername(db *sql.DB, username string) (UserEntity, error) {
	var user = UserEntity{}
	stmt, err := db.Prepare("select user_id, password from users where username = $1")
	if err != nil {
		slog.Error("Unable to prepare find user by username statement.", err)
		log.Fatal(err)

	}
	defer stmt.Close()

	err = stmt.QueryRow(username).Scan(&user.UserId, &user.Password)
	user = UserEntity{UserId: user.UserId, Username: username, Password: user.Password}
	user = populateUserRoles(db, user)
	return user, err
}

func Exists(db *sql.DB, username string) (bool, error) {
	stmt, err := db.Prepare("select count(user_id) from users where username = $1")
	if err != nil {
		slog.Error("Unable to prepare user exists statement.", err)
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
		INSERT INTO users (username, password, created_at, updated_at)
		VALUES ($1, $2, current_timestamp, current_timestamp)
		RETURNING user_id`

	stmt, err := db.Prepare(sqlStatement)
	if err != nil {
		slog.Error("Unable to prepare insert user statement.", err)
		log.Fatal(err)
	}

	if err = stmt.QueryRow(user.Username, user.Password).Scan(&user.UserId); err != nil {
		return user, err
	}

	for _, v := range user.Roles {
		grantRole(db, user, v)
	}

	return user, err
}

func GrantRole(db *sql.DB, user UserEntity, role string) bool {

	if pos := slices.Index(user.Roles, role); user.UserId > 0 && 0 > pos {
		return false
	}

	return grantRole(db, user, role)
}

func grantRole(db *sql.DB, user UserEntity, role string) bool {
	sqlStatement := `INSERT INTO user_roles (user_id, role) VALUES ($1, $2);`

	stmt, err := db.Prepare(sqlStatement)
	if err != nil {
		slog.Error("Unable to prepare insert user statement.", err)
		log.Fatal(err)
	}

	if err = stmt.QueryRow(user.UserId, role).Err(); err != nil {
		return false
	}

	return true
}

func Delete(db *sql.DB, userId int64) bool {
	deleteUserRules(db, userId)

	stmt, err := db.Prepare(`DELETE FROM users WHERE user_id = $1;`)
	if err != nil {
		slog.Error("Unable to prepare user delete statement.", err)
		log.Fatal(err)
	}

	if err = stmt.QueryRow(userId).Err(); err != nil {
		slog.Error("Unable to scan user exist row.", err)
		return false
	}

	return true
}

func deleteUserRules(db *sql.DB, userId int64) {
	stmt, err := db.Prepare(`DELETE FROM user_roles WHERE user_id = $1;`)
	if err != nil {
		slog.Error("Unable to prepare user delete statement.", err)
		log.Fatal(err)
	}

	if err = stmt.QueryRow(userId).Err(); err != nil {
		slog.Error("Unable to scan user exist row.", err)
	}
}

func populateUserRoles(db *sql.DB, user UserEntity) UserEntity {
	stmt, err := db.Prepare("select role from user_roles where user_id = $1")
	if err != nil {
		slog.Error("Unable to prepare select roles statement.", err)
		log.Fatal(err)

	}
	roles := []string{}
	rows, err := stmt.Query(user.UserId)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	for rows.Next() {
		var claim string
		rows.Scan(&claim)
		roles = append(roles, claim)
	}

	user.Roles = roles

	return user
}
