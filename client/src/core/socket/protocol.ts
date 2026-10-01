// Defines the canonical WebSocket protocol structures as mandated by the Go backend.

export interface Envelope {
  type: string;
  id: string;
  ts: string;
  payload: any;
}

export interface HelloPayload {
  device_id: string;
  last_seq: number;
}

export interface HelloOkPayload {
  connection_id: string;
  server_time: string;
}

export interface SendPayload {
  to: string;
  text: string;
}

export interface AckServerPayload {
  ref_id: string;
}

export interface AckDevicePayload {
  message_id: string;
}

export interface ReadPayload {
  message_id: string;
}

export interface DeliverPayload {
  message_id: string;
  conversation_id: string;
  from: string;
  text: string;
  sent_at: string;
  server_at: string;
  seq: number;
}

export interface SyncBatchPayload {
  messages: DeliverPayload[];
  has_more: boolean;
}

export interface PresencePayload {
  device_id: string;
  status: 'online' | 'offline';
}

export interface DeviceJoinedPayload {
  device_id: string;
  name: string;
}
