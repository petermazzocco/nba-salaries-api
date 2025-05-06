# NBA Salaries API

A simple API that provides NBA team and player salary information.

## Overview

This project scrapes and stores NBA salary data in a PostgreSQL database and exposes it through a RESTful API. It consists of two main components:

- **Data Component**: Scrapes NBA salary information and stores it in a PostgreSQL database
- **API Component**: Provides REST endpoints to access the stored salary information

## Live API

The API is available at:
```
https://nba-salaries-api.traefik.me/
```

## API Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /` | API information |
| `GET /players` | Get all player salaries |
| `GET /players/{id}` | Get player salary by ID |
| `GET /teams` | Get all team salaries |
| `GET /teams/{id}` | Get team salary by ID |

Each API returns valid JSON for you to use.

## Rate Limiting

The API is currently rate-limited to 5 requests per minute.

## Local Development

### Prerequisites

- Go 1.21+
- PostgreSQL
- Make (optional, for using the Makefile commands)

### Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/petermazzocco/nba-salaries-api.git
   cd nba-salaries-api
   ```

2. Set up your environment variables:
   Create a `.env` file with your PostgreSQL connection string:
   ```
   DB_URL=postgres://username:password@localhost:5432/dbname
   ```
   With the database, you can run the `make data` command and store all salary information to your database.

### Running the Application

Using Make:

```bash
# Build both components
make build

# Run the data component (scrapes and stores data)
make run-data
# or directly without building
make data

# Run the API server
make run-api
# or directly without building
make api
```

Without Make:

```bash
# Build the data component
go build -o bin/nba-salaries-data ./cmd/data

# Run the data component
./bin/nba-salaries-data

# Build the API component
go build -o bin/nba-salaries-api ./cmd/api

# Run the API server
./bin/nba-salaries-api
```

### Available Make Commands

Use `make help` to see all available commands.

## License

[MIT License](LICENSE)
