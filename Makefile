.PHONY: run build test fuzz fuzz-validator fuzz-json migrate-up migrate-down seed seed-docker docker docker-prod lint frontend-dev frontend-lint frontend-build check

run:
	go run ./cmd/server

build:
	CGO_ENABLED=0 go build -o bin/server ./cmd/server

test:
	go test ./... -v -count=1

fuzz:
	go test ./internal/validator/ -fuzz=. -fuzztime=30s
	go test ./internal/handler/ -fuzz=FuzzDecodeJSON -fuzztime=30s

fuzz-validator:
	go test ./internal/validator/ -fuzz=. -fuzztime=30s

fuzz-json:
	go test ./internal/handler/ -fuzz=FuzzDecodeJSON -fuzztime=30s

migrate-up:
	goose -dir migrations postgres "$$DATABASE_URL" up

migrate-down:
	goose -dir migrations postgres "$$DATABASE_URL" down

seed:
	go run ./seeds/seed.go

seed-docker:
	docker compose --profile tools run --rm seed

docker:
	docker compose up --build

docker-prod:
	docker compose -f docker-compose.prod.yml up --build -d

lint:
	golangci-lint run ./...

frontend-dev:
	cd frontend && npm run dev

frontend-lint:
	cd frontend && npm run lint

frontend-build:
	cd frontend && npm run build

check: test frontend-lint frontend-build
	docker compose config
