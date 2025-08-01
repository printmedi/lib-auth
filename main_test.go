package pmdauth

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func setupTestDB(t *testing.T) *mongo.Collection {
	// *** HERE IS THE MONGODB CONNECTION STRING ***
	mongoURI := "mongodb://adminz:adM1nzEn5@localhost:27017/?authSource=zero&authMechanism=SCRAM-SHA-1"
	// *** END OF CONNECTION STRING ***

	mongoDB := "zero"
	testSecret := "test-jwt-secret-key"

	// Set global variables directly for testing
	jwtSecret = testSecret

	// Connect to test database
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// *** CONNECTION STRING USED HERE ***
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	// *** END OF USAGE ***

	if err != nil {
		t.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	// Test the connection with authentication
	err = client.Ping(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to ping MongoDB: %v", err)
	}

	// Test database access by listing collections
	database := client.Database(mongoDB)
	_, err = database.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		t.Fatalf("Failed to access database '%s': %v", mongoDB, err)
	}

	collection := database.Collection("users")

	// Test collection access by counting documents
	count, err := collection.CountDocuments(ctx, bson.M{})
	if err != nil {
		t.Fatalf("Failed to access collection 'users': %v", err)
	}
	log.Printf("✓ Connected to database '%s', collection 'users' has %d documents", mongoDB, count)

	// Set the global collection for the library
	mongoCollection = collection

	log.Println("✓ Test database setup complete")
	return collection
}

func cleanupTestDB(t *testing.T, collection *mongo.Collection) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Instead of dropping the database, just clean up test data
	filter := bson.M{"email": "lib-auth-test@gmail.com"}
	_, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		log.Printf("Warning: Could not clean up test users: %v", err)
	} else {
		log.Println("✓ Test data cleaned up")
	}
}

func TestHealthCheckSimple(t *testing.T) {
	// *** HERE IS THE MONGODB CONNECTION STRING AGAIN ***
	err := os.Setenv("MONGODB_URI", "mongodb://adminz:adM1nzEn5@localhost:27017/?authSource=zero&authMechanism=SCRAM-SHA-1")
	if err != nil {
		return
	}
	// *** END OF CONNECTION STRING ***

	err = os.Setenv("MONGODB_DATABASE", "zero")
	if err != nil {
		return
	}
	err = os.Setenv("JWT_SECRET", "test-secret")
	if err != nil {
		return
	}

	// Reset and call InitAuthLib
	initError = nil
	jwtSecret = ""
	mongoCollection = nil

	// Manually call the initialization logic
	mongoURI := os.Getenv("MONGODB_URI")
	mongoDB := os.Getenv("MONGODB_DATABASE")
	jwtSecret = os.Getenv("JWT_SECRET")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// *** CONNECTION STRING USED HERE ***
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(ctx, clientOptions)
	// *** END OF USAGE ***

	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to ping: %v", err)
	}

	mongoCollection = client.Database(mongoDB).Collection("users")

	// Now test HealthCheck
	err = HealthCheck()
	if err != nil {
		t.Fatalf("HealthCheck failed: %v", err)
	}
	log.Println("✓ HealthCheck passed")
}

func TestAuthLibraryOperationsSimple(t *testing.T) {
	// Setup test database (uses the connection string in setupTestDB function)
	collection := setupTestDB(t)
	defer cleanupTestDB(t, collection)

	ctx := context.Background()
	testEmail := "lib-auth-test@gmail.com"

	// Step 1: Clean up any existing test user (optional since we clean up at the end)
	filter := bson.M{"email": testEmail}
	_, err := collection.DeleteMany(ctx, filter)
	if err != nil {
		return
	} // Ignore errors for cleanup
	log.Printf("✓ Cleaned up any existing users with email %s", testEmail)

	// Step 2: Create a test user directly in MongoDB
	testUser := User{
		GoogleID:  "test-google-id-123",
		Email:     testEmail,
		Name:      "Test User",
		Picture:   "https://example.com/picture.jpg",
		Status:    1,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	insertResult, err := collection.InsertOne(ctx, testUser)
	if err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}
	log.Printf("✓ Inserted test user with ID: %v", insertResult.InsertedID)

	// Get the inserted user to get the actual ID
	var insertedUser User
	err = collection.FindOne(ctx, bson.M{"email": testEmail}).Decode(&insertedUser)
	if err != nil {
		t.Fatalf("Failed to retrieve inserted user: %v", err)
	}

	log.Printf("Retrieved User Information:")
	log.Printf("  ID: %s", insertedUser.ID.Hex())
	log.Printf("  Google ID: %s", insertedUser.GoogleID)
	log.Printf("  Email: %s", insertedUser.Email)
	log.Printf("  Name: %s", insertedUser.Name)
	log.Printf("  Picture: %s", insertedUser.Picture)
	log.Printf("  Status: %d", insertedUser.Status)

	// Step 3: Test JWT generation
	token, err := GenerateToken(insertedUser, 24)
	if err != nil {
		t.Fatalf("Failed to generate JWT token: %v", err)
	}
	log.Printf("✓ Generated JWT token: %s", token[:50]+"...")

	// Step 4: Test JWT validation (the main function)
	validatedUser, err := ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed to validate token: %v", err)
	}

	// Step 5: Verify the validated user matches our test user
	if validatedUser.Email != testEmail {
		t.Errorf("Expected email %s, got %s", testEmail, validatedUser.Email)
	}
	if validatedUser.Name != "Test User" {
		t.Errorf("Expected name 'Test User', got %s", validatedUser.Name)
	}
	if validatedUser.GoogleID != "test-google-id-123" {
		t.Errorf("Expected GoogleID 'test-google-id-123', got %s", validatedUser.GoogleID)
	}
	if validatedUser.ID != insertedUser.ID {
		t.Errorf("Expected ID %s, got %s", insertedUser.ID.Hex(), validatedUser.ID.Hex())
	}

	log.Printf("✓ Token validation successful! User: %s", validatedUser.Email)

	// Step 6: Test with invalid token
	_, err = ValidateToken("invalid-token")
	if err == nil {
		t.Error("Expected ValidateToken to fail with invalid token, but it succeeded")
	}
	log.Printf("✓ Invalid token correctly rejected: %v", err)

	log.Println("✓ All auth library tests completed successfully!")
}
