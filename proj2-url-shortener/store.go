package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// URLStore manages URL mappings using Redis
type URLStore struct {
	client          *redis.Client
	ctx             context.Context
	expirationHours int
}

// NewURLStore creates a new URL store with Redis
func NewURLStore() *URLStore {
	// Read Redis configuration from environment variables
	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}

	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}

	redisPassword := os.Getenv("REDIS_PASSWORD")

	redisDB := 0
	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if db, err := strconv.Atoi(dbStr); err == nil {
			redisDB = db
		}
	}

	// Read URL expiration hours from environment variable
	expirationHours := 720 // default 30 days
	if expStr := os.Getenv("URL_EXPIRATION_HOURS"); expStr != "" {
		if exp, err := strconv.Atoi(expStr); err == nil && exp > 0 {
			expirationHours = exp
		}
	}

	// Create Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: redisPassword,
		DB:       redisDB,
	})

	return &URLStore{
		client:          rdb,
		ctx:             context.Background(),
		expirationHours: expirationHours,
	}
}

// Ping tests the Redis connection
func (s *URLStore) Ping() error {
	return s.client.Ping(s.ctx).Err()
}

// Close closes the Redis connection
func (s *URLStore) Close() error {
	return s.client.Close()
}

// GenerateShortKey generates a random short key
func (s *URLStore) GenerateShortKey() (string, error) {
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}

	// Encode to base64 and make URL-safe
	key := base64.URLEncoding.EncodeToString(b)
	// Take first 8 characters
	if len(key) > 8 {
		key = key[:8]
	}

	return key, nil
}

// Save stores a URL mapping in Redis with expiration
func (s *URLStore) Save(shortKey, longURL string) error {
	if shortKey == "" || longURL == "" {
		return errors.New("short key and long URL cannot be empty")
	}

	// Store with configured expiration time
	expiration := time.Duration(s.expirationHours) * time.Hour
	err := s.client.Set(s.ctx, shortKey, longURL, expiration).Err()
	return err
}

// Get retrieves a long URL by short key from Redis
func (s *URLStore) Get(shortKey string) (string, error) {
	longURL, err := s.client.Get(s.ctx, shortKey).Result()
	if err == redis.Nil {
		return "", errors.New("short URL not found")
	} else if err != nil {
		return "", err
	}

	return longURL, nil
}
