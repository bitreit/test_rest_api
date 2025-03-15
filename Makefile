BINARY_NAME = bin/myapp.exe
MAIN_FILE = cmd/url-shortener/main.go  # Путь правильный!
DOCKER_COMPOSE_PATH = internal/storage/Docker-compose.yaml

.PHONY: all build run test clean migrate

all: migrate run

migrate:
	@echo "Running database migrations..."
	docker-compose -f $(DOCKER_COMPOSE_PATH) run migrate -path=/migrations -database="postgres://Arslan:1234@postgres:5432/url?sslmode=disable" up

build:
	@echo "Building project..."
	go build -o $(BINARY_NAME) $(MAIN_FILE)

run: build
	@echo "Starting application..."
	./$(BINARY_NAME)

test:
	@echo "Running tests..."
	go test -v ./...

clean:
	@echo "Cleaning up..."
	rm -f $(BINARY_NAME)
	docker-compose -f $(DOCKER_COMPOSE_PATH) down


# docker-compose exec postgres psql -U Arslan -d url -c "\d+ url"