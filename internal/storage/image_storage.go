package storage

import (
	"context"

	"github.com/google/uuid"
)

// ImageStorage defines the interface for storing and retrieving cached images.
type ImageStorage interface {
	// Store saves the image data for a specific user and returns its publicly accessible URL/path.
	// imageName should preferably be unique (e.g., content hash + extension).
	Store(ctx context.Context, userID uuid.UUID, imageName string, imageData []byte) (publicURL string, err error)

	// GetPublicURL retrieves the public URL for a stored image.
	// It might check for existence internally or just calculate the expected path.
	// Returns an error if the imageName format is invalid or calculation fails.
	GetPublicURL(ctx context.Context, userID uuid.UUID, imageName string) (publicURL string, err error)

	// GetURL retrieves the public URL for a stored image without fetching the data.
	// Returns an error if the image doesn't exist (or based on implementation specifics).
	// GetURL(ctx context.Context, userID uuid.UUID, imageName string) (string, error)

	// Exists checks if an image already exists in the storage.
	Exists(ctx context.Context, userID uuid.UUID, imageName string) (bool, error)

	// TODO: Consider adding a Delete method if needed later.
}
