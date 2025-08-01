# PMD Auth

A simple, secure JWT authentication library for Go applications with MongoDB integration.

## Features

- 🔐 JWT token generation and validation
- 🗄️ MongoDB user storage integration
- 🔒 Secure environment-based configuration
- ⚡ Thread-safe initialization
- 🚀 Simple API design
- ⏱️ Configurable token expiration
- 📝 Automatic .env file loading

## Installation

```bash
go get github.com/golang-jwt/jwt/v5
go get github.com/joho/godotenv
go get go.mongodb.org/mongo-driver/mongo
go get github.com/your-username/pmdauth
```

## Quick Start

### 1. Environment Setup

Create a `.env` file in your project root:

```env
MONGODB_URI=mongodb://localhost:27017
MONGODB_DATABASE=your_database_name
JWT_SECRET=your-super-secret-jwt-key-here
```

### 2. Initialize the Library

```go
package main

import (
    "log"
    "github.com/your-username/pmdauth"
)

func main() {
    // Initialize the auth library
    pmdauth.InitAuthLib()
    
    // Check if initialization was successful
    if err := pmdauth.HealthCheck(); err != nil {
        log.Fatal("Auth initialization failed:", err)
    }
    
    log.Println("Auth library ready!")
}
```

### 3. Generate Tokens

```go
// Create a user
user := pmdauth.User{
    GoogleID:  "123456789",
    Email:     "user@example.com",
    Name:      "John Doe",
    Picture:   "https://example.com/avatar.jpg",
    Status:    1,
    CreatedAt: time.Now(),
    UpdatedAt: time.Now(),
}

// Generate a token that expires in 24 hours
token, err := pmdauth.GenerateToken(user, 24)
if err != nil {
    log.Fatal("Token generation failed:", err)
}

fmt.Println("Generated token:", token)
```

### 4. Validate Tokens

```go
// Validate a token
user, err := pmdauth.ValidateToken(tokenString)
if err != nil {
    log.Println("Invalid token:", err)
    return
}

fmt.Printf("Valid user: %s (%s)\n", user.Name, user.Email)
```

## API Reference

### Types

#### User
```go
type User struct {
    ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    GoogleID  string             `json:"google_id" bson:"google_id"`
    Email     string             `json:"email" bson:"email"`
    Name      string             `json:"name" bson:"name"`
    Picture   string             `json:"picture" bson:"picture"`
    Status    int                `json:"status" bson:"status"`
    CreatedAt time.Time          `json:"created_at" bson:"created_at"`
    UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}
```

#### Claims
```go
type Claims struct {
    UserID   primitive.ObjectID `json:"user_id"`
    GoogleID string             `json:"google_id"`
    Email    string             `json:"email"`
    jwt.RegisteredClaims
}
```

### Functions

#### `InitAuthLib()`
Initializes the authentication library. Must be called before using other functions.
- Loads environment variables from `.env` file
- Connects to MongoDB
- Sets up JWT secret

#### `ValidateToken(tokenString string) (*User, error)`
Validates a JWT token and returns the associated user.
- **Parameters:**
    - `tokenString`: The JWT token to validate
- **Returns:**
    - `*User`: User object if token is valid
    - `error`: Error if validation fails

#### `GenerateToken(user User, expireHours int) (string, error)`
Generates a JWT token for a user.
- **Parameters:**
    - `user`: User object to generate token for
    - `expireHours`: Token expiration time in hours
- **Returns:**
    - `string`: Generated JWT token
    - `error`: Error if generation fails

#### `HealthCheck() error`
Checks if the auth library is properly initialized.
- **Returns:**
    - `error`: nil if healthy, error if there are issues

## HTTP Handler Example

```go
package main

import (
    "encoding/json"
    "net/http"
    "strings"
    "github.com/your-username/printmedi-auth"
)

func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        // Get token from Authorization header
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Missing authorization header", http.StatusUnauthorized)
            return
        }

        // Extract token from "Bearer <token>"
        tokenString := strings.TrimPrefix(authHeader, "Bearer ")
        if tokenString == authHeader {
            http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
            return
        }

        // Validate token
        user, err := printmedi.ValidateToken(tokenString)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        // Add user to request context
        ctx := context.WithValue(r.Context(), "user", user)
        next(w, r.WithContext(ctx))
    }
}

func meHandler(w http.ResponseWriter, r *http.Request) {
    user := r.Context().Value("user").(*printmedi.User)
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(user)
}

func main() {
    printmedi.InitAuthLib()
    
    http.HandleFunc("/auth/me", authMiddleware(meHandler))
    log.Println("Server starting on :8181")
    log.Fatal(http.ListenAndServe(":8181", nil))
}
```

## Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `MONGODB_URI` | Yes | MongoDB connection string |
| `MONGODB_DATABASE` | Yes | MongoDB database name |
| `JWT_SECRET` | Yes | Secret key for JWT signing |

## Error Handling

The library returns descriptive errors for common scenarios:
- Missing environment variables
- MongoDB connection failures
- Invalid tokens
- User not found
- Token parsing errors

## Security Considerations

- 🔒 JWT secrets should be strong and unique
- ⏰ Use appropriate token expiration times
- 🔄 Consider implementing token refresh mechanisms
- 🚫 Validate all inputs and handle errors properly
- 🔐 Store JWT secrets securely (environment variables, not in code)

## MongoDB Collection Structure

The library expects a `users` collection with documents matching the `User` struct:

```json
{
  "_id": ObjectId("..."),
  "google_id": "123456789",
  "email": "user@example.com",
  "name": "John Doe",
  "picture": "https://example.com/avatar.jpg",
  "status": 1,
  "created_at": ISODate("..."),
  "updated_at": ISODate("...")
}
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.