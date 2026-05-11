.PHONY: run build test fuzz migrate-up migrate-down seed docker docker-prod lint frontend-dev frontend-build

run:
	go run ./cmd/server

build:
	CGO_ENABLED=0 go build -o bin/server ./cmd/server

test:
	go test ./... -v -count=1

fuzz:
	go test ./internal/validator/ -fuzz=. -fuzztime=30s

migrate-up:
	goose -dir migrations postgres "$$DATABASE_URL" up

migrate-down:
	goose -dir migrations postgres "$$DATABASE_URL" down

seed:
	go run ./seeds/seed.go

docker:
	docker-compose up --build

docker-prod:
	docker-compose -f docker-compose.prod.yml up --build -d

lint:
	golangci-lint run ./...

frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build
