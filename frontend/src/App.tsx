import { useState, type ChangeEvent } from 'react'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'

const API = '/api'

async function api(method: string, path: string, body?: unknown) {
  const res = await fetch(API + path, {
    method,
    headers: body ? { 'Content-Type': 'application/json' } : {},
    body: body ? JSON.stringify(body) : undefined,
  })
  return res.json()
}

function ResultBox({ data }: { data: unknown }) {
  if (!data) return null
  return (
    <pre className="mt-3 rounded-md bg-muted p-3 text-xs overflow-auto max-h-64 text-left">
      {JSON.stringify(data, null, 2)}
    </pre>
  )
}

function Field({
  label,
  value,
  onChange,
  placeholder,
  type = 'text',
}: {
  label: string
  value: string
  onChange: (v: string) => void
  placeholder?: string
  type?: string
}) {
  return (
    <div className="grid gap-1">
      <Label>{label}</Label>
      <Input
        type={type}
        value={value}
        onChange={(e: ChangeEvent<HTMLInputElement>) => onChange(e.target.value)}
        placeholder={placeholder}
      />
    </div>
  )
}

// ── Seed ──────────────────────────────────────────────────────────────────────

function SeedSimple() {
  const [result, setResult] = useState<unknown>(null)
  const [loading, setLoading] = useState(false)

  async function run() {
    setLoading(true)
    setResult(await api('POST', '/seed'))
    setLoading(false)
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Seed Simple Graph</CardTitle>
        <p className="text-sm text-muted-foreground">
          Creates 5 users + 5 movies + 11 RATED relationships
        </p>
      </CardHeader>
      <CardContent>
        <Button onClick={run} disabled={loading}>
          {loading ? 'Seeding…' : 'Run Seed'}
        </Button>
        <ResultBox data={result} />
      </CardContent>
    </Card>
  )
}

function SeedExtended() {
  const [result, setResult] = useState<unknown>(null)
  const [loading, setLoading] = useState(false)

  async function run() {
    setLoading(true)
    setResult(await api('POST', '/seed-extended'))
    setLoading(false)
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Seed Extended Graph</CardTitle>
        <p className="text-sm text-muted-foreground">
          Wipes DB then creates Person:Actor, Person:Director, Movie (full schema), Genre + all relationships (inciso 4)
        </p>
      </CardHeader>
      <CardContent>
        <Button onClick={run} disabled={loading} variant="secondary">
          {loading ? 'Seeding…' : 'Run Extended Seed'}
        </Button>
        <ResultBox data={result} />
      </CardContent>
    </Card>
  )
}

// ── Simple graph ──────────────────────────────────────────────────────────────

function CreateUser() {
  const [name, setName] = useState('')
  const [result, setResult] = useState<unknown>(null)

  async function submit() {
    setResult(await api('POST', '/users', { name }))
  }

  return (
    <Card>
      <CardHeader><CardTitle className="text-base">Create User</CardTitle></CardHeader>
      <CardContent className="grid gap-3">
        <Field label="Name" value={name} onChange={setName} placeholder="Alice Johnson" />
        <Button onClick={submit} disabled={!name}>Create</Button>
        <ResultBox data={result} />
      </CardContent>
    </Card>
  )
}

function CreateMovie() {
  const [title, setTitle] = useState('')
  const [year, setYear] = useState('')
  const [plot, setPlot] = useState('')
  const [result, setResult] = useState<unknown>(null)

  async function submit() {
    setResult(await api('POST', '/movies', { title, year: Number(year), plot }))
  }

  return (
    <Card>
      <CardHeader><CardTitle className="text-base">Create Movie</CardTitle></CardHeader>
      <CardContent className="grid gap-3">
        <Field label="Title" value={title} onChange={setTitle} placeholder="The Matrix" />
        <Field label="Year" value={year} onChange={setYear} placeholder="1999" type="number" />
        <Field label="Plot" value={plot} onChange={setPlot} placeholder="A hacker discovers…" />
        <Button onClick={submit} disabled={!title}>Create</Button>
        <ResultBox data={result} />
      </CardContent>
    </Card>
  )
}

function CreateRating() {
  const [userId, setUserId] = useState('')
  const [movieId, setMovieId] = useState('')
  const [rating, setRating] = useState('')
  const [result, setResult] = useState<unknown>(null)

  async function submit() {
    setResult(await api('POST', '/ratings', { userId, movieId: Number(movieId), rating: Number(rating) }))
  }

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Create Rating</CardTitle>
        <p className="text-sm text-muted-foreground">Creates a RATED relationship between a user and a movie</p>
      </CardHeader>
      <CardContent className="grid gap-3">
        <Field label="User ID" value={userId} onChange={setUserId} placeholder="u1" />
        <Field label="Movie ID" value={movieId} onChange={setMovieId} placeholder="1" type="number" />
        <Field label="Rating (0-5)" value={rating} onChange={setRating} placeholder="5" type="number" />
        <Button onClick={submit} disabled={!userId || !movieId || !rating}>Rate</Button>
        <ResultBox data={result} />
      </CardContent>
    </Card>
  )
}

// ── Find ──────────────────────────────────────────────────────────────────────


function FindUserByName() {
  const [name, setName] = useState('')
  const [result, setResult] = useState<unknown>(null)

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Find User by Name</CardTitle>
        <p className="text-sm text-muted-foreground">Lookup a user or fetch their rated movies</p>
      </CardHeader>
      <CardContent className="grid gap-3">
        <Field label="Name" value={name} onChange={setName} placeholder="Alice Johnson" />
        <div className="flex gap-2">
          <Button onClick={() => api('GET', `/users/search?name=${encodeURIComponent(name)}`).then(setResult)} disabled={!name} variant="outline">
            Find User
          </Button>
          <Button onClick={() => api('GET', `/users/search?name=${encodeURIComponent(name)}&ratings=true`).then(setResult)} disabled={!name}>
            User + Ratings
          </Button>
        </div>
        <ResultBox data={result} />
      </CardContent>
    </Card>
  )
}


function FindMovieByTitle() {
  const [title, setTitle] = useState('')
  const [result, setResult] = useState<unknown>(null)

  return (
    <Card>
      <CardHeader><CardTitle className="text-base">Find Movie by Title</CardTitle></CardHeader>
      <CardContent className="grid gap-3">
        <Field label="Title" value={title} onChange={setTitle} placeholder="The Matrix" />
        <Button onClick={() => api('GET', `/movies/search?title=${encodeURIComponent(title)}`).then(setResult)} disabled={!title} variant="outline">
          Search
        </Button>
        <ResultBox data={result} />
      </CardContent>
    </Card>
  )
}

function FindRating() {
  const [userName, setUserName] = useState('')
  const [movieTitle, setMovieTitle] = useState('')
  const [result, setResult] = useState<unknown>(null)

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Find Rating</CardTitle>
        <p className="text-sm text-muted-foreground">Look up a user's rating for a specific movie</p>
      </CardHeader>
      <CardContent className="grid gap-3">
        <Field label="User Name" value={userName} onChange={setUserName} placeholder="Alice Johnson" />
        <Field label="Movie Title" value={movieTitle} onChange={setMovieTitle} placeholder="The Matrix" />
        <Button
          onClick={() => api('GET', `/ratings/search?user=${encodeURIComponent(userName)}&movie=${encodeURIComponent(movieTitle)}`).then(setResult)}
          disabled={!userName || !movieTitle}
          variant="outline"
        >
          Find Rating
        </Button>
        <ResultBox data={result} />
      </CardContent>
    </Card>
  )
}

// ── App ───────────────────────────────────────────────────────────────────────

export default function App() {
  return (
    <div className="min-h-screen bg-background p-6">
      <div className="max-w-3xl mx-auto space-y-4">
        <div>
          <h1 className="text-2xl font-semibold">The Nodes World Cup</h1>
          <p className="text-sm text-muted-foreground">Lab 08 — Neo4j Graph Backend</p>
        </div>

        <Separator />

        <Tabs defaultValue="seed">
          <TabsList>
            <TabsTrigger value="seed">Seed</TabsTrigger>
            <TabsTrigger value="users">Users</TabsTrigger>
            <TabsTrigger value="movies">Movies</TabsTrigger>
            <TabsTrigger value="ratings">Ratings</TabsTrigger>
            <TabsTrigger value="find">Find</TabsTrigger>
          </TabsList>

          <TabsContent value="seed" className="space-y-4 mt-4">
            <div className="flex gap-2">
              <Badge variant="outline">Inciso 1 &amp; 2</Badge>
              <Badge variant="outline">Inciso 4</Badge>
            </div>
            <SeedSimple />
            <SeedExtended />
          </TabsContent>

          <TabsContent value="users" className="mt-4 space-y-4">
            <Badge variant="outline">Inciso 1</Badge>
            <CreateUser />
          </TabsContent>

          <TabsContent value="movies" className="mt-4 space-y-4">
            <Badge variant="outline">Inciso 1</Badge>
            <CreateMovie />
          </TabsContent>

          <TabsContent value="ratings" className="mt-4 space-y-4">
            <Badge variant="outline">Inciso 1</Badge>
            <CreateRating />
          </TabsContent>

          <TabsContent value="find" className="mt-4 space-y-4">
            <Badge variant="outline">Inciso 3</Badge>
            <FindUserByName />
            <FindMovieByTitle />
            <FindRating />
          </TabsContent>

        </Tabs>
      </div>
    </div>
  )
}
