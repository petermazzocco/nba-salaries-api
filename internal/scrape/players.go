package scrapper

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/petermazzocco/nba-salaries/nbaData"
)

type PlayerSalary struct {
	Name       string `json:"name"`
	Salary2025 string `json:"salary2025"`
	Salary2026 string `json:"salary2026"`
	Salary2027 string `json:"salary2027"`
	Salary2028 string `json:"salary2028"`
	Salary2029 string `json:"salary2029"`
}

// Helper function to clean salary data
func cleanSalary(salary string) string {
	// Remove $ and commas
	salary = strings.ReplaceAll(salary, "$", "")
	salary = strings.ReplaceAll(salary, ",", "")
	return strings.TrimSpace(salary)
}

func NBAPlayerSalaries(ctx context.Context, pool *pgxpool.Pool, c *colly.Collector) {
	queries := nbaData.New(pool)
	var players []PlayerSalary

	// Extract data from the table rows
	c.OnHTML("table.hh-salaries-ranking-table.hh-salaries-table-sortable.responsive tbody tr", func(e *colly.HTMLElement) {
		player := PlayerSalary{}
		// Extract player name
		player.Name = e.ChildText("td.name")
		// Extract all salary years
		player.Salary2025 = e.ChildText("td:nth-of-type(4)")
		player.Salary2026 = e.ChildText("td:nth-of-type(5)")
		player.Salary2027 = e.ChildText("td:nth-of-type(6)")
		player.Salary2028 = e.ChildText("td:nth-of-type(7)")
		player.Salary2029 = e.ChildText("td:nth-of-type(8)")
		player.Salary2025 = cleanSalary(player.Salary2025)
		player.Salary2026 = cleanSalary(player.Salary2026)
		player.Salary2027 = cleanSalary(player.Salary2027)
		player.Salary2028 = cleanSalary(player.Salary2028)
		player.Salary2029 = cleanSalary(player.Salary2029)
		// Only add players with non-empty names
		if player.Name != "" {
			players = append(players, player)
		}
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	// Start scraping
	fmt.Println("Starting web scraping...")
	c.Visit("https://hoopshype.com/salaries/players/")
	fmt.Printf("Scraping complete. Found %d players.\n", len(players))

	var wg sync.WaitGroup

	// Create a semaphore to limit concurrency
	const maxConcurrent = 10
	sem := make(chan struct{}, maxConcurrent)

	errorCh := make(chan error, len(players))
	successCount := 0
	var countMutex sync.Mutex

	// For each player, insert into database using a goroutine
	for _, player := range players {
		wg.Add(1)

		playerCopy := player

		go func() {
			sem <- struct{}{}

			defer func() {
				<-sem
				wg.Done()
			}()

			insertCtx, insertCancel := context.WithTimeout(ctx, 10*time.Second)
			defer insertCancel()

			// Convert strings to pgtype.Text
			salary2025 := pgtype.Text{String: playerCopy.Salary2025, Valid: playerCopy.Salary2025 != ""}
			salary2026 := pgtype.Text{String: playerCopy.Salary2026, Valid: playerCopy.Salary2026 != ""}
			salary2027 := pgtype.Text{String: playerCopy.Salary2027, Valid: playerCopy.Salary2027 != ""}
			salary2028 := pgtype.Text{String: playerCopy.Salary2028, Valid: playerCopy.Salary2028 != ""}
			salary2029 := pgtype.Text{String: playerCopy.Salary2029, Valid: playerCopy.Salary2029 != ""}

			_, err := queries.CreatePlayerSalaries(insertCtx, nbaData.CreatePlayerSalariesParams{
				Name:       playerCopy.Name,
				Salary2025: salary2025,
				Salary2026: salary2026,
				Salary2027: salary2027,
				Salary2028: salary2028,
				Salary2029: salary2029,
			})

			if err != nil {
				errorCh <- fmt.Errorf("error inserting player %s: %v", playerCopy.Name, err)
			} else {
				countMutex.Lock()
				successCount++
				countMutex.Unlock()
			}
		}()
	}

	go func() {
		wg.Wait()
		close(errorCh)
		fmt.Println("All database operations completed")
	}()

	// Process and report errors
	errorCount := 0
	for err := range errorCh {
		fmt.Println(err)
		errorCount++
	}

	// Print summary
	fmt.Printf("\nTotal players found: %d\n", len(players))
	fmt.Printf("Players successfully inserted: %d\n", successCount)
	fmt.Printf("Players with errors: %d\n", errorCount)
}
