<p align="center">
  <h1 align="center" style="font-weight: bold;">
    S3tfy (Shortify)
  </h1>
</p>


A Go URL shortener API with user registration, JWT authentication, refresh tokens, protected URL management routes, and public redirect routes.

The project uses:

- Go standard `net/http`
- `chi` for routing
- PostgreSQL with `pgxpool`
- JWT access tokens
- PostgreSQL-backed refresh tokens
- `golang-migrate` SQL migrations
- A Redis cache helper package, currently prepared but not wired into the main request flow

## Project Structure

```text
cmd/server/main.go              Application entrypoint, dependency wiring, and route registration
internal/auth/                  Password hashing, JWT creation, token validation, login, refresh logic
internal/cache/                 Redis helper wrapper for Set/Get/Delete operations
internal/db/                    PostgreSQL connection pool setup
internal/handlers/              HTTP handlers for auth, users, and URL shortening
internal/middleware/            JWT authentication middleware
internal/models/                Data models used by repositories and handlers
internal/repository/            PostgreSQL data access layer
internal/shortener/             URL code generation and shortener service logic
migrations/                     SQL migrations for users, urls, clicks, and refresh tokens
Makefile                        Migration helper commands
```

## Environment Variables

Create a `.env` file in the project root.

```env
PORT=8080
HOST_URL=http://localhost:8080
DATABASE_URL=postgres://user:password@localhost:5432/url_shortener?sslmode=disable
MIGRATIONS_DIR=migrations
REDIS_URL=redis://localhost:6379/0
```

Notes:

- `PORT` is optional. The server defaults to `8080`.
- `HOST_URL` is used to build full shortened URLs like `http://localhost:8080/abc123`.
- `DATABASE_URL` is required for PostgreSQL.
- `REDIS_URL` is only used by the Redis helper package right now. Redis is not yet active in the main server flow.
- The JWT secret is generated on every server start in `cmd/server/main.go`. This means existing access tokens become invalid after a restart.

## Running The Project

Install dependencies:

```bash
go mod download
```

Run migrations:

```bash
make mp
```

Start the server:

```bash
go run ./cmd/server
```

The server starts on:

```text
http://localhost:8080
```

unless `PORT` is set to another value.

## Database Tables

### `users`

Stores application users.

Important columns:

- `id`
- `email`
- `full_name`
- `password_hash`
- `created_at`
- `updated_at`
- `last_login`

### `urls`

Stores each shortened URL owned by a user.

Important columns:

- `id`
- `user_id`
- `shortened_url_code`
- `original_url`
- `created_at`

### `clicks`

Exists in the migration and is intended for tracking visits to short URLs.

Important columns:

- `id`
- `url_id`
- `clicked_at`
- `ip_address`

This table is not currently written to by the Go redirect handler.

### `refresh_tokens`

Stores refresh tokens for authenticated sessions.

Important columns:

- `id`
- `user_id`
- `token`
- `expires_at`
- `created_at`
- `revoked`

## API Routes

### Public Routes

These routes do not require an access token.

### `POST /api/v1/auth/register`

Creates a new user.

Request body:

```json
{
  "email": "ada@example.com",
  "fullName": "Ada Lovelace",
  "password": "password123"
}
```

Response:

```json
{
  "id": "user-uuid",
  "fullName": "Ada Lovelace",
  "email": "ada@example.com"
}
```

Handler: `internal/handlers/auth.go`

Flow:

1. Decodes the JSON request body.
2. Checks that email and password are present.
3. Performs a basic email validation by checking for `@`.
4. Calls `AuthService.Register`.
5. `AuthService` checks for an existing user, hashes the password, and creates the user through `UserRepository`.

Possible responses:

- `201 Created` when registration succeeds.
- `400 Bad Request` for invalid JSON or missing/invalid fields.
- `409 Conflict` when the email is already in use.
- `500 Internal Server Error` for unexpected database or service errors.

### `POST /api/v1/auth/login`

Authenticates a user and returns an access token plus a refresh token.

Request body:

```json
{
  "email": "ada@example.com",
  "password": "password123"
}
```

Response:

```json
{
  "token": "jwt-access-token",
  "refreshToken": "refresh-token"
}
```

Handler: `internal/handlers/auth.go`

Flow:

1. Decodes the JSON request body.
2. Calls `AuthService.LoginWithRefresh`.
3. `AuthService` finds the user by email.
4. The password is verified with the stored password hash.
5. A JWT access token is generated.
6. A refresh token is created and saved in PostgreSQL.

Possible responses:

- `200 OK` when login succeeds.
- `400 Bad Request` for invalid JSON.
- `401 Unauthorized` for invalid credentials.
- `500 Internal Server Error` for unexpected errors.

### `POST /api/v1/auth/refresh`

Uses a refresh token to generate a new access token.

Request body:

```json
{
  "refreshToken": "refresh-token"
}
```

Response:

```json
{
  "accessToken": "new-jwt-access-token"
}
```

Handler: `internal/handlers/auth.go`

Flow:

1. Decodes the JSON request body.
2. Calls `AuthService.RefreshAccessToken`.
3. Finds the refresh token in PostgreSQL.
4. Rejects revoked or expired tokens.
5. Finds the owning user.
6. Generates a new JWT access token.

Possible responses:

- `200 OK` when refresh succeeds.
- `400 Bad Request` for invalid JSON.
- `401 Unauthorized` for invalid or expired refresh tokens.
- `500 Internal Server Error` for unexpected errors.

### `GET /api/v1/url{code}`

Redirects a short URL code to the original long URL.

Path parameter:

```text
/api/v1/url/{code}
```

Example:

```text
GET /api/v1/url/abc123
```

Handler: `internal/handlers/shortener.go`

Flow:

1. Reads the dynamic `{code}` value from the URL path using `chi.URLParam`.
2. Calls `ShortenerService.FindOriginalUrlPublic`.
3. Looks up the original URL by short code without checking user ownership.
4. Redirects to the original URL with `302 Found`.

Possible responses:

- `302 Found` when the code exists.
- `404 Not Found` when the code is missing or unknown.

## Protected Routes

These routes require a JWT access token.

Send the token with every protected request:

```http
Authorization: Bearer jwt-access-token
```

The auth middleware validates the token, extracts the `sub` claim, parses it as a user UUID, and stores it in the request context.

### `GET /api/v1/user`

Returns the authenticated user's profile.

Request body: none.

Response:

```json
{
  "id": "user-uuid",
  "email": "ada@example.com",
  "fullName": "Ada Lovelace"
}
```

Handler: `internal/handlers/user.go`

Flow:

1. Gets the authenticated user ID from request context.
2. Calls `UserRepository.FindUserByID`.
3. Returns public profile fields.

Possible responses:

- `200 OK` when the user exists.
- `401 Unauthorized` when the token is missing or invalid.
- `404 Not Found` when the user cannot be found.

### `POST /api/v1/shorten/create`

Creates and saves a short URL for the authenticated user.

Request body:

```json
{
  "originalUrl": "https://example.com/articles/my-long-url"
}
```

Response:

```json
{
  "shortenedUrl": "http://localhost:8080/abc123"
}
```

Handler: `internal/handlers/shortener.go`

Flow:

1. Gets the authenticated user ID from request context.
2. Decodes the JSON request body.
3. Validates that `originalUrl` is present.
4. Calls `ShortenerService.SaveUrls`.
5. The service generates a 6-character short code.
6. `ShortenerRepository.SaveUrls` stores the original URL, code, and user ID in PostgreSQL.
7. The service builds the full shortened URL using `HOST_URL`.

Possible responses:

- `201 Created` when the URL is shortened.
- `400 Bad Request` for invalid JSON or missing `originalUrl`.
- `401 Unauthorized` when the token is missing or invalid.
- `404 Not Found` when saving fails.

### `GET /api/v1/shorten/original-url`

Finds the original URL for a short code or full shortened URL owned by the authenticated user.

Query parameters:

```text
code=abc123
```

or:

```text
shortenedUrl=http://localhost:8080/abc123
```

Example:

```text
GET /api/v1/shorten/original-url?code=abc123
```

Response:

```json
{
  "originalUrl": "https://example.com/articles/my-long-url"
}
```

Handler: `internal/handlers/shortener.go`

Flow:

1. Gets the authenticated user ID from request context.
2. Reads `code` from the query string.
3. If `code` is empty, reads `shortenedUrl` from the query string.
4. Calls `ShortenerService.FindOriginalUrl`.
5. The service extracts the code if a full URL was provided.
6. `ShortenerRepository.FindOriginalUrl` checks for a matching code and user ID.

Possible responses:

- `200 OK` when the original URL is found.
- `400 Bad Request` when both `code` and `shortenedUrl` are missing.
- `401 Unauthorized` when the token is missing or invalid.
- `404 Not Found` when no matching URL is found.

### `GET /api/v1/shorten/shortened-url`

Finds the shortened URL for an original URL owned by the authenticated user.

Query parameters:

```text
originalUrl=https://example.com/articles/my-long-url
```

Example:

```text
GET /api/v1/shorten/shortened-url?originalUrl=https://example.com/articles/my-long-url
```

Response:

```json
{
  "shortenedUrl": "http://localhost:8080/abc123"
}
```

Handler: `internal/handlers/shortener.go`

Flow:

1. Gets the authenticated user ID from request context.
2. Reads `originalUrl` from the query string.
3. Calls `ShortenerService.FindShortenedUrl`.
4. `ShortenerRepository.FindShortenedUrl` finds the short code for that original URL and user ID.
5. The service returns the full shortened URL using `HOST_URL`.

Possible responses:

- `200 OK` when the shortened URL is found.
- `400 Bad Request` when `originalUrl` is missing.
- `401 Unauthorized` when the token is missing or invalid.
- `404 Not Found` when no matching URL is found.

## Layer Breakdown

### `cmd/server`

`cmd/server/main.go` is the application composition root.

It:

- Loads `.env` values with `godotenv`.
- Creates the PostgreSQL pool.
- Creates repositories.
- Creates auth and shortener services.
- Creates handlers.
- Registers public and protected routes.
- Starts the HTTP server.

### Handlers

Handlers are responsible for HTTP-level behavior:

- Decode JSON bodies.
- Read query parameters and path parameters.
- Read authenticated user IDs from request context.
- Validate required request fields.
- Call services or repositories.
- Write JSON responses, redirects, and HTTP errors.

Files:

- `internal/handlers/auth.go`
- `internal/handlers/user.go`
- `internal/handlers/shortener.go`

### Services

Services hold application logic that should not live directly inside handlers.

`internal/auth/service.go` handles:

- User registration.
- Password verification.
- JWT generation.
- JWT validation.
- Refresh token login flow.
- Access token refresh flow.

`internal/shortener/shorten.go` handles:

- Short code generation.
- Full shortened URL building.
- Extracting a code from either a short code or full shortened URL.
- Calling repository methods for saving and retrieving URL records.

### Repositories

Repositories handle PostgreSQL queries.

`internal/repository/user.go`:

- Creates users.
- Finds users by email.
- Finds users by ID.

`internal/repository/refresh_token.go`:

- Creates refresh tokens.
- Finds refresh tokens by token string.
- Revokes refresh tokens.

`internal/repository/shortener.go`:

- Saves original URLs and short codes.
- Finds a short code from an original URL for a specific user.
- Finds an original URL from a short code for a specific user.
- Finds an original URL from a short code publicly for redirects.

### Middleware

`internal/middleware/auth.go` protects routes with JWT authentication.

It expects:

```http
Authorization: Bearer jwt-access-token
```

It validates the token and stores the user ID in request context so handlers can call:

```go
middleware.GetUserID(r)
```

### Models

Models represent application data:

- `models.User`
- `models.RefreshToken`
- `models.ShortenerModel`

These are used mainly by repositories and services.

### Cache

`internal/cache/redis.go` provides a small Redis wrapper with:

- `Set`
- `Get`
- `Delete`

Redis is currently not wired into URL redirects, refresh tokens, or route handlers. It is prepared for future caching/session work.

## Example Request Flow

### Creating a short URL

1. Client logs in through `POST /api/v1/auth/login`.
2. Server returns a JWT access token.
3. Client sends `POST /api/v1/shorten/create` with the bearer token.
4. Auth middleware validates the token and stores the user ID in context.
5. Shortener handler reads `originalUrl`.
6. Shortener service generates a short code.
7. Shortener repository saves the URL record in PostgreSQL.
8. Server returns the full shortened URL.

### Visiting a short URL

1. Browser requests `GET /api/v1/url/abc123`.
2. Public redirect handler reads `/api/v1/url/{code}` from the path.
3. Shortener service asks the repository for the original URL.
4. Server redirects the browser to the original URL.

## TODO

- Use Redis for storing refresh tokens instead of relying only on PostgreSQL.
- Implement URL tracking by recording and exposing the number of clicks for each shortened URL.
- Cache frequently visited short URLs in Redis to reduce PostgreSQL lookups during redirects.
- Build HTML templates for visual interaction so the app can be used from a browser instead of only Postman or API clients.

## Current Limitations

- Access tokens are signed with a new random JWT secret on every server restart.
- Refresh tokens are stored in PostgreSQL, not Redis.
- The `clicks` table exists, but redirects do not currently create click records.
- Short code collision handling is left to the database unique constraint.
- There are no HTML views yet.
