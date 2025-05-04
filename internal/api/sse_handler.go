package api

import (
	"bufio"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/hackerchai/memento/internal/middleware"
	"github.com/hackerchai/memento/internal/response"
	"github.com/hackerchai/memento/internal/sse"
	"github.com/hackerchai/memento/pkg/xlog"
)

// SSEHandler handles Server-Sent Event connections.
type SSEHandler struct {
	broker *sse.Broker
	logger *xlog.Logger
}

// NewSSEHandler creates a new SSEHandler.
func NewSSEHandler(broker *sse.Broker, logger *xlog.Logger) *SSEHandler {
	return &SSEHandler{
		broker: broker,
		logger: logger.With().Str("handler", "SSEHandler").Logger(),
	}
}

// ConnectSSE establishes an SSE connection for the authenticated user using Fiber's StreamWriter.
// @Summary Establish Server-Sent Event connection
// @Description Establishes an SSE connection to receive real-time updates (e.g., article processing status).
// @Tags sse
// @Produce text/event-stream
// @Security BearerAuth
// @Success 200 {string} string "SSE stream established"
// @Failure 401 {object} response.ErrorResponse "Unauthorized (invalid/missing token - code: 01001)"
// @Failure 500 {object} response.ErrorResponse "Internal server error (failed to establish stream - code: 00001)"
// @Router /sse [get]
func (h *SSEHandler) ConnectSSE(c *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(c, h.logger)
	if err != nil {
		return response.HandleError(c, h.logger, err)
	}

	// Set headers for SSE
	c.Set(fiber.HeaderContentType, "text/event-stream")
	c.Set(fiber.HeaderCacheControl, "no-cache")
	c.Set(fiber.HeaderConnection, fiber.HeaderKeepAlive)
	c.Set(fiber.HeaderTransferEncoding, "chunked")

	// Create a channel for this specific client connection
	// Use the buffer size defined in the broker if possible, otherwise default
	// bufferSize := h.broker.BufferSize() // Assuming Broker has a BufferSize getter
	clientChan := make(chan []byte, 10) // Defaulting to 10 for now

	// Create and register the client with the broker
	client := &sse.Client{
		UserID: userID,
		Chan:   clientChan,
	}
	h.broker.Register(client)
	h.logger.InfoX().Stringer("user_id", userID).Msg("SSE client registered, starting stream")

	// Use Fiber's StreamWriter for handling the SSE connection
	c.Context().SetBodyStreamWriter(func(w *bufio.Writer) {
		// Ensure client is unregistered when the stream writer function exits
		defer func() {
			h.broker.Unregister(userID)
			h.logger.InfoX().Stringer("user_id", userID).Msg("SSE stream writer stopped, client unregistered")
		}()

		// Send an initial confirmation event? Optional.
		// fmt.Fprintf(w, "event: connected\ndata: {}\n\n")
		// if err := w.Flush(); err != nil { ... }

		keepAliveTicker := time.NewTicker(15 * time.Second) // Send keep-alive pings
		defer keepAliveTicker.Stop()

		h.logger.InfoX().Stringer("user_id", userID).Msg("SSE stream loop started")

		for {
			select {
			case msg, ok := <-clientChan:
				if !ok {
					// Channel closed by the broker (e.g., new connection for same user)
					h.logger.InfoX().Stringer("user_id", userID).Msg("SSE client channel closed by broker, stopping stream")
					return // Exit the loop
				}
				// Write the message from the broker
				_, err := w.Write(msg)
				if err != nil {
					h.logger.WarnX().Err(err).Stringer("user_id", userID).Msg("Error writing to SSE stream")
					return // Exit on write error
				}
				// Flush the data to the client
				err = w.Flush()
				if err != nil {
					h.logger.WarnX().Err(err).Stringer("user_id", userID).Msg("Error flushing SSE stream")
					return // Exit on flush error
				}

			case <-keepAliveTicker.C:
				// Send a keep-alive comment
				_, err := fmt.Fprintf(w, ": keep-alive\n\n")
				if err != nil {
					h.logger.WarnX().Err(err).Stringer("user_id", userID).Msg("Error sending SSE keep-alive")
					return // Exit on write error
				}
				// Flush the keep-alive
				err = w.Flush()
				if err != nil {
					h.logger.WarnX().Err(err).Stringer("user_id", userID).Msg("Error flushing SSE keep-alive")
					return // Exit on flush error
				}
			}
		}
	})

	// Return nil as the response is handled by the StreamWriter
	return nil
}
