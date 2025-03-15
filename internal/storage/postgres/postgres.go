package postgres

import (
	"database/sql"
	"errors"
	"fmt"
	_ "github.com/lib/pq"
	"url-shortener/internal/config"
	"url-shortener/internal/storage"
)

type Storage struct {
	db *sql.DB
}

func New(storagePath string) (*Storage, error) {
	cfg := config.MustLoad()
	const op = "storage.Postgres.New"

	db, err := sql.Open("postgres", cfg.StoragePostgres)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{db: db}, nil
}
func (s *Storage) SaveURL(urlToSave string, alias string) (int64, error) {
	const op = "storage.Postgres.New"
	// Проверка подключения
	stmt, err := s.db.Prepare("INSERT INTO url(alias, url) VALUES ($1, $2) RETURNING id")

	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	var id int64
	err = stmt.QueryRow(alias, urlToSave).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil
}
func (s *Storage) GetURL(alias string) (string, error) {
	const op = "storage.Postgres.GetURL"

	stmt, err := s.db.Prepare("SELECT url FROM url WHERE alias = ?")
	if err != nil {
		return "", fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	var resURL string

	err = stmt.QueryRow(alias).Scan(&resURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", storage.ErrURLNotFound
		}

		return "", fmt.Errorf("%s: execute statement: %w", op, err)
	}

	return resURL, nil
}
func (s *Storage) SearchAlias(alias string) (bool, error) {
	const op = "storage.Postgres.searchAlias"

	stmt, err := s.db.Prepare("SELECT COUNT(*) FROM storage WHERE alias = ?")
	if err != nil {
		return true, fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	var count int
	err = stmt.QueryRow(alias).Scan(&count)
	if err != nil {
		return true, fmt.Errorf("%s: ошибка выполнения запроса: %w", op, err)
	}

	return count == 0, nil
}

func (s *Storage) DeleteURL(alias string, urlToDelete string) error {
	const op = "storage.Postgres.DeleteURL"

	// Подготавливаем SQL-запрос
	query := "DELETE FROM url WHERE alias = $1 AND url = $2"
	stmt, err := s.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("%s: prepare statement: %w", op, err)
	}

	result, err := stmt.Exec(alias, urlToDelete)
	if err != nil {
		return fmt.Errorf("%s: execute statement: %w", op, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%s: get rows affected: %w", op, err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("%s: no rows deleted (alias: %s, url: %s)", op, alias, urlToDelete)
	}

	return nil
}
