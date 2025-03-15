package postgres

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

type PostgresSuite struct {
	suite.Suite
	container *postgres.PostgresContainer
	db        *sql.DB
}

// 1. Настройка контейнера и подключение к БД
func (s *PostgresSuite) SetupSuite() {
	ctx := context.Background()

	// Запуск PostgreSQL контейнера с автоматическим health check
	container, err := postgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	assert.NoError(s.T(), err)
	s.container = container

	// Получение строки подключения
	connStr, err := container.ConnectionString(ctx, "sslmode=disable")
	assert.NoError(s.T(), err)

	// Подключение к БД
	db, err := sql.Open("postgres", connStr)
	assert.NoError(s.T(), err)

	// Проверка миграций
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS url (
		id SERIAL PRIMARY KEY,
		alias TEXT NOT NULL UNIQUE,
		url TEXT NOT NULL
	)`)
	assert.NoError(s.T(), err)

	s.db = db
}

// 2. Очистка данных перед каждым тестом
func (s *PostgresSuite) SetupTest() {
	_, err := s.db.Exec("TRUNCATE TABLE url RESTART IDENTITY")
	assert.NoError(s.T(), err)
}

// 3. Тест сохранения и получения URL
func (s *PostgresSuite) TestSaveAndGetURL() {
	storage := &Storage{db: s.db}

	// Сохранение нового URL
	id, err := storage.SaveURL("https://example.com", "example")
	assert.NoError(s.T(), err)
	assert.NotZero(s.T(), id)

	// Получение существующего URL
	url, err := storage.GetURL("example")
	assert.NoError(s.T(), err)
	assert.Equal(s.T(), "https://example.com", url)

	// Попытка получить несуществующий URL
	_, err = storage.GetURL("nonexistent")

}

// 4. Тест уникальности алиаса
func (s *PostgresSuite) TestUniqueAliasConstraint() {
	storage := &Storage{db: s.db}

	// Первое сохранение
	_, err := storage.SaveURL("https://example.com", "example")
	assert.NoError(s.T(), err)

	// Попытка дублирования алиаса
	_, err = storage.SaveURL("https://another.com", "example")
	assert.ErrorContains(s.T(), err, "duplicate key value violates unique constraint")
}

// 5. Тест удаления URL
func (s *PostgresSuite) TestDeleteURL() {
	storage := &Storage{db: s.db}

	// Сохранение и удаление
	_, err := storage.SaveURL("https://example.com", "example")
	assert.NoError(s.T(), err)

	err = storage.DeleteURL("example", "https://example.com")
	assert.NoError(s.T(), err)

	// Проверка удаления
	var count int
	err = s.db.QueryRow("SELECT COUNT(*) FROM url WHERE alias = $1", "example").Scan(&count)
	assert.NoError(s.T(), err)
	assert.Zero(s.T(), count)
}

// 6. Завершение работы
func (s *PostgresSuite) TearDownSuite() {
	if s.db != nil {
		s.db.Close()
	}
	if s.container != nil {
		assert.NoError(s.T(), s.container.Terminate(context.Background()))
	}
}

func TestPostgresSuite(t *testing.T) {
	suite.Run(t, new(PostgresSuite))
}
