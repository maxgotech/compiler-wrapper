package postgres

import (
	"database/sql"
	"fmt"
	"log/slog"
	"os"

	storage "compiler-wrapper/internal/db"
	"compiler-wrapper/internal/lib/hash"

	_ "github.com/lib/pq"
)

type Storage struct {
	db *sql.DB
}

func NewStorage(log *slog.Logger) (*Storage, error) {
	const op = "storage.postgres.NewStorage"
	connStr := fmt.Sprintf(
		"host=postgres user=%s password=%s dbname=%s port=5432 sslmode=disable",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Error("could not open db connection")
		return nil, fmt.Errorf("connection error")
	}

	stmt, err := db.Prepare(`
	CREATE TABLE IF NOT EXISTS users(
		id 			serial PRIMARY KEY,
		name		varchar(40) NOT NULL,
		mail		text NOT NULL UNIQUE,
		password 	text NOT NULL UNIQUE
	);
	`)

	if err != nil {
		log.Error("Unable to create users statement:" + err.Error())
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	stmt, err = db.Prepare(`
	CREATE TABLE IF NOT EXISTS compiles(
		id 			serial PRIMARY KEY,
		language	varchar(40) NOT NULL,
		result		text NOT NULL,
		password 	text NOT NULL,
		id_user		integer REFERENCES users(id) NOT NULL,
		code		text,
		createdAt 	timestamp default current_timestamp
	);
	`)
	if err != nil {
		log.Error("Unable to create compiles statement")
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: db}, nil
}

func (s *Storage) Reg(name, mail, pass string) error {
	const op = "storage.postgres.Reg"

	_stmt := `INSERT INTO users(name, mail, password) VALUES($1, $2, $3)`

	stmt, err := s.db.Prepare(_stmt)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	// hash pass
	pass, err = hash.HashPassword(pass)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec(name, mail, pass)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (s *Storage) Log(mail, pass string) (bool, error) {
	const op = "storage.postgres.Log"

	_stmt := `SELECT password FROM users WHERE mail = &1`

	row := s.db.QueryRow(_stmt, pass)

	var hashpass string

	// result or ErrNoRows
	err := row.Scan(&hashpass)
	if err == sql.ErrNoRows {
		return false, storage.ErrUserNotFound
	}
	if err != nil {
		return false, fmt.Errorf("%s: couldn't scan: %w", op, err)
	}

	if res := hash.VerifyPassword(pass, hashpass); !res {
		return false, nil
	}

	return true, nil
}

type User struct {
	Name string `json:"name"`
	Mail string `json:"mail"`
	Pass string `json:"pass"`
}

func (s *Storage) GetUsers() ([]User, error) {
	const op = "storage.postgres.GetUsers"

	sql := `select name, mail, password from users`

	var users []User

	rows, err := s.db.Query(sql)
	if err != nil {
		return nil, fmt.Errorf("%s: couldn't query users: %s", op, err)
	}

	for rows.Next() {
		var user User
		if err := rows.Scan(&user.Name, &user.Mail, &user.Pass); err != nil {
			return users, err
		}
		users = append(users, user)
	}

	return users, nil
}
