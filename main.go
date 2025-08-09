// Package pmdauth provides simple JWT token validation
package pmdauth

import (
	"context"
	"errors"
	"log"
	"os"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// User represents the user structure
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

// Claims represents JWT claims
type Claims struct {
	UserID   primitive.ObjectID `json:"user_id"`
	GoogleID string             `json:"google_id"`
	Email    string             `json:"email"`
	jwt.RegisteredClaims
}

var (
	jwtSecret       string
	mongoCollection *mongo.Collection
	once            sync.Once
	initError       error
	devToken        string
)

// InitAuthLib automatically sets up the auth library from environment variables
func InitAuthLib() {
	once.Do(func() {
		// Load .env file if it exists (ignore error if not found)
		_ = godotenv.Load()

		// Read environment variables
		mongoURI := os.Getenv("MONGODB_URI")
		mongoDB := os.Getenv("MONGODB_DATABASE")
		jwtSecret = os.Getenv("JWT_SECRET")
		devToken = os.Getenv("DEV_TOKEN")

		if mongoURI == "" || mongoDB == "" || jwtSecret == "" {
			initError = errors.New("missing required environment variables: MONGODB_URI, MONGODB_DATABASE, JWT_SECRET")
			return
		}

		// Connect to MongoDB
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		clientOptions := options.Client().ApplyURI(mongoURI)
		client, err := mongo.Connect(ctx, clientOptions)
		if err != nil {
			initError = err
			return
		}

		// Test the connection
		err = client.Ping(ctx, nil)
		if err != nil {
			initError = err
			return
		}

		mongoCollection = client.Database(mongoDB).Collection("users")
		log.Println("Auth library initialized successfully")
	})
}

// ValidateToken validates a JWT token and returns the user if valid
func ValidateToken(tokenString string) (*User, error) {
	if initError != nil {
		return nil, initError
	}

	isDevToken := tokenString == devToken

	// Parse and validate JWT
	claims := &Claims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil && !isDevToken {
		return nil, err
	}

	if !token.Valid && !isDevToken {
		return nil, errors.New("invalid token")
	}

	// Get user from database
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	objectID, err := primitive.ObjectIDFromHex(claims.Subject)
	if err != nil {
		return nil, errors.New("invalid user ID in token")
	}

	var user User
	err = mongoCollection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		return nil, errors.New("user not found")
	}

	return &user, nil
}

// GenerateToken creates a JWT token for a user
func GenerateToken(user User, expireHours int) (string, error) {
	if initError != nil {
		return "", initError
	}

	expirationTime := time.Now().Add(time.Duration(expireHours) * time.Hour)

	claims := &Claims{
		UserID:   user.ID,
		GoogleID: user.GoogleID,
		Email:    user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   user.ID.Hex(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(jwtSecret))
}

// HealthCheck returns whether the auth library is properly initialized
func HealthCheck() error {
	return initError
}
