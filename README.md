# Game Manager Backend Service

A Go backend service with Google OAuth authentication, PostgreSQL, Redis, and RabbitMQ.

## Quick Start

### Prerequisites

- Go 1.22+
- Docker and Docker Compose
- golangci-lint (for linting)

### Setup

1. **Enable Git Hooks**
   ```bash
   git config core.hooksPath .githooks
   ```
   This enables the pre-commit hook that automatically runs `gofmt` and `golangci-lint` before each commit.

2. **Configure VS Code (Optional)**
   VS Code settings are pre-configured in `.vscode/settings.json` to use golangci-lint for linting and auto-format Go files on save.

3. **Download Go Dependencies**
   ```bash
   go mod download
   ```

4. **Configure Environment Variables**
   ```bash
   cp .env.example .env
   ```
   Edit `.env` and add your Google OAuth credentials:
   ```
   GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
   GOOGLE_CLIENT_SECRET=your-client-secret
   ```

5. **Start Services with Docker Compose**
   ```bash
   docker-compose up --build
   ```

   This starts:
   - **game-manager** app on port `8080`
   - **PostgreSQL** on port `5432`
   - **Redis** on port `6379`
   - **RabbitMQ** on port `5672`
   - **RabbitMQ Management UI** on port `15672` (http://localhost:15672)

## Development

### Running Locally

```bash
# Start dependencies only
docker-compose up -d postgres redis rabbitmq

# Run the application
go run cmd/server/main.go
```

### Git Hooks & Linting

A pre-commit hook automatically runs on every commit:
- `gofmt -s -w` - Auto-formats all Go code
- `golangci-lint run` - Lints all Go code

** Commits will be blocked if linting fails.** Fix any linting errors before committing.


### Manual Testing

```bash
# Run tests
./scripts/test.sh
```

## Environment Variables

See `.env.example` for all available configuration options.

Required for Google OAuth:
- `GOOGLE_CLIENT_ID` - Your Google OAuth Client ID
- `GOOGLE_CLIENT_SECRET` - Your Google OAuth Client Secret
- `GOOGLE_REDIRECT_URL` - OAuth callback URL (default: `http://localhost:8080/auth/google/callback`)
