# Design Direction (UI/UX)

This document outlines the design direction for the frontend application, which serves to visualize the offline messaging architecture in real-time.

## Visual Concept
The UI is a unified desktop canvas representing a global view of the system.
- **The Playground**: A horizontal row of realistic phone mockups. Each phone is a completely independent simulated client/device.
- **The God-View**: A wider administrative panel (dashboard) adjacent to or above/below the phones, offering full visibility into the backend state.

## Phone Mock Anatomy
Each phone simulator is an interactive mini-app:
- **Top Bar**: Displays the device name, a unique avatar, and an "Active" toggle.
- **People Strip**: A horizontally scrolling list of active chats.
  - The `General` group is pinned first.
  - Followed by all other devices represented as avatars with live presence dots.
- **Chat Area**: The main conversation view displaying message bubbles.
- **Composer**: Text input field and send button. If the phone is offline (Active=false), the composer is disabled or shows a visual warning.

## People Strip
- **Navigation**: Horizontal scroll. Tap an avatar to switch the chat view to that DM or group.
- **Presence Status**: A dot on the avatar (Green = Online, Grey = Offline).
- **Unread Badge**: A red badge with a counter for unread messages per conversation.

## Message Bubbles
Inspired by modern chat apps (e.g., WhatsApp).
- **Sent Messages**: Right-aligned, Dark Teal bubble (`#075E54`), white text.
- **Received Messages**: Left-aligned, Light Grey bubble, dark text.
- **Delivery Status Ticks** (Sent messages only):
  - ⏱ `Clock`: Queued locally, device is offline.
  - ✓ `One Tick`: Sent to server.
  - ✓✓ `Two Ticks`: Delivered to recipient device.
  - ✓✓ `Two Blue Ticks`: Read by recipient (optional extension).
- **Timestamps**: Display both `sent_at` (client clock) and `server_at` (server authoritative time) for debugging clock skew scenarios.

## Notification Banners
- When a message is received in a chat that is *not* currently open, a banner slides down from the top of the phone screen.
- Shows the sender's avatar, name, and a text preview.
- Auto-dismisses after 4 seconds.
- Stacks vertically if multiple notifications arrive quickly.
- Updates the unread badge on the people strip.

## Active Toggle
The core interactive element for testing architecture resilience.
- **State: ON (Active)**: WebSocket is connected. Phone is fully opaque. When turned on, the client instantly reconnects, drains its mailbox, and banners fire rapidly for missed messages.
- **State: OFF (Inactive)**: WebSocket is explicitly closed. Phone dims (greys out slightly). Messages sent from this state sit in the outbox with a ⏱ icon.

## God-View Dashboard
A global command center for visualizing the backend cluster.
- **Gateways Panel**: Lists all active gateways (`gw-1`, `gw-2`). Shows current WebSocket connection count and health status.
- **Sessions Panel**: A mapping of which device ID is currently connected to which Gateway ID.
- **Mailbox Panel**: Shows pending (undelivered) message counts per device queue.
- **Event Log**: A live SSE (Server-Sent Events) stream displaying backend actions (e.g., "Message Saved", "Redis Pub", "Ack Received"). Events are color-coded by type.
- **Wire Tap (Episode 4)**: A real-time inspector for raw WebSocket frames and HTTP payloads passing through the system.

## Color Palette + Typography
- **Primary / Sent Bubble**: Dark Teal (`#075E54`)
- **Secondary / Received Bubble**: Light Grey (`#F0F0F0`)
- **Online Indicator**: Bright Green
- **Offline Indicator**: Slate Grey
- **Background**: Soft dark or neutral canvas for the God-view to make the phones pop.
- **Typography**: `Inter` or standard `system-ui`. Clean, highly legible sans-serif.

## Animations
Powered by **Framer Motion** for fluid transitions:
- **Banner Slide-down**: Smooth spring physics entry from the top bar.
- **Presence Dot Pulse**: Subtle pulsing animation when presence state toggles.
- **Message Appear**: New message bubbles slide up from the bottom while fading in.
