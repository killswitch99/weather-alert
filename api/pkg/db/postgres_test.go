package db

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDefaultConfig verifies the default configuration settings
func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()
	
	assert.Equal(t, 25, config.MaxOpenConns)
	assert.Equal(t, 25, config.MaxIdleConns)
	assert.Equal(t, 5*time.Minute, config.ConnMaxLifetime)
	assert.Equal(t, 10*time.Second, config.QueryTimeout)
}

func TestConnect(t *testing.T) {
	// Skip if no test database available
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	config := DefaultConfig()
	config.URI = testDBURL

	// Test successful connection
	err := Connect(config)
	require.NoError(t, err)
	defer Disconnect()

	assert.NotNil(t, pool)
	assert.NoError(t, pool.Ping(context.Background()))

	// Test GetPool returns the pool
	gotPool := GetPool()
	assert.Equal(t, pool, gotPool)
}

func TestConnectError(t *testing.T) {
	// Test connection error with invalid URI
	config := DefaultConfig()
	config.URI = "postgres://invalid:invalid@localhost:5432/nonexistentdb"

	err := Connect(config)
	assert.Error(t, err)
}

func TestWithTimeout(t *testing.T) {
	baseCtx := context.Background()
	ctx, cancel := WithTimeout(baseCtx)
	defer cancel()

	deadline, ok := ctx.Deadline()
	assert.True(t, ok)
	assert.WithinDuration(t, time.Now().Add(defaultQueryTimeout), deadline, time.Second)
}

func TestHealthCheck(t *testing.T) {
	// Skip if no test database available
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	config := DefaultConfig()
	config.URI = testDBURL
	
	err := Connect(config)
	require.NoError(t, err)
	defer Disconnect()

	err = HealthCheck(context.Background())
	assert.NoError(t, err)

	// Test when pool is closed
	Disconnect()
	err = HealthCheck(context.Background())
	assert.Error(t, err)
}

func TestWithTransaction(t *testing.T) {
	// Skip if no test database available
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	config := DefaultConfig()
	config.URI = testDBURL
	
	err := Connect(config)
	require.NoError(t, err)
	defer Disconnect()

	// Test successful transaction
	err = WithTransaction(context.Background(), func(tx pgx.Tx) error {
		// Simply return nil to test successful commit
		return nil
	})
	assert.NoError(t, err)

	// Test transaction rollback
	expectedErr := assert.AnError
	err = WithTransaction(context.Background(), func(tx pgx.Tx) error {
		return expectedErr
	})
	assert.ErrorIs(t, err, expectedErr)

	// Test transaction begin error (requires mocking)
	// In a real test environment, you might use a mock library
}

// TestDisconnect verifies the Disconnect function
func TestDisconnect(t *testing.T) {
	// Skip if no test database available
	testDBURL := os.Getenv("TEST_DATABASE_URL")
	if testDBURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration test")
	}

	config := DefaultConfig()
	config.URI = testDBURL
	
	err := Connect(config)
	require.NoError(t, err)
	
	// Verify pool is not nil before disconnecting
	assert.NotNil(t, pool)
	
	// Call Disconnect
	Disconnect()
	
	// Since we can't directly check if pool is closed (it's still not nil),
	// we can verify that operations on it fail
	err = HealthCheck(context.Background())
	assert.Error(t, err)
}
