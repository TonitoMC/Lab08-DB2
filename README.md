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

## Neo4j queries (`queries.go`)

All Cypher lives in `queries.go`. The functions below are what actually talk to the database.

### Creating nodes

**Users**
```cypher
CREATE (u:User {name: $name, userId: $userId}) RETURN u
```

**Movies**
```cypher
CREATE (m:Movie {title: $title, movieId: $movieId, year: $year, plot: $plot}) RETURN m
```

**Actors, Directors, and Actor-Directors**
```cypher
CREATE (a:Person:Actor { name: $name, tmdbId: $tmdbId, born: date($born), ... }) RETURN a
CREATE (d:Person:Director { name: $name, tmdbId: $tmdbId, born: date($born), ... }) RETURN d
CREATE (p:Person:Actor:Director { name: $name, tmdbId: $tmdbId, born: date($born), ... }) RETURN p
```

**Genres** — uses `MERGE` so running it twice won't create duplicates
```cypher
MERGE (g:Genre {name: $name}) RETURN g
```

**Extended movies** — full schema with ratings, runtime, countries, etc.
```cypher
CREATE (m:Movie {
    title: $title, tmdbId: $tmdbId, released: date($released), imdbRating: $imdbRating,
    movieId: $movieId, year: $year, runtime: $runtime, countries: $countries,
    revenue: $revenue, budget: $budget, languages: $languages, ...
}) RETURN m
```

---

### Creating relationships

**User rates a movie**
```cypher
MATCH (u:User {userId: $userId}), (m:Movie {movieId: $movieId})
CREATE (u)-[r:RATED {rating: $rating, timestamp: $timestamp}]->(m)
RETURN u.name AS userName, m.title AS movieTitle, r.rating AS rating, r.timestamp AS timestamp
```

**Actor played a role in a movie**
```cypher
MATCH (a:Person:Actor {tmdbId: $actorId}), (m:Movie {movieId: $movieId})
CREATE (a)-[r:ACTED_IN {role: $role}]->(m)
RETURN a.name AS actor, m.title AS movie, r.role AS role
```

**Director directed a movie**
```cypher
MATCH (d:Person:Director {tmdbId: $directorId}), (m:Movie {movieId: $movieId})
CREATE (d)-[r:DIRECTED {role: $role}]->(m)
RETURN d.name AS director, m.title AS movie, r.role AS role
```

**Movie belongs to a genre**
```cypher
MATCH (m:Movie {movieId: $movieId}), (g:Genre {name: $genreName})
CREATE (m)-[:IN_GENRE]->(g)
RETURN m.title AS movie, g.name AS genre
```

---

### Finding things

**Find a user by name**
```cypher
MATCH (u:User {name: $name}) RETURN u
```

**Find a user and everything they've rated**
```cypher
MATCH (u:User {name: $name})
OPTIONAL MATCH (u)-[r:RATED]->(m:Movie)
RETURN u.name AS name, u.userId AS userId,
       collect({title: m.title, movieId: m.movieId, rating: r.rating, timestamp: r.timestamp}) AS ratings
```

**Find a specific rating a user gave to a movie**
```cypher
MATCH (u:User {name: $name})-[r:RATED]->(m:Movie {title: $title})
RETURN u.name AS userName, m.title AS movieTitle, r.rating AS rating, r.timestamp AS timestamp
```

**Find a movie by title**
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
