package main

import (
	"context"
	"math/rand"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
)

var driver neo4j.DriverWithContext

func session(ctx context.Context) neo4j.SessionWithContext {
	return driver.NewSession(ctx, neo4j.SessionConfig{DatabaseName: os.Getenv("NEO4J_DATABASE")})
}

// Structs to represent the entities for / creation

type PersonInput struct {
	Name   string `json:"name"`
	TmdbID int    `json:"tmdbId"`
	Born   string `json:"born"` // datetime string e.g. "1970-01-01"
	Died   string `json:"died"`
	BornIn string `json:"bornIn"`
	URL    string `json:"url"`
	ImdbID int    `json:"imdbId"`
	Bio    string `json:"bio"`
	Poster string `json:"poster"`
}

type ExtMovieInput struct {
	Title      string   `json:"title"`
	TmdbID     int      `json:"tmdbId"`
	Released   string   `json:"released"`
	ImdbRating float64  `json:"imdbRating"`
	Year       int      `json:"year"`
	ImdbID     int      `json:"imdbId"`
	Runtime    int      `json:"runtime"`
	Countries  []string `json:"countries"`
	ImdbVotes  int      `json:"imdbVotes"`
	URL        string   `json:"url"`
	Revenue    int      `json:"revenue"`
	Plot       string   `json:"plot"`
	Poster     string   `json:"poster"`
	Budget     int      `json:"budget"`
	Languages  []string `json:"languages"`
}

// CreateUser creates a :User node with an auto-generated userId.
func CreateUser(ctx context.Context, name string) (map[string]any, error) {
	s := session(ctx)
	defer s.Close(ctx)
	res, err := s.Run(ctx,
		`CREATE (u:User {name: $name, userId: $userId}) RETURN u`,
		map[string]any{"name": name, "userId": uuid.NewString()},
	)
	if err != nil {
		return nil, err
	}
	rec, err := res.Single(ctx)
	if err != nil {
		return nil, err
	}
	return rec.AsMap(), nil
}

// CreateMovie creates a :Movie node with an auto-generated movieId.
func CreateMovie(ctx context.Context, title string, year int, plot string) (map[string]any, error) {
	s := session(ctx)
	defer s.Close(ctx)
	res, err := s.Run(ctx,
		`CREATE (m:Movie {title: $title, movieId: $movieId, year: $year, plot: $plot}) RETURN m`,
		map[string]any{"title": title, "movieId": rand.Intn(1_000_000), "year": year, "plot": plot},
	)
	if err != nil {
		return nil, err
	}
	rec, err := res.Single(ctx)
	if err != nil {
		return nil, err
	}
	return rec.AsMap(), nil
}

// CreateRating creates a RATED relationship between a User and a Movie.
func CreateRating(ctx context.Context, userID string, movieID, rating int, timestamp int64) (map[string]any, error) {
	s := session(ctx)
	defer s.Close(ctx)
	if timestamp == 0 {
		timestamp = time.Now().Unix()
	}
	res, err := s.Run(ctx,
		`MATCH (u:User {userId: $userId}), (m:Movie {movieId: $movieId})
		 CREATE (u)-[r:RATED {rating: $rating, timestamp: $timestamp}]->(m)
		 RETURN u.name AS userName, m.title AS movieTitle, r.rating AS rating, r.timestamp AS timestamp`,
		map[string]any{"userId": userID, "movieId": movieID, "rating": rating, "timestamp": timestamp},
	)
	if err != nil {
		return nil, err
	}
	rec, err := res.Single(ctx)
	if err != nil {
		return nil, err
	}
	return rec.AsMap(), nil
}

// FindUser returns a :User node by userId.
func FindUser(ctx context.Context, userID string) (map[string]any, error) {
	s := session(ctx)
	defer s.Close(ctx)
	res, err := s.Run(ctx,
		`MATCH (u:User {userId: $userId}) RETURN u`,
		map[string]any{"userId": userID},
	)
	if err != nil {
		return nil, err
	}
	rec, err := res.Single(ctx)
	if err != nil {
		return nil, err
	}
	return rec.AsMap(), nil
}

// FindMovie returns a :Movie node by movieId.
func FindMovie(ctx context.Context, movieID int) (map[string]any, error) {
	s := session(ctx)
	defer s.Close(ctx)
	res, err := s.Run(ctx,
		`MATCH (m:Movie {movieId: $movieId}) RETURN m`,
		map[string]any{"movieId": movieID},
	)
	if err != nil {
		return nil, err
	}
	rec, err := res.Single(ctx)
	if err != nil {
		return nil, err
	}
	return rec.AsMap(), nil
}

// FindUserByName returns a :User node by name.
func FindUserByName(ctx context.Context, name string) (map[string]any, error) {
	s := session(ctx)
	defer s.Close(ctx)
	res, err := s.Run(ctx,
		`MATCH (u:User {name: $name}) RETURN u`,
		map[string]any{"name": name},
	)
	if err != nil {
		return nil, err
	}
	rec, err := res.Single(ctx)
	if err != nil {
		return nil, err
	}
	return rec.AsMap(), nil
}

// FindUserWithRatingsByName returns a :User and all their RATED relationships, matched by name.
func FindUserWithRatingsByName(ctx context.Context, name string) (map[string]any, error) {
	s := session(ctx)
	defer s.Close(ctx)
	res, err := s.Run(ctx,
		`MATCH (u:User {name: $name})
		 OPTIONAL MATCH (u)-[r:RATED]->(m:Movie)
		 RETURN u.name AS name, u.userId AS userId,
		        collect({title: m.title, movieId: m.movieId, rating: r.rating, timestamp: r.timestamp}) AS ratings`,
		map[string]any{"name": name},
	)
	if err != nil {
		return nil, err
	}
	rec, err := res.Single(ctx)
	if err != nil {
		return nil, err
	}
	return rec.AsMap(), nil
}

// FindUserRatingForMovie returns a specific user's rating for a specific movie.
func FindUserRatingForMovie(ctx context.Context, userName, movieTitle string) (map[string]any, error) {
	s := session(ctx)
	defer s.Close(ctx)
	res, err := s.Run(ctx,
		`MATCH (u:User {name: $name})-[r:RATED]->(m:Movie {title: $title})
		 RETURN u.name AS userName, m.title AS movieTitle, r.rating AS rating, r.timestamp AS timestamp`,
		map[string]any{"name": userName, "title": movieTitle},
	)
	if err != nil {
		return nil, err
	}
	rec, err := res.Single(ctx)
	if err != nil {
		return nil, err
	}
	return rec.AsMap(), nil
}

// FindMovieByTitle returns a :Movie node by title.
func FindMovieByTitle(ctx context.Context, title string) (map[string]any, error) {
	s := session(ctx)
	defer s.Close(ctx)
	res, err := s.Run(ctx,
		`MATCH (m:Movie {title: $title}) RETURN m`,
		map[string]any{"title": title},
	)
	if err != nil {
		return nil, err
	}
	rec, err := res.Single(ctx)
	if err != nil {
		return nil, err
	}
	return rec.AsMap(), nil
}

// FindUserWithRatings returns a :User and all their RATED relationships with movies.
func FindUserWithRatings(ctx context.Context, userID string) (map[string]any, error) {
	s := session(ctx)
	defer s.Close(ctx)
	res, err := s.Run(ctx,
		`MATCH (u:User {userId: $userId})-[r:RATED]->(m:Movie)
		 RETURN u.name AS name, u.userId AS userId,
		        collect({title: m.title, movieId: m.movieId, rating: r.rating, timestamp: r.timestamp}) AS ratings`,
		map[string]any{"userId": userID},
	)
	if err != nil {
		return nil, err
	}
	rec, err := res.Single(ctx)
	if err != nil {
		return nil, err
	}
	return rec.AsMap(), nil
}

// CreateActor creates a :Person:Actor node.
func CreateActor(ctx context.Context, p PersonInput) (map[string]any, error) {
	s := session(ctx)
	defer s.Close(ctx)
	res, err := s.Run(ctx,
		`CREATE (a:Person:Actor {
			name: $name, tmdbId: $tmdbId, born: date($born), died: $died,
			bornIn: $bornIn, url: $url, imdbId: $imdbId, bio: $bio, poster: $poster
		}) RETURN a`,
		map[string]any{
			"name": p.Name, "tmdbId": p.TmdbID, "born": p.Born, "died": p.Died,
			"bornIn": p.BornIn, "url": p.URL, "imdbId": p.ImdbID, "bio": p.Bio, "poster": p.Poster,
		},
	)
	if err != nil {
		return nil, err
	}
	rec, err := res.Single(ctx)
	if err != nil {
		return nil, err
	}
	return rec.AsMap(), nil
}

// CreateDirector creates a :Person:Director node.
func CreateDirector(ctx context.Context, p PersonInput) (map[string]any, error) {
	s := session(ctx)
	defer s.Close(ctx)
	res, err := s.Run(ctx,
		`CREATE (d:Person:Director {
			name: $name, tmdbId: $tmdbId, born: date($born), died: $died,
			bornIn: $bornIn, url: $url, imdbId: $imdbId, bio: $bio, poster: $poster
		}) RETURN d`,
		map[string]any{
			"name": p.Name, "tmdbId": p.TmdbID, "born": p.Born, "died": p.Died,
			"bornIn": p.BornIn, "url": p.URL, "imdbId": p.ImdbID, "bio": p.Bio, "poster": p.Poster,
		},
	)
	if err != nil {
		return nil, err
	}
	rec, err := res.Single(ctx)
	if err != nil {
		return nil, err
	}
	return rec.AsMap(), nil
}

// CreateActorDirector creates a :Person:Actor:Director node.
func CreateActorDirector(ctx context.Context, p PersonInput) (map[string]any, error) {
	s := session(ctx)
	defer s.Close(ctx)
	res, err := s.Run(ctx,
		`CREATE (p:Person:Actor:Director {
			name: $name, tmdbId: $tmdbId, born: date($born), died: $died,
			bornIn: $bornIn, url: $url, imdbId: $imdbId, bio: $bio, poster: $poster
		}) RETURN p`,
		map[string]any{
			"name": p.Name, "tmdbId": p.TmdbID, "born": p.Born, "died": p.Died,
			"bornIn": p.BornIn, "url": p.URL, "imdbId": p.ImdbID, "bio": p.Bio, "poster": p.Poster,
		},
	)
	if err != nil {
		return nil, err
	}
	rec, err := res.Single(ctx)
	if err != nil {
		return nil, err
	}
	return rec.AsMap(), nil
}

// CreateExtMovie creates a :Movie node with the full extended schema and an auto-generated movieId.
func CreateExtMovie(ctx context.Context, m ExtMovieInput) (map[string]any, error) {
	s := session(ctx)
	defer s.Close(ctx)
	res, err := s.Run(ctx,
		`CREATE (m:Movie {
			title: $title, tmdbId: $tmdbId, released: date($released), imdbRating: $imdbRating,
			movieId: $movieId, year: $year, imdbId: $imdbId, runtime: $runtime,
			countries: $countries, imdbVotes: $imdbVotes, url: $url, revenue: $revenue,
			plot: $plot, poster: $poster, budget: $budget, languages: $languages
		}) RETURN m`,
		map[string]any{
			"title": m.Title, "tmdbId": m.TmdbID, "released": m.Released, "imdbRating": m.ImdbRating,
			"movieId": rand.Intn(1_000_000), "year": m.Year, "imdbId": m.ImdbID, "runtime": m.Runtime,
			"countries": m.Countries, "imdbVotes": m.ImdbVotes, "url": m.URL, "revenue": m.Revenue,
			"plot": m.Plot, "poster": m.Poster, "budget": m.Budget, "languages": m.Languages,
		},
	)
	if err != nil {
		return nil, err
	}
	rec, err := res.Single(ctx)
	if err != nil {
		return nil, err
	}
	return rec.AsMap(), nil
}

// CreateGenre creates a :Genre node.
func CreateGenre(ctx context.Context, name string) (map[string]any, error) {
	s := session(ctx)
	defer s.Close(ctx)
	res, err := s.Run(ctx,
		`MERGE (g:Genre {name: $name}) RETURN g`,
		map[string]any{"name": name},
	)
	if err != nil {
		return nil, err
	}
	rec, err := res.Single(ctx)
	if err != nil {
		return nil, err
	}
	return rec.AsMap(), nil
}

// CreateActedIn creates an ACTED_IN relationship between a :Person:Actor and a :Movie.
func CreateActedIn(ctx context.Context, actorTmdbID, movieID int, role string) (map[string]any, error) {
	s := session(ctx)
	defer s.Close(ctx)
	res, err := s.Run(ctx,
		`MATCH (a:Person:Actor {tmdbId: $actorId}), (m:Movie {movieId: $movieId})
		 CREATE (a)-[r:ACTED_IN {role: $role}]->(m)
		 RETURN a.name AS actor, m.title AS movie, r.role AS role`,
		map[string]any{"actorId": actorTmdbID, "movieId": movieID, "role": role},
	)
	if err != nil {
		return nil, err
	}
	rec, err := res.Single(ctx)
	if err != nil {
		return nil, err
	}
	return rec.AsMap(), nil
}

// CreateDirected creates a DIRECTED relationship between a :Person:Director and a :Movie.
func CreateDirected(ctx context.Context, directorTmdbID, movieID int, role string) (map[string]any, error) {
	s := session(ctx)
	defer s.Close(ctx)
	res, err := s.Run(ctx,
		`MATCH (d:Person:Director {tmdbId: $directorId}), (m:Movie {movieId: $movieId})
		 CREATE (d)-[r:DIRECTED {role: $role}]->(m)
		 RETURN d.name AS director, m.title AS movie, r.role AS role`,
		map[string]any{"directorId": directorTmdbID, "movieId": movieID, "role": role},
	)
	if err != nil {
		return nil, err
	}
	rec, err := res.Single(ctx)
	if err != nil {
		return nil, err
	}
	return rec.AsMap(), nil
}

// CreateInGenre creates an IN_GENRE relationship between a :Movie and a :Genre.
func CreateInGenre(ctx context.Context, movieID int, genreName string) (map[string]any, error) {
	s := session(ctx)
	defer s.Close(ctx)
	res, err := s.Run(ctx,
		`MATCH (m:Movie {movieId: $movieId}), (g:Genre {name: $genreName})
		 CREATE (m)-[:IN_GENRE]->(g)
		 RETURN m.title AS movie, g.name AS genre`,
		map[string]any{"movieId": movieID, "genreName": genreName},
	)
	if err != nil {
		return nil, err
	}
	rec, err := res.Single(ctx)
	if err != nil {
		return nil, err
	}
	return rec.AsMap(), nil
}

// Seed functions

// seedSimple runs a simple 5 user + 5 movie seed where each user has 2 ratings
func seedSimple(ctx context.Context) (map[string]any, error) {
	s := session(ctx)
	defer s.Close(ctx)
	_, err := s.Run(ctx, `MATCH (n) WHERE n:User OR n:Movie DETACH DELETE n`, nil)
	if err != nil {
		return nil, err
	}

	uAlice, uBob, uCarlos, uDiana, uEthan := uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString(), uuid.NewString()
	mMatrix, mInception, mInterstellar, mDarkKnight, mParasite := rand.Intn(1_000_000), rand.Intn(1_000_000), rand.Intn(1_000_000), rand.Intn(1_000_000), rand.Intn(1_000_000)

	users := []struct{ name, id string }{
		{"Alice Johnson", uAlice},
		{"Bob Smith", uBob},
		{"Carlos Rivera", uCarlos},
		{"Diana Lee", uDiana},
		{"Ethan Brown", uEthan},
	}
	movies := []struct {
		title string
		id    int
		year  int
		plot  string
	}{
		{"The Matrix", mMatrix, 1999, "A hacker discovers reality is a simulation."},
		{"Inception", mInception, 2010, "A thief enters dreams to plant an idea."},
		{"Interstellar", mInterstellar, 2014, "Astronauts travel through a wormhole near Saturn."},
		{"The Dark Knight", mDarkKnight, 2008, "Batman faces the Joker in Gotham City."},
		{"Parasite", mParasite, 2019, "A poor family schemes to work for a wealthy household."},
	}

	for _, u := range users {
		if _, err := s.Run(ctx,
			`CREATE (u:User {name: $name, userId: $userId}) RETURN u`,
			map[string]any{"name": u.name, "userId": u.id},
		); err != nil {
			return nil, err
		}
	}
	for _, m := range movies {
		if _, err := s.Run(ctx,
			`CREATE (m:Movie {title: $title, movieId: $movieId, year: $year, plot: $plot}) RETURN m`,
			map[string]any{"title": m.title, "movieId": m.id, "year": m.year, "plot": m.plot},
		); err != nil {
			return nil, err
		}
	}

	ratings := []struct {
		uid    string
		mid    int
		rating int
	}{
		{uAlice, mMatrix, 5},
		{uAlice, mInception, 4},
		{uBob, mInception, 5},
		{uBob, mInterstellar, 3},
		{uCarlos, mMatrix, 4},
		{uCarlos, mDarkKnight, 5},
		{uDiana, mInterstellar, 4},
		{uDiana, mParasite, 5},
		{uEthan, mDarkKnight, 3},
		{uEthan, mParasite, 4},
		{uEthan, mMatrix, 5},
	}
	for _, r := range ratings {
		if _, err := CreateRating(ctx, r.uid, r.mid, r.rating, time.Now().Unix()); err != nil {
			return nil, err
		}
	}

	return map[string]any{"seeded": "5 users, 5 movies, 11 ratings"}, nil
}

// seedExtended executes a more complex seed, where we include Person, Actor, Director, Movie, Genre, etc. relations in the graph
func seedExtended(ctx context.Context) (map[string]any, error) {
	s := session(ctx)
	_, err := s.Run(ctx, `MATCH (n) DETACH DELETE n`, nil)
	s.Close(ctx)
	if err != nil {
		return nil, err
	}

	// Genres
	for _, g := range []string{"Action", "Sci-Fi", "Drama", "Thriller", "Western"} {
		if _, err := CreateGenre(ctx, g); err != nil {
			return nil, err
		}
	}

	// Movies — IDs generated here so we can wire relationships below
	mMatrix := rand.Intn(1_000_000)
	mInception := rand.Intn(1_000_000)
	mParasite := rand.Intn(1_000_000)
	mUnforgiven := rand.Intn(1_000_000)

	movies := []struct {
		id int
		m  ExtMovieInput
	}{
		{mMatrix, ExtMovieInput{
			Title: "The Matrix", TmdbID: 603, Released: "1999-03-31", ImdbRating: 8.7,
			Year: 1999, ImdbID: 133093, Runtime: 136, Countries: []string{"USA", "Australia"},
			ImdbVotes: 1800000, URL: "https://www.themoviedb.org/movie/603", Revenue: 463517383,
			Plot:   "A hacker discovers reality is a simulation.",
			Poster: "https://image.tmdb.org/t/p/w500/f89U3ADr1oiB1s9GkdPOEpXUk5H.jpg",
			Budget: 63000000, Languages: []string{"English"},
		}},
		{mInception, ExtMovieInput{
			Title: "Inception", TmdbID: 27205, Released: "2010-07-16", ImdbRating: 8.8,
			Year: 2010, ImdbID: 1375666, Runtime: 148, Countries: []string{"USA", "UK"},
			ImdbVotes: 2300000, URL: "https://www.themoviedb.org/movie/27205", Revenue: 836836967,
			Plot:   "A thief enters dreams to plant an idea.",
			Poster: "https://image.tmdb.org/t/p/w500/9gk7adHYeDvHkCSEqAvQNLV5Uge.jpg",
			Budget: 160000000, Languages: []string{"English", "Japanese", "French"},
		}},
		{mParasite, ExtMovieInput{
			Title: "Parasite", TmdbID: 496243, Released: "2019-05-30", ImdbRating: 8.5,
			Year: 2019, ImdbID: 6751668, Runtime: 132, Countries: []string{"South Korea"},
			ImdbVotes: 780000, URL: "https://www.themoviedb.org/movie/496243", Revenue: 258774021,
			Plot:   "A poor family schemes to infiltrate and leech off a wealthy household.",
			Poster: "https://image.tmdb.org/t/p/w500/7IiTTgloROVKhJs96xMhEBhCSRu.jpg",
			Budget: 11400000, Languages: []string{"Korean", "English"},
		}},
		{mUnforgiven, ExtMovieInput{
			Title: "Unforgiven", TmdbID: 1424, Released: "1992-08-07", ImdbRating: 8.2,
			Year: 1992, ImdbID: 105695, Runtime: 130, Countries: []string{"USA"},
			ImdbVotes: 430000, URL: "https://www.themoviedb.org/movie/1424", Revenue: 159157447,
			Plot:   "A retired outlaw takes on one last job to collect a bounty.",
			Poster: "https://image.tmdb.org/t/p/w500/lMB1JnkEBVDJFNdCMRH6E6X7pId.jpg",
			Budget: 14400000, Languages: []string{"English"},
		}},
	}
	sm2 := session(ctx)
	defer sm2.Close(ctx)
	for _, em := range movies {
		m := em.m
		if _, err := sm2.Run(ctx,
			`CREATE (m:Movie {
				title: $title, tmdbId: $tmdbId, released: date($released), imdbRating: $imdbRating,
				movieId: $movieId, year: $year, imdbId: $imdbId, runtime: $runtime,
				countries: $countries, imdbVotes: $imdbVotes, url: $url, revenue: $revenue,
				plot: $plot, poster: $poster, budget: $budget, languages: $languages
			}) RETURN m`,
			map[string]any{
				"title": m.Title, "tmdbId": m.TmdbID, "released": m.Released, "imdbRating": m.ImdbRating,
				"movieId": em.id, "year": m.Year, "imdbId": m.ImdbID, "runtime": m.Runtime,
				"countries": m.Countries, "imdbVotes": m.ImdbVotes, "url": m.URL, "revenue": m.Revenue,
				"plot": m.Plot, "poster": m.Poster, "budget": m.Budget, "languages": m.Languages,
			},
		); err != nil {
			return nil, err
		}
	}

	// Pure actors
	if _, err := CreateActor(ctx, PersonInput{
		Name: "Keanu Reeves", TmdbID: 6384, Born: "1964-09-02", BornIn: "Beirut",
		URL: "https://www.themoviedb.org/person/6384", ImdbID: 206,
		Bio:    "Canadian actor known for action and sci-fi roles.",
		Poster: "https://image.tmdb.org/t/p/w500/4D0PpNI0kmP58hgrwGC3wCjxhnm.jpg",
	}); err != nil {
		return nil, err
	}
	if _, err := CreateActor(ctx, PersonInput{
		Name: "Leonardo DiCaprio", TmdbID: 6193, Born: "1974-11-11", BornIn: "Los Angeles",
		URL: "https://www.themoviedb.org/person/6193", ImdbID: 138,
		Bio:    "American actor and film producer.",
		Poster: "https://image.tmdb.org/t/p/w500/wo2hJpn04vbtmh0B9utCFdsQhxM.jpg",
	}); err != nil {
		return nil, err
	}
	if _, err := CreateActor(ctx, PersonInput{
		Name: "Song Kang-ho", TmdbID: 21685, Born: "1967-01-17", BornIn: "Gimhae",
		URL: "https://www.themoviedb.org/person/21685", ImdbID: 490071,
		Bio:    "South Korean actor, one of the most acclaimed in Korean cinema.",
		Poster: "https://image.tmdb.org/t/p/w500/vu1ElHGDCHFT5HWNB1BYTW3QGNT.jpg",
	}); err != nil {
		return nil, err
	}

	// Pure directors
	if _, err := CreateDirector(ctx, PersonInput{
		Name: "Lana Wachowski", TmdbID: 9340, Born: "1965-06-21", BornIn: "Chicago",
		URL: "https://www.themoviedb.org/person/9340", ImdbID: 905154,
		Bio:    "American film director and screenwriter.",
		Poster: "https://image.tmdb.org/t/p/w500/cGSMBrRdQG3h2hbZNgLekzBnCjy.jpg",
	}); err != nil {
		return nil, err
	}
	if _, err := CreateDirector(ctx, PersonInput{
		Name: "Christopher Nolan", TmdbID: 525, Born: "1970-07-30", BornIn: "London",
		URL: "https://www.themoviedb.org/person/525", ImdbID: 634240,
		Bio:    "British-American director known for complex, large-scale films.",
		Poster: "https://image.tmdb.org/t/p/w500/xuAIuYSmsUhAq3Cv3eoWA2DBq45.jpg",
	}); err != nil {
		return nil, err
	}
	if _, err := CreateDirector(ctx, PersonInput{
		Name: "Bong Joon-ho", TmdbID: 21684, Born: "1969-09-14", BornIn: "Daegu",
		URL: "https://www.themoviedb.org/person/21684", ImdbID: 349752,
		Bio:    "South Korean director who became the first Asian to win the Palme d'Or.",
		Poster: "https://image.tmdb.org/t/p/w500/oRKnHCwHnvEuHCHKnbqkBQr0bB1.jpg",
	}); err != nil {
		return nil, err
	}

	// Clint Eastwood — both actor and director on Unforgiven
	if _, err := CreateActorDirector(ctx, PersonInput{
		Name: "Clint Eastwood", TmdbID: 190, Born: "1930-05-31", BornIn: "San Francisco",
		URL: "https://www.themoviedb.org/person/190", ImdbID: 441452,
		Bio:    "American actor and filmmaker, iconic for his Western and anti-hero roles.",
		Poster: "https://image.tmdb.org/t/p/w500/3XOGp6WdEFCXxKGXrGpLZtNPFjm.jpg",
	}); err != nil {
		return nil, err
	}

	// ACTED_IN
	if _, err := CreateActedIn(ctx, 6384, mMatrix, "Neo"); err != nil {
		return nil, err
	}
	if _, err := CreateActedIn(ctx, 6193, mInception, "Dom Cobb"); err != nil {
		return nil, err
	}
	if _, err := CreateActedIn(ctx, 21685, mParasite, "Ki-taek"); err != nil {
		return nil, err
	}
	if _, err := CreateActedIn(ctx, 190, mUnforgiven, "William Munny"); err != nil {
		return nil, err
	}

	// DIRECTED
	if _, err := CreateDirected(ctx, 9340, mMatrix, "Director"); err != nil {
		return nil, err
	}
	if _, err := CreateDirected(ctx, 525, mInception, "Director"); err != nil {
		return nil, err
	}
	if _, err := CreateDirected(ctx, 21684, mParasite, "Director"); err != nil {
		return nil, err
	}
	if _, err := CreateDirected(ctx, 190, mUnforgiven, "Director"); err != nil {
		return nil, err
	}

	// IN_GENRE
	for _, g := range []struct {
		id    int
		genre string
	}{
		{mMatrix, "Action"},
		{mMatrix, "Sci-Fi"},
		{mInception, "Action"},
		{mInception, "Sci-Fi"},
		{mInception, "Thriller"},
		{mParasite, "Drama"},
		{mParasite, "Thriller"},
		{mUnforgiven, "Western"},
		{mUnforgiven, "Drama"},
	} {
		if _, err := CreateInGenre(ctx, g.id, g.genre); err != nil {
			return nil, err
		}
	}

	// Users, generate IDs upfront so we can wire ratings below
	uMaria, uJames, uSophia := uuid.NewString(), uuid.NewString(), uuid.NewString()
	sm := session(ctx)
	defer sm.Close(ctx)
	for _, u := range []struct{ id, name string }{
		{uMaria, "Maria Garcia"}, {uJames, "James Wilson"}, {uSophia, "Sophia Chen"},
	} {
		if _, err := sm.Run(ctx,
			`CREATE (u:User {name: $name, userId: $userId})`,
			map[string]any{"name": u.name, "userId": u.id},
		); err != nil {
			return nil, err
		}
	}

	// Ratings, varied scores, not all 5s
	for _, r := range []struct {
		uid    string
		mid    int
		rating int
	}{
		{uMaria, mMatrix, 5},
		{uMaria, mInception, 3},
		{uMaria, mUnforgiven, 4},
		{uJames, mInception, 5},
		{uJames, mParasite, 4},
		{uSophia, mMatrix, 2},
		{uSophia, mParasite, 5},
		{uSophia, mUnforgiven, 4},
	} {
		if _, err := CreateRating(ctx, r.uid, r.mid, r.rating, time.Now().Unix()); err != nil {
			return nil, err
		}
	}

	return map[string]any{"seeded": "extended graph: 4 movies, 3 actors, 3 directors, 1 actor-director, 3 users, 8 ratings"}, nil
}
