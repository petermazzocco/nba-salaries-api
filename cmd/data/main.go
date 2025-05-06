package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	scrapper "github.com/petermazzocco/nba-salaries/internal/scrape"
)

// Create the table if it doesn't exist
func createPlayerTableIfNotExists(ctx context.Context, conn *pgx.Conn) error {
	// Drop the table if it exists (for a clean start)
	_, err := conn.Exec(ctx, `DROP TABLE IF EXISTS nba_player_salaries`)
	if err != nil {
		return err
	}

	// Create the table with column names matching what playerData.CreateAuthor expects
	_, err = conn.Exec(ctx, `
        CREATE TABLE IF NOT EXISTS nba_player_salaries (
            id SERIAL PRIMARY KEY,
            name TEXT NOT NULL,
            salary2025 TEXT,
            salary2026 TEXT,
            salary2027 TEXT,
            salary2028 TEXT,
            salary2029 TEXT,
            salary2030 TEXT
        )
    `)
	return err
}
func createTeamTableIfNotExists(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, `DROP TABLE IF EXISTS nba_team_salaries`)
	if err != nil {
		return err
	}

	_, err = conn.Exec(ctx, `
        CREATE TABLE IF NOT EXISTS nba_team_salaries (
            id SERIAL PRIMARY KEY,
            name TEXT NOT NULL,
            salary2025 TEXT,
            salary2026 TEXT,
            salary2027 TEXT,
            salary2028 TEXT,
            salary2029 TEXT,
            salary2030 TEXT
        )
    `)
	return err
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Could not find local .env file, continuing with environment variables")
	}

	// Set up context with timeout for the entire operation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Database connection string
	url := os.Getenv("DB_URL")
	// Create a connection pool for concurrent operations
	poolConfig, err := pgxpool.ParseConfig(url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to parse connection string: %v\n", err)
		os.Exit(1)
	}

	// Set the max connections to allow for concurrent operations
	poolConfig.MaxConns = 20

	// Create the connection pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create connection pool: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	// Create a single connection for table creation
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	// Create a new collector
	c := colly.NewCollector(
		colly.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/109.0.0.0 Safari/537.36"),
	)

	err = createPlayerTableIfNotExists(ctx, conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create table: %v\n", err)
		os.Exit(1)
	}
	err = createTeamTableIfNotExists(ctx, conn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to create table: %v\n", err)
		os.Exit(1)
	}

	// Scrape the sites for salaries
	scrapper.NBAPlayerSalaries(ctx, pool, c)
	scrapper.NBATeamSalaries(ctx, pool, c)
}
