# ParkEase

![Go](https://img.shields.io/badge/Go-1.22+-00ADD8)
![Next.js](https://img.shields.io/badge/Next.js-14-black)
![License](https://img.shields.io/badge/license-MIT-green)

Fullstack service for searching and booking parking spots. The backend is written in Go with chi, pgx and PostgreSQL; the frontend uses Next.js App Router, TypeScript, Tailwind CSS and Leaflet.

## Screenshot

![Home page placeholder](docs/screenshot-placeholder.svg)

## Quick Start

```bash
docker compose up --build
```

Backend: `http://localhost:8080`  
Frontend: `http://localhost:3000`  
Health: `http://localhost:8080/health`

On startup the backend applies SQL migrations from `migrations/*.up.sql` once using the `schema_migrations` table. Disable this behavior with `DB_AUTO_MIGRATE=false`.

## Development

```bash
go run ./cmd/server
cd frontend && npm run dev
```

Set local variables from `.env.example`. PostgreSQL is expected at `postgres://postgres:postgres@localhost:5432/parking?sslmode=disable`.

Seed demo data:

```bash
make seed
# or, if the database is running in Docker:
docker compose --profile tools run --rm seed
```

Demo credentials:

| Email | Password | Role |
| --- | --- | --- |
| `admin@parkease.ru` | `password123` | admin |
| `owner1@parkease.ru` | `password123` | owner |
| `user1@parkease.ru` | `password123` | user |

## Project Structure

```text
cmd/server              Go HTTP entrypoint
internal/config         environment configuration
internal/handler        HTTP handlers and middleware
internal/model          domain models
internal/repository     pgx repositories
internal/service        business logic
internal/validator      input validation
migrations              SQL schema migrations
seeds                   demo data loader
frontend/src/app        Next.js App Router pages
frontend/src/components UI and feature components
frontend/src/lib        API client and utilities
frontend/src/types      TypeScript contracts
docs/uml                Mermaid UML diagrams
```

## Useful Commands

```bash
make test
make fuzz
make seed
make seed-docker
make frontend-dev
make frontend-lint
make frontend-build
make check
```

## Technologies

| Layer | Stack |
| --- | --- |
| Backend | Go, chi, pgx, golang-jwt, bcrypt, slog |
| Frontend | Next.js 14 App Router, TypeScript, Tailwind CSS |
| Database | PostgreSQL 16 |
| Map | Leaflet, react-leaflet, OpenStreetMap |
| Containers | Docker, docker-compose |
| Tests | Go testing, fuzz tests |

## API Documentation

See [docs/api.md](docs/api.md).

## Testing

```bash
go test ./... -count=1
go test ./internal/validator/ -fuzz=. -fuzztime=30s
cd frontend && npm run lint && npm run build
```

## Smoke Checklist

1. Run `docker compose up --build`.
2. Open `http://localhost:8080/health` and check `db: connected`.
3. Run `docker compose --profile tools run --rm seed`.
4. Open `http://localhost:3000`.
5. Login as `user1@parkease.ru / password123`.
6. Search parking lots, open a parking lot, create a booking, cancel it in dashboard.
7. Login as `owner1@parkease.ru / password123`, create and edit a parking lot.
8. Login as `admin@parkease.ru / password123`, check stats and users table.

## 12-Factor App

| Factor | Implementation |
| --- | --- |
| Config | Environment variables in `.env.example` |
| Backing services | PostgreSQL via `DATABASE_URL` |
| Processes | Stateless Go API and Next.js frontend |
| Logs | Structured JSON logs to stdout through `slog` |
| Port binding | `PORT` for backend, `3000` for frontend |
| Disposability | Graceful shutdown in `cmd/server/main.go` |

## Deploy

For Railway or another Docker platform, provide `DATABASE_URL`, `JWT_SECRET`, `CORS_ORIGINS`, `NEXT_PUBLIC_API_URL`, then deploy backend and frontend containers from the included Dockerfiles.

## UML

- [Use case](docs/uml/use-case.md)
- [ER diagram](docs/uml/er-diagram.md)
- [Booking sequence](docs/uml/sequence-booking.md)
- [Component diagram](docs/uml/component.md)
- [Class diagram](docs/uml/class-diagram.md)

## License

MIT
