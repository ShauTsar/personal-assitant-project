package postgre

import (
	"database/sql"
	"errors"
	"fmt"
	"personal-assitant-project/config"
	userpb "personal-assitant-project/personal-assitant-server/grpc/proto/gen"
)

func CheckUserExists(username string) (bool, error) {
	db := LoadDBFromConfig()
	query := "SELECT COUNT(*) FROM users WHERE username = $1"
	var count int
	if err := db.QueryRow(query, username).Scan(&count); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}

func SaveUserData(request *userpb.RegisterRequest) error {
	db := LoadDBFromConfig()
	query := `
        INSERT INTO users (username, password, email, phone, timezone)
        VALUES ($1, $2, $3, $4, $5)
    `
	_, err := db.Exec(query, request.Username, request.Password, request.Email, request.Phone, request.Timezone)
	if err != nil {
		return err
	}

	return nil
}
func LoadDBFromConfig() *sql.DB {
	dbConfig := config.LoadDatabaseConfig()
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbConfig.Host, dbConfig.Port, dbConfig.Username, dbConfig.Password, dbConfig.Database, dbConfig.SSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		// Здесь нужно обработать ошибку, если соединение не установлено.
		panic(err)
	}
	return db
}
func CheckCredentials(username, password string) (bool, error) {
	db := LoadDBFromConfig()
	defer db.Close()

	query := "SELECT COUNT(*) FROM users WHERE username = $1 AND password = $2"
	var count int
	if err := db.QueryRow(query, username, password).Scan(&count); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return count > 0, nil
}
