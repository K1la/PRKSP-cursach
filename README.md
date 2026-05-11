# ParkEase

![Go](https://img.shields.io/badge/Go-1.22+-00ADD8)
![Next.js](https://img.shields.io/badge/Next.js-14-black)
![License](https://img.shields.io/badge/license-MIT-green)

Fullstack service for searching and booking parking spots. The backend is written in Go with chi, pgx and PostgreSQL; the frontend uses Next.js App Router, TypeScript, Tailwind CSS and Leaflet.

## Quick Start

```bash
docker-compose up --build
```

Backend: `http://localhost:8080`  
Frontend: `http://localhost:3000`  
Health: `http://localhost:8080/health`

## Development

```bash
go run ./cmd/server
cd frontend && npm run dev
```

Set local variables from `.env.example`. PostgreSQL is expected at `postgres://postgres:postgres@localhost:5432/parking?sslmode=disable`.

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
```

## Useful Commands

```bash
make test
make fuzz
make seed
make frontend-dev
make frontend-build
```

## API Snapshot

| Method | Path | Description |
| --- | --- | --- |
| GET | `/health` | Service and database status |
| POST | `/api/v1/auth/register` | User registration |
| POST | `/api/v1/auth/login` | User login |
| GET | `/api/v1/parking-lots` | Parking search |
| POST | `/api/v1/bookings` | Create booking |

More endpoints will be implemented in the next backend stage.

## License

MIT
