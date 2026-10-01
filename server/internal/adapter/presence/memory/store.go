package memory

import (
	"context"
	"sync"

	"messaging/server/internal/domain"
)

// Store is an in-memory implementation of port.PresenceStore.
// Suitable for Episode 1 (single node).
type Store struct {
	mu      sync.RWMutex
	devices map[domain.DeviceID]string // deviceID -> gatewayID
}

func NewStore() *Store {
	return &Store{
		devices: make(map[domain.DeviceID]string),
	}
}

func (s *Store) SetOnline(ctx context.Context, deviceID domain.DeviceID, gatewayID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.devices[deviceID] = gatewayID
	return nil
}

func (s *Store) SetOffline(ctx context.Context, deviceID domain.DeviceID) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.devices, deviceID)
	return nil
}

func (s *Store) IsOnline(ctx context.Context, deviceID domain.DeviceID) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	_, ok := s.devices[deviceID]
	return ok, nil
}

func (s *Store) GatewayOf(ctx context.Context, deviceID domain.DeviceID) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.devices[deviceID], nil
}

func (s *Store) OnlineDevices(ctx context.Context) ([]domain.DeviceID, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	out := make([]domain.DeviceID, 0, len(s.devices))
	for id := range s.devices {
		out = append(out, id)
	}
	return out, nil
}
