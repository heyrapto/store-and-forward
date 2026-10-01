package usecase

import (
	"context"
	"fmt"

	"messaging/server/internal/usecase/port"
)

// SweepExpired is a background job that hard-deletes delivery rows
// whose expires_at has passed.
// Run it on a ticker (every 10 minutes in production, every 30 seconds in demo mode).
type SweepExpired struct {
	Deliveries port.DeliveryRepo
}

// Execute runs one sweep pass and returns the number of rows deleted.
func (uc *SweepExpired) Execute(ctx context.Context) (int64, error) {
	n, err := uc.Deliveries.DeleteExpired(ctx)
	if err != nil {
		return 0, fmt.Errorf("sweep expired: %w", err)
	}
	return n, nil
}
