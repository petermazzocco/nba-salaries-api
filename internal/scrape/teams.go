package scrapper

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gocolly/colly/v2"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/petermazzocco/nba-salaries/nbaData"
)

type TeamSalary struct {
	Name       string `json:"name"`
	Salary2025 string `json:"salary2025"`
	Salary2026 string `json:"salary2026"`
	Salary2027 string `json:"salary2027"`
	Salary2028 string `json:"salary2028"`
	Salary2029 string `json:"salary2029"`
}

func NBATeamSalaries(ctx context.Context, pool *pgxpool.Pool, c *colly.Collector) {

	queries := nbaData.New(pool)
	var teams []TeamSalary

	// Main scraping function
	c.OnHTML("table#team_summary", func(e *colly.HTMLElement) {
		e.ForEach("tbody tr", func(i int, row *colly.HTMLElement) {
			team := TeamSalary{}
			team.Name = row.ChildText("td[data-stat='team_name']")

			// Extract salary data with debug info
			team.Salary2025 = row.ChildText("td[data-stat='y1']")
			team.Salary2026 = row.ChildText("td[data-stat='y2']")
			team.Salary2027 = row.ChildText("td[data-stat='y3']")
			team.Salary2028 = row.ChildText("td[data-stat='y4']")
			team.Salary2029 = row.ChildText("td[data-stat='y5']")
			team.Salary2025 = cleanSalary(team.Salary2025)
			team.Salary2026 = cleanSalary(team.Salary2026)
			team.Salary2027 = cleanSalary(team.Salary2027)
			team.Salary2028 = cleanSalary(team.Salary2028)
			team.Salary2029 = cleanSalary(team.Salary2029)

			teams = append(teams, team)
		})

	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting", r.URL)
	})

	c.OnError(func(r *colly.Response, err error) {
		fmt.Printf("Error while scraping %s: %s\n", r.Request.URL, err)
		fmt.Printf("Response status code: %d\n", r.StatusCode)
	})

	// Start scraping
	fmt.Println("Starting web scraping for team salaries...")
	c.Visit("https://www.basketball-reference.com/contracts/#team_summary")
	fmt.Printf("Scraping complete. Found %d teams.\n", len(teams))

	var wg sync.WaitGroup

	// Create a semaphore to limit concurrency
	const maxConcurrent = 10
	sem := make(chan struct{}, maxConcurrent)

	errorCh := make(chan error, len(teams))
	successCount := 0
	var countMutex sync.Mutex

	// For each player, insert into database using a goroutine
	for _, team := range teams {
		wg.Add(1)

		teamCopy := team

		go func() {
			sem <- struct{}{}

			defer func() {
				<-sem
				wg.Done()
			}()

			insertCtx, insertCancel := context.WithTimeout(ctx, 10*time.Second)
			defer insertCancel()

			// Convert strings to pgtype.Text
			salary2025 := pgtype.Text{String: teamCopy.Salary2025, Valid: teamCopy.Salary2025 != ""}
			salary2026 := pgtype.Text{String: teamCopy.Salary2026, Valid: teamCopy.Salary2026 != ""}
			salary2027 := pgtype.Text{String: teamCopy.Salary2027, Valid: teamCopy.Salary2027 != ""}
			salary2028 := pgtype.Text{String: teamCopy.Salary2028, Valid: teamCopy.Salary2028 != ""}
			salary2029 := pgtype.Text{String: teamCopy.Salary2029, Valid: teamCopy.Salary2029 != ""}

			_, err := queries.CreateTeamSalaries(insertCtx, nbaData.CreateTeamSalariesParams{
				Name:       teamCopy.Name,
				Salary2025: salary2025,
				Salary2026: salary2026,
				Salary2027: salary2027,
				Salary2028: salary2028,
				Salary2029: salary2029,
			})

			if err != nil {
				errorCh <- fmt.Errorf("error inserting team %s: %v", teamCopy.Name, err)
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
	fmt.Printf("\nTotal teams found: %d\n", len(teams))
	fmt.Printf("Teams successfully inserted: %d\n", successCount)
	fmt.Printf("Teams with errors: %d\n", errorCount)
}
