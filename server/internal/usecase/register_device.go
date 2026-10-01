package usecase

import (
	"context"
	"fmt"
	"time"

	"messaging/server/internal/domain"
	"messaging/server/internal/usecase/port"
)

// RegisterDevice creates a new device and auto-joins it to the General group.
// Called by POST /devices.
type RegisterDevice struct {
	Devices port.DeviceRepo
	Groups  port.GroupRepo
	Bus     port.EventBus
	Clock   port.Clock
	IDs     port.IDGenerator
}

// Input is the request payload for registering a device.
type RegisterDeviceInput struct {
	Name   string
	Avatar string
}

// Output is the result of a successful registration.
type RegisterDeviceOutput struct {
	Device domain.Device
}

// Execute runs the use case.
func (uc *RegisterDevice) Execute(ctx context.Context, in RegisterDeviceInput) (RegisterDeviceOutput, error) {
	now := uc.Clock.Now()

	device := domain.Device{
		ID:         domain.DeviceID(uc.IDs.New()),
		Name:       in.Name,
		Avatar:     in.Avatar,
		CreatedAt:  now,
		LastSeenAt: now,
	}

	if err := uc.Devices.Save(ctx, device); err != nil {
		return RegisterDeviceOutput{}, fmt.Errorf("register device: save: %w", err)
	}

	// Auto-join the General group. Idempotent — safe to call multiple times.
	if err := uc.Groups.AddMember(ctx, domain.GeneralGroupID, device.ID); err != nil {
		return RegisterDeviceOutput{}, fmt.Errorf("register device: add to general: %w", err)
	}

	// Broadcast the device_joined event so all connected clients update
	// their people strip without polling.
	_ = uc.Bus.Publish(ctx, domain.Event{
		Type:       domain.EventDeviceJoined,
		OccurredAt: now,
		Payload: domain.DeviceJoinedPayload{
			Device: device,
		},
	})

	return RegisterDeviceOutput{Device: device}, nil
}

// DefaultTTL is how long an undelivered message lives in the mailbox.
// 30 days in production; shortened in demo mode via config.
const DefaultTTL = 30 * 24 * time.Hour
