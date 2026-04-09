package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

func jsonOK(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func jsonErr(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func decode(r *http.Request, dst any) error {
	return json.NewDecoder(r.Body).Decode(dst)
}

// Handlers

// POST /api/users
func handleUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	switch r.Method {
	case http.MethodPost:
		var body struct {
			Name string `json:"name"`
		}
		if err := decode(r, &body); err != nil {
			jsonErr(w, http.StatusBadRequest, "invalid body")
			return
		}
		res, err := CreateUser(ctx, body.Name)
		if err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonOK(w, res)
	default:
		jsonErr(w, http.StatusMethodNotAllowed, "use POST")
	}
}

// GET /api/users/{id}  or  GET /api/users/{id}/ratings
func handleUserDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		jsonErr(w, http.StatusBadRequest, "missing user id")
		return
	}
	userID := parts[2]

	if len(parts) >= 4 && parts[3] == "ratings" {
		res, err := FindUserWithRatings(ctx, userID)
		if err != nil {
			jsonErr(w, http.StatusNotFound, err.Error())
			return
		}
		jsonOK(w, res)
		return
	}

	res, err := FindUser(ctx, userID)
	if err != nil {
		jsonErr(w, http.StatusNotFound, err.Error())
		return
	}
	jsonOK(w, res)
}

// POST /api/movies
func handleMovies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodPost {
		jsonErr(w, http.StatusMethodNotAllowed, "use POST")
		return
	}
	var body struct {
		Title string `json:"title"`
		Year  int    `json:"year"`
		Plot  string `json:"plot"`
	}
	if err := decode(r, &body); err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	res, err := CreateMovie(ctx, body.Title, body.Year, body.Plot)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, res)
}

// GET /api/users/search?name=...&ratings=true
func handleUserSearch(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		jsonErr(w, http.StatusBadRequest, "missing name query param")
		return
	}
	var (
		res map[string]any
		err error
	)
	if r.URL.Query().Get("ratings") == "true" {
		res, err = FindUserWithRatingsByName(r.Context(), name)
	} else {
		res, err = FindUserByName(r.Context(), name)
	}
	if err != nil {
		jsonErr(w, http.StatusNotFound, err.Error())
		return
	}
	jsonOK(w, res)
}

// GET /api/ratings/search?user=...&movie=...
func handleRatingSearch(w http.ResponseWriter, r *http.Request) {
	userName := r.URL.Query().Get("user")
	movieTitle := r.URL.Query().Get("movie")
	if userName == "" || movieTitle == "" {
		jsonErr(w, http.StatusBadRequest, "missing user and movie query params")
		return
	}
	res, err := FindUserRatingForMovie(r.Context(), userName, movieTitle)
	if err != nil {
		jsonErr(w, http.StatusNotFound, err.Error())
		return
	}
	jsonOK(w, res)
}

// GET /api/movies/search?title=...
func handleMovieSearch(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Query().Get("title")
	if title == "" {
		jsonErr(w, http.StatusBadRequest, "missing title query param")
		return
	}
	res, err := FindMovieByTitle(r.Context(), title)
	if err != nil {
		jsonErr(w, http.StatusNotFound, err.Error())
		return
	}
	jsonOK(w, res)
}

// GET /api/movies/{movieId}  or  POST /api/movies/{movieId}/genre
func handleMovieDetail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 3 {
		jsonErr(w, http.StatusBadRequest, "missing movie id")
		return
	}

	if len(parts) >= 4 && parts[3] == "genre" {
		movieID, err := strconv.Atoi(parts[2])
		if err != nil {
			jsonErr(w, http.StatusBadRequest, "movieId must be integer")
			return
		}
		var body struct {
			Genre string `json:"genre"`
		}
		if err := decode(r, &body); err != nil {
			jsonErr(w, http.StatusBadRequest, "invalid body")
			return
		}
		res, err := CreateInGenre(ctx, movieID, body.Genre)
		if err != nil {
			jsonErr(w, http.StatusInternalServerError, err.Error())
			return
		}
		jsonOK(w, res)
		return
	}

	movieID, err := strconv.Atoi(parts[2])
	if err != nil {
		jsonErr(w, http.StatusBadRequest, "movieId must be an integer")
		return
	}
	res, err := FindMovie(ctx, movieID)
	if err != nil {
		jsonErr(w, http.StatusNotFound, err.Error())
		return
	}
	jsonOK(w, res)
}

// POST /api/ratings
func handleRatings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodPost {
		jsonErr(w, http.StatusMethodNotAllowed, "use POST")
		return
	}
	var body struct {
		UserID    string `json:"userId"`
		MovieID   int    `json:"movieId"`
		Rating    int    `json:"rating"`
		Timestamp int64  `json:"timestamp"`
	}
	if err := decode(r, &body); err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	res, err := CreateRating(ctx, body.UserID, body.MovieID, body.Rating, body.Timestamp)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, res)
}

// POST /api/seed
func handleSeed(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonErr(w, http.StatusMethodNotAllowed, "use POST")
		return
	}
	res, err := seedSimple(r.Context())
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, res)
}

// POST /api/seed-extended
func handleSeedExtended(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonErr(w, http.StatusMethodNotAllowed, "use POST")
		return
	}
	res, err := seedExtended(r.Context())
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, res)
}

// POST /api/actors
func handleActors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodPost {
		jsonErr(w, http.StatusMethodNotAllowed, "use POST")
		return
	}
	var body PersonInput
	if err := decode(r, &body); err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	res, err := CreateActor(ctx, body)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, res)
}

// POST /api/actors/{tmdbId}/acted-in
func handleActorRelations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		jsonErr(w, http.StatusBadRequest, "expected /api/actors/{tmdbId}/acted-in")
		return
	}
	actorID, err := strconv.Atoi(parts[2])
	if err != nil {
		jsonErr(w, http.StatusBadRequest, "tmdbId must be integer")
		return
	}
	if parts[3] != "acted-in" {
		jsonErr(w, http.StatusNotFound, "unknown sub-route")
		return
	}
	var body struct {
		MovieID int    `json:"movieId"`
		Role    string `json:"role"`
	}
	if err = decode(r, &body); err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	res, err := CreateActedIn(ctx, actorID, body.MovieID, body.Role)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, res)
}

// POST /api/directors
func handleDirectors(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodPost {
		jsonErr(w, http.StatusMethodNotAllowed, "use POST")
		return
	}
	var body PersonInput
	if err := decode(r, &body); err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	res, err := CreateDirector(ctx, body)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, res)
}

// POST /api/directors/{tmdbId}/directed
func handleDirectorRelations(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) < 4 {
		jsonErr(w, http.StatusBadRequest, "expected /api/directors/{tmdbId}/directed")
		return
	}
	directorID, err := strconv.Atoi(parts[2])
	if err != nil {
		jsonErr(w, http.StatusBadRequest, "tmdbId must be integer")
		return
	}
	if parts[3] != "directed" {
		jsonErr(w, http.StatusNotFound, "unknown sub-route")
		return
	}
	var body struct {
		MovieID int    `json:"movieId"`
		Role    string `json:"role"`
	}
	if err = decode(r, &body); err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	res, err := CreateDirected(ctx, directorID, body.MovieID, body.Role)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, res)
}

// POST /api/genres
func handleGenres(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodPost {
		jsonErr(w, http.StatusMethodNotAllowed, "use POST")
		return
	}
	var body struct {
		Name string `json:"name"`
	}
	if err := decode(r, &body); err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	res, err := CreateGenre(ctx, body.Name)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, res)
}

// POST /api/ext-movies
func handleExtMovies(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	if r.Method != http.MethodPost {
		jsonErr(w, http.StatusMethodNotAllowed, "use POST")
		return
	}
	var body ExtMovieInput
	if err := decode(r, &body); err != nil {
		jsonErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	res, err := CreateExtMovie(ctx, body)
	if err != nil {
		jsonErr(w, http.StatusInternalServerError, err.Error())
		return
	}
	jsonOK(w, res)
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading env vars directly")
	}

	uri := os.Getenv("NEO4J_URI")
	username := os.Getenv("NEO4J_USERNAME")
	password := os.Getenv("NEO4J_PASSWORD")

	var err error
	driver, err = neo4j.NewDriverWithContext(uri, neo4j.BasicAuth(username, password, ""))
	if err != nil {
		log.Fatalf("Failed to create Neo4j driver: %v", err)
	}
	defer driver.Close(context.Background())

	ctx := context.Background()
	if err := driver.VerifyConnectivity(ctx); err != nil {
		log.Fatalf("Failed to connect to Neo4j: %v", err)
	}
	log.Println("Connected to Neo4j AuraDB")

	mux := http.NewServeMux()

	mux.HandleFunc("/api/users", handleUsers)
	mux.HandleFunc("/api/users/search", handleUserSearch)
	mux.HandleFunc("/api/users/", handleUserDetail)
	mux.HandleFunc("/api/movies", handleMovies)
	mux.HandleFunc("/api/movies/search", handleMovieSearch)
	mux.HandleFunc("/api/movies/", handleMovieDetail)
	mux.HandleFunc("/api/ratings", handleRatings)
	mux.HandleFunc("/api/ratings/search", handleRatingSearch)
	mux.HandleFunc("/api/seed", handleSeed)

	// Extended graph
	mux.HandleFunc("/api/actors", handleActors)
	mux.HandleFunc("/api/actors/", handleActorRelations)
	mux.HandleFunc("/api/directors", handleDirectors)
	mux.HandleFunc("/api/directors/", handleDirectorRelations)
	mux.HandleFunc("/api/genres", handleGenres)
	mux.HandleFunc("/api/ext-movies", handleExtMovies)
	mux.HandleFunc("/api/seed-extended", handleSeedExtended)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, mux))
}
