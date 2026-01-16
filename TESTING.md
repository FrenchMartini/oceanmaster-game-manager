# Testing Guide - Google OAuth Authentication

## Prerequisites

1. **Set up Google OAuth Credentials:**
   - Go to [Google Cloud Console](https://console.cloud.google.com/)
   - Create a new project or select an existing one
   - Enable Google+ API
   - Go to "Credentials" → "Create Credentials" → "OAuth 2.0 Client ID"
   - Set authorized redirect URI: `http://localhost:8080/auth/google/callback`
   - Copy the Client ID and Client Secret

2. **Configure Environment Variables:**
   ```bash
   cp .env.example .env
   ```
   Edit `.env` and add your Google OAuth credentials:
   ```
   GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
   GOOGLE_CLIENT_SECRET=your-client-secret
   GOOGLE_REDIRECT_URL=http://localhost:8080/auth/google/callback
   ```

3. **Start Docker services:**
   ```bash
   docker-compose up -d
   ```
   This starts:
   - PostgreSQL on port 5432
   - Redis on port 6379
   - RabbitMQ on port 5672 (and management UI on 15672)

4. **Wait for services to be ready** (about 10-15 seconds)

## Running the Server

### Option 1: Run with Docker Compose
```bash
docker-compose up --build
```

### Option 2: Run Locally
```bash
go run cmd/server/main.go
```

The server will start on port 8080.

## Testing Endpoints

### 1. Health Check
```bash
curl http://localhost:8080/health
```

**Expected Response:**
```json
{
  "status": "healthy",
  "service": "game-manager"
}
```

### 2. Google OAuth Login Flow

**Step 1: Initiate Google Login**
Open in browser or use curl:
```bash
curl -L http://localhost:8080/auth/google/login
```

Or visit in browser:
```
http://localhost:8080/auth/google/login
```

This will redirect you to Google's OAuth consent screen.

**Step 2: After Google Authentication**
After you authorize the application, Google will redirect you to:
```
http://localhost:8080/auth/google/callback?code=...&state=...
```

The callback endpoint will:
1. Exchange the authorization code for an access token
2. Fetch user information from Google
3. Create or find the user in the database
4. Generate a JWT token
5. Return the token and user info as JSON

**Expected Response (200 OK):**
```json
{
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": 1,
    "email": "user@gmail.com",
    "name": "John Doe",
    "picture": "https://lh3.googleusercontent.com/...",
    "created_at": "2024-01-01T00:00:00Z"
  }
}
```

### 3. Using the JWT Token

Once you have the token, you can use it for authenticated requests:

```bash
TOKEN="your-jwt-token-here"
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/your-protected-endpoint
```

## Testing with Browser

1. **Start the server:**
   ```bash
   go run cmd/server/main.go
   ```

2. **Open browser and navigate to:**
   ```
   http://localhost:8080/auth/google/login
   ```

3. **Sign in with Google** - You'll be redirected to Google's sign-in page

4. **Authorize the app** - After signing in, Google will redirect back with the token

5. **Check the response** - You should see JSON with your JWT token and user info

## Manual Database Verification

You can verify users were created in PostgreSQL:

```bash
docker exec -it game-manager-postgres psql -U postgres -d gamemanager -c "SELECT id, email, name, google_id, created_at FROM users;"
```

## Database Schema

The users table now has:
- `id` - Primary key
- `email` - User's email (unique)
- `google_id` - Google user ID (unique)
- `name` - User's display name
- `picture` - User's profile picture URL
- `created_at` - Timestamp

## Troubleshooting

1. **"Google OAuth not configured" error:**
   - Make sure `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET` are set in `.env`
   - Restart the server after updating `.env`

2. **Redirect URI mismatch:**
   - Ensure the redirect URI in Google Console matches exactly: `http://localhost:8080/auth/google/callback`
   - For production, update both `.env` and Google Console

3. **Connection refused errors:**
   - Make sure Docker services are running: `docker-compose ps`
   - Check logs: `docker-compose logs`

4. **Database connection errors:**
   - Wait a bit longer for PostgreSQL to fully start
   - Check: `docker-compose logs postgres`

5. **Build errors:**
   - Run `go mod tidy` to ensure dependencies are up to date
   - Run `go build ./...` to check for compilation errors

## Production Setup

For production:
1. Update `GOOGLE_REDIRECT_URL` to your production URL
2. Add the production redirect URI in Google Cloud Console
3. Set `ENV=production` in your environment variables
4. Use a strong `JWT_SECRET` (at least 32 characters)
5. Enable HTTPS for secure cookie handling
