package cordkit

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func TestConcurrentSessions(t *testing.T) {
	// Create a test bot with mock configuration
	bot, err := NewBot("client.json")
	if err != nil {
		t.Fatalf("Failed to create bot: %v", err)
	}

	// Start the bot
	bot.Start()
	defer bot.Stop()

	// Number of concurrent sessions
	numSessions := 10
	var wg sync.WaitGroup
	wg.Add(numSessions)

	// Create a channel to track active connections
	connections := make(chan *Connection, numSessions)

	// Function to simulate a session
	simulateSession := func(id int) {
		defer wg.Done()

		// Create a unique session ID
		sessionID := fmt.Sprintf("testsession-%d", id)

		// Handle the connection
		conn := bot.HandleConnection(sessionID)
		connections <- conn

		// Simulate random activity duration
		activityDuration := time.Duration(rand.Intn(5)+1) * time.Second
		time.Sleep(activityDuration)

		// Kill the connection
		bot.KillConnection(conn)
	}

	// Launch concurrent sessions
	for i := 0; i < numSessions; i++ {
		go simulateSession(i)
	}

	// Wait for all sessions to complete
	wg.Wait()
	close(connections)

	// Verify all connections were handled
	connCount := 0
	for conn := range connections {
		if conn == nil {
			t.Error("Received nil connection")
		}
		connCount++
	}

	if connCount != numSessions {
		t.Errorf("Expected %d connections, got %d", numSessions, connCount)
	}
}
