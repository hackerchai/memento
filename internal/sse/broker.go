package sse

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hackerchai/memento/pkg/xlog"
)

// Client represents a single SSE connection for a user.
type Client struct {
	UserID uuid.UUID
	Chan   chan []byte // Channel to send messages to this client
}

// Broker manages all active SSE client connections.
type Broker struct {
	clients    map[uuid.UUID]*Client // Map UserID to Client
	mu         sync.RWMutex          // Mutex for concurrent access to clients map
	logger     *xlog.Logger          // Optional: for logging broker events
	bufferSize int                   // Buffer size for client channels
}

// NewBroker creates a new SSE Broker.
func NewBroker(logger *xlog.Logger) *Broker {
	const defaultBufferSize = 10
	var brokerLogger *xlog.Logger
	if logger != nil {
		// Assume With().Str().Logger() pattern is correct if logger is available
		brokerLogger = logger.With().Str("component", "SSEBroker").Logger()
	}

	return &Broker{
		clients:    make(map[uuid.UUID]*Client),
		logger:     brokerLogger,
		bufferSize: defaultBufferSize,
	}
}

// Register adds a new client to the broker.
// If a client for the same user already exists, it closes the old client's channel
// before registering the new one.
func (b *Broker) Register(client *Client) {
	if client == nil || client.Chan == nil {
		if b.logger != nil {
			b.logger.WarnX().Msg("Attempted to register nil client or client with nil channel")
		}
		return
	}

	b.mu.Lock()
	defer b.mu.Unlock()

	// If another connection exists for the same user, close its channel first.
	if existingClient, ok := b.clients[client.UserID]; ok {
		if b.logger != nil {
			// Revert to chaining after level method
			b.logger.InfoX().Stringer("user_id", client.UserID).Msg("Closing existing SSE channel for user due to new connection")
		}
		if existingClient.Chan != nil {
			close(existingClient.Chan)
		}
	}

	// Add the new client
	b.clients[client.UserID] = client
	if b.logger != nil {
		b.logger.InfoX().Stringer("user_id", client.UserID).Msg("SSE client registered")
	}
}

// Unregister removes a client from the broker and closes its channel.
// It's safe to call Unregister even if the client is already unregistered.
func (b *Broker) Unregister(userID uuid.UUID) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if client, ok := b.clients[userID]; ok {
		// Check if channel is not nil and not already closed before trying to close it.
		// This requires knowing the channel state or handling potential panics,
		// or simply ensuring close is only called once.
		// For simplicity, we assume the handler ensures Unregister is called appropriately.
		// A select-based close might be safer if unsure:
		// select {
		// case <-client.Chan:
		//     // Channel already closed
		// default:
		//     close(client.Chan)
		// }
		if client.Chan != nil {
			// Consider safer closing mechanism if needed
			close(client.Chan)
			client.Chan = nil // Set to nil after closing
		}
		delete(b.clients, userID)
		if b.logger != nil {
			b.logger.InfoX().Stringer("user_id", userID).Msg("SSE client unregistered")
		}
	} else {
		if b.logger != nil {
			b.logger.WarnX().Stringer("user_id", userID).Msg("Attempted to unregister non-existent client")
		}
	}
}

// NotifyUser sends a message to a specific user's active SSE connection, if any.
// It sends the message in a non-blocking way to prevent blocking the caller.
func (b *Broker) NotifyUser(userID uuid.UUID, message []byte) {
	b.mu.RLock()
	client, ok := b.clients[userID]
	b.mu.RUnlock() // Unlock after getting the client reference

	if ok && client != nil && client.Chan != nil {
		select {
		case client.Chan <- message:
			if b.logger != nil {
				// Revert to chaining after level method
				b.logger.DebugX().Stringer("user_id", userID).Int("msg_len", len(message)).Msg("Sent message to SSE client channel")
			}
		case <-time.After(100 * time.Millisecond): // Timeout to prevent blocking indefinitely
			if b.logger != nil {
				b.logger.WarnX().Stringer("user_id", userID).Msg("Failed to send message to SSE client channel: send timed out (channel might be full or blocked)")
			}
			// Optionally, unregister the client here if timeout implies disconnection
			// b.Unregister(userID)
			// Add a case for channel closed? A read would detect this.
			// case _, ok := <-client.Chan: if !ok { ... channel closed ... }
			// But that requires reading, which isn't the goal here. The write attempt failing
			// or timing out is usually sufficient indication of a problem.
		}
	} else {
		if b.logger != nil {
			b.logger.DebugX().Stringer("user_id", userID).Msg("No active SSE client found for user to notify")
		}
	}
}

// CloseAll closes all client connections and clears the broker.
// Useful during graceful shutdown.
func (b *Broker) CloseAll() {
	b.mu.Lock()
	defer b.mu.Unlock()

	if b.logger != nil {
		b.logger.InfoX().Int("count", len(b.clients)).Msg("Closing all SSE client connections")
	}
	for userID, client := range b.clients {
		if client != nil && client.Chan != nil {
			close(client.Chan)
		}
		delete(b.clients, userID) // Remove from map
	}
}
