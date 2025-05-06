package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httprate"
	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/petermazzocco/nba-salaries/nbaData"
)

func writeFormattedJSON(v any) (string, error) {
	json, err := json.Marshal(v)
	if err != nil {
		return "", fmt.Errorf("An error occurred marshalling JSON")
	}

	return string(json), nil
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("Could not find local .env file, continuing with environment variables")
	}

	// Set up context with timeout for the entire operation
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Database connection
	url := os.Getenv("DB_URL")
	conn, err := pgx.Connect(ctx, url)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(ctx)

	// Chi routing
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	// Rate limiting
	r.Use(httprate.Limit(
		5,
		time.Minute,
		httprate.WithLimitHandler(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, `{"error": "Rate-limited. Max request 5 per minute. Please, slow down."}`, http.StatusTooManyRequests)
		}),
	))

	// API Routes
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("NBA salaries API for teams and players."))
	})

	r.Route("/players", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			queries := nbaData.New(conn)

			players, err := queries.GetPlayersSalaries(ctx)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
			}

			json, err := writeFormattedJSON(players)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}

			w.Write([]byte(json))
		})
		r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			queries := nbaData.New(conn)
			idInt, err := strconv.Atoi(id)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
			}
			player, err := queries.GetPlayersSalaryByID(ctx, int64(idInt))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
			}

			json, err := writeFormattedJSON(player)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}

			w.Write([]byte(json))
		})
	})

	r.Route("/teams", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			queries := nbaData.New(conn)

			teams, err := queries.GetTeamsSalaries(ctx)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
			}

			json, err := writeFormattedJSON(teams)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}

			w.Write([]byte(json))
		})
		r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "id")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)

			queries := nbaData.New(conn)
			idInt, err := strconv.Atoi(id)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
			}
			team, err := queries.GetTeamSalaryByID(ctx, int64(idInt))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
			}

			json, err := writeFormattedJSON(team)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
			}

			w.Write([]byte(json))
		})
	})

	http.ListenAndServe(":8080", r)
}
