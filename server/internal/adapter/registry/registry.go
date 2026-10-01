package registry

import (
	"hash/fnv"
	"sync"

	"messaging/server/internal/domain"
)

const numShards = 64

type shard struct {
	mu    sync.RWMutex
	conns map[domain.DeviceID]chan<- []byte
}

// Registry is a sharded in-memory registry of active WebSocket connections.
type Registry struct {
	shards [numShards]*shard
}

func NewRegistry() *Registry {
	r := &Registry{}
	for i := 0; i < numShards; i++ {
		r.shards[i] = &shard{
			conns: make(map[domain.DeviceID]chan<- []byte),
		}
	}
	return r
}

func (r *Registry) getShard(id domain.DeviceID) *shard {
	h := fnv.New32a()
	h.Write([]byte(id))
	return r.shards[h.Sum32()%numShards]
}

func (r *Registry) Register(deviceID domain.DeviceID, send chan<- []byte) {
	s := r.getShard(deviceID)
	s.mu.Lock()
	defer s.mu.Unlock()

	if old, exists := s.conns[deviceID]; exists {
		// Close the old channel to force the old writer goroutine to exit.
		// We use a non-blocking close approach in real systems, but for simplicity here we just close.
		close(old)
	}
	s.conns[deviceID] = send
}

func (r *Registry) Unregister(deviceID domain.DeviceID) {
	s := r.getShard(deviceID)
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.conns, deviceID)
}

func (r *Registry) Send(deviceID domain.DeviceID, frame []byte) bool {
	s := r.getShard(deviceID)
	s.mu.RLock()
	ch, ok := s.conns[deviceID]
	s.mu.RUnlock()

	if !ok {
		return false
	}

	select {
	case ch <- frame:
		return true
	default:
		// Outbound channel is full. The device is too slow.
		// We drop the frame. The client will rely on the mailbox drain upon reconnect.
		// A robust system might unregister the client here to kill the stuck connection.
		return false
	}
}

func (r *Registry) IsLocal(deviceID domain.DeviceID) bool {
	s := r.getShard(deviceID)
	s.mu.RLock()
	_, ok := s.conns[deviceID]
	s.mu.RUnlock()
	return ok
}

func (r *Registry) ConnectedDevices() []domain.DeviceID {
	var out []domain.DeviceID
	for i := 0; i < numShards; i++ {
		s := r.shards[i]
		s.mu.RLock()
		for id := range s.conns {
			out = append(out, id)
		}
		s.mu.RUnlock()
	}
	return out
}

func (r *Registry) Count() int {
	var count int
	for i := 0; i < numShards; i++ {
		s := r.shards[i]
		s.mu.RLock()
		count += len(s.conns)
		s.mu.RUnlock()
	}
	return count
}
