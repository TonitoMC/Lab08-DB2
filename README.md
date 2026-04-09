# Lab 08 – The Nodes World Cup Part I

Go + Neo4j AuraDB graph database lab.

## Running

**Backend** (port 8080):
```bash
go run .
```

**Frontend** (port 5173):
```bash
cd frontend && npm run dev
```

Requires a `.env` in the root with `NEO4J_URI`, `NEO4J_USERNAME`, `NEO4J_PASSWORD`, `NEO4J_DATABASE`.

---

## Queries (`queries.go`)

### Creating nodes

`CreateUser(name)`
```cypher
CREATE (u:User {name: $name, userId: $userId}) RETURN u
```

`CreateMovie(title, year, plot)`
```cypher
CREATE (m:Movie {title: $title, movieId: $movieId, year: $year, plot: $plot}) RETURN m
```

`CreateActor(PersonInput)`
```cypher
CREATE (a:Person:Actor {name: $name, tmdbId: $tmdbId, born: date($born), bornIn: $bornIn, ...}) RETURN a
```

`CreateDirector(PersonInput)`
```cypher
CREATE (d:Person:Director {name: $name, tmdbId: $tmdbId, born: date($born), bornIn: $bornIn, ...}) RETURN d
```

`CreateActorDirector(PersonInput)`
```cypher
CREATE (p:Person:Actor:Director {name: $name, tmdbId: $tmdbId, born: date($born), bornIn: $bornIn, ...}) RETURN p
```

`CreateExtMovie(ExtMovieInput)`
```cypher
CREATE (m:Movie {
    title: $title, tmdbId: $tmdbId, released: date($released), imdbRating: $imdbRating,
    movieId: $movieId, year: $year, runtime: $runtime, countries: $countries,
    revenue: $revenue, budget: $budget, languages: $languages, ...
}) RETURN m
```

`CreateGenre(name)`
```cypher
MERGE (g:Genre {name: $name}) RETURN g
```

---

### Creating relationships

`CreateRating(userID, movieID, rating, timestamp)`
```cypher
MATCH (u:User {userId: $userId}), (m:Movie {movieId: $movieId})
CREATE (u)-[r:RATED {rating: $rating, timestamp: $timestamp}]->(m)
RETURN u.name AS userName, m.title AS movieTitle, r.rating AS rating, r.timestamp AS timestamp
```

`CreateActedIn(actorTmdbID, movieID, role)`
```cypher
MATCH (a:Person:Actor {tmdbId: $actorId}), (m:Movie {movieId: $movieId})
CREATE (a)-[r:ACTED_IN {role: $role}]->(m)
RETURN a.name AS actor, m.title AS movie, r.role AS role
```

`CreateDirected(directorTmdbID, movieID, role)`
```cypher
MATCH (d:Person:Director {tmdbId: $directorId}), (m:Movie {movieId: $movieId})
CREATE (d)-[r:DIRECTED {role: $role}]->(m)
RETURN d.name AS director, m.title AS movie, r.role AS role
```

`CreateInGenre(movieID, genreName)`
```cypher
MATCH (m:Movie {movieId: $movieId}), (g:Genre {name: $genreName})
CREATE (m)-[:IN_GENRE]->(g)
RETURN m.title AS movie, g.name AS genre
```

---

### Finding nodes and relationships

`FindUserByName(name)`
```cypher
MATCH (u:User {name: $name}) RETURN u
```

`FindUserWithRatingsByName(name)`
```cypher
MATCH (u:User {name: $name})
OPTIONAL MATCH (u)-[r:RATED]->(m:Movie)
RETURN u.name AS name, u.userId AS userId,
       collect({title: m.title, movieId: m.movieId, rating: r.rating, timestamp: r.timestamp}) AS ratings
```

`FindUserRatingForMovie(userName, movieTitle)`
```cypher
MATCH (u:User {name: $name})-[r:RATED]->(m:Movie {title: $title})
RETURN u.name AS userName, m.title AS movieTitle, r.rating AS rating, r.timestamp AS timestamp
```

`FindMovieByTitle(title)`
```cypher
MATCH (m:Movie {title: $title}) RETURN m
```

---

## API endpoints

| Method | Path | Description |
|---|---|---|
| `POST` | `/api/seed` | Seed simple graph |
| `POST` | `/api/seed-extended` | Seed extended graph |
| `POST` | `/api/users` | Create user |
| `GET` | `/api/users/search?name=` | Find user by name (`&ratings=true` to include ratings) |
| `POST` | `/api/movies` | Create movie |
| `GET` | `/api/movies/search?title=` | Find movie by title |
| `POST` | `/api/ratings` | Create rating |
| `GET` | `/api/ratings/search?user=&movie=` | Find a user's rating for a specific movie |
| `POST` | `/api/actors` | Create actor |
| `POST` | `/api/actors/{tmdbId}/acted-in` | Create `ACTED_IN` relationship |
| `POST` | `/api/directors` | Create director |
| `POST` | `/api/directors/{tmdbId}/directed` | Create `DIRECTED` relationship |
| `POST` | `/api/genres` | Create genre |
| `POST` | `/api/ext-movies` | Create extended movie |
| `POST` | `/api/movies/{movieId}/genre` | Create `IN_GENRE` relationship |
