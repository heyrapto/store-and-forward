import { useEffect, useRef, useState } from 'react';
import { useAppStore } from '../store';
import type { Envelope, HelloPayload, SendPayload } from './protocol';

const WS_URL = 'ws://localhost:8080/ws';

const EMPTY_ARRAY: any[] = [];

export function useDeviceSocket(deviceId: string) {
  const wsRef = useRef<WebSocket | null>(null);
  const [connected, setConnected] = useState(false);
  const isActive = useAppStore((state) => state.activeDevices[deviceId]);
  
  const addMessage = useAppStore((state) => state.addMessage);
  const updateMessageState = useAppStore((state) => state.updateMessageState);
  const clearOutbox = useAppStore((state) => state.clearOutbox);
  const outbox = useAppStore((state) => state.outbox[deviceId] || EMPTY_ARRAY);
  const messages = useAppStore((state) => state.messages[deviceId] || EMPTY_ARRAY);


  // We need the highest seq to send in the hello frame
  const highestSeqRef = useRef<number>(0);
  useEffect(() => {
    let max = 0;
    messages.forEach(m => {
      if (m.seq && m.seq > max) max = m.seq;
    });
    highestSeqRef.current = max;
  }, [messages]);

  useEffect(() => {
    if (!isActive) {
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
        setConnected(false);
      }
      return;
    }

    const connect = () => {
      const ws = new WebSocket(WS_URL);
      wsRef.current = ws;

      ws.onopen = () => {
        // Handshake
        const hello: Envelope = {
          type: 'hello',
          id: crypto.randomUUID(),
          ts: new Date().toISOString(),
          payload: {
            device_id: deviceId,
            last_seq: highestSeqRef.current,
          } as HelloPayload
        };
        ws.send(JSON.stringify(hello));
        setConnected(true);
      };

      ws.onmessage = (event) => {
        try {
          const env = JSON.parse(event.data) as Envelope;
          
          switch (env.type) {
            case 'hello_ok':
              // Handshake complete. Flush outbox!
              outbox.forEach((msg) => {
                if (msg.state === 'pending') {
                  const sendEnv: Envelope = {
                    type: 'send',
                    id: msg.id,
                    ts: new Date().toISOString(),
                    payload: {
                      to: msg.conversation_id.startsWith('grp:') ? msg.conversation_id : msg.conversation_id.replace('dm:', '').replace(deviceId, '').replace(':', ''),
                      text: msg.text,
                    } as SendPayload
                  };
                  ws.send(JSON.stringify(sendEnv));
                }
              });
              break;

            case 'ack_server':
              // Sender receives checkmark
              const ackServer = env.payload;
              updateMessageState(deviceId, ackServer.ref_id, 'sent');
              clearOutbox(deviceId, ackServer.ref_id);
              break;

            case 'receipt':
              // Sender receives double checkmark
              const receipt = env.payload;
              updateMessageState(deviceId, receipt.message_id, receipt.status === 'read' ? 'read' : 'delivered');
              break;

            case 'deliver':
              // Recipient receives message
              const deliver = env.payload;
              
              // Only process if we don't already have it
              const exists = useAppStore.getState().messages[deviceId]?.find(m => m.id === deliver.message_id);
              if (!exists) {
                addMessage(deviceId, {
                  id: deliver.message_id,
                  conversation_id: deliver.conversation_id,
                  from: deliver.from,
                  text: deliver.text,
                  sent_at: deliver.sent_at,
                  server_at: deliver.server_at,
                  state: 'delivered', // For the recipient, it's just 'delivered' to them
                  seq: deliver.seq,
                });
              }

              // Send ack_device back to server
              const ackDevice: Envelope = {
                type: 'ack_device',
                id: crypto.randomUUID(),
                ts: new Date().toISOString(),
                payload: {
                  message_id: deliver.message_id,
                }
              };
              ws.send(JSON.stringify(ackDevice));
              break;

            case 'sync_batch':
              const sync = env.payload;
              if (sync.messages) {
                sync.messages.forEach((m: any) => {
                  addMessage(deviceId, {
                    id: m.message_id,
                    conversation_id: m.conversation_id,
                    from: m.from,
                    text: m.text,
                    sent_at: m.sent_at,
                    server_at: m.server_at,
                    state: 'delivered',
                    seq: m.seq,
                  });
                  // Ack each synced message
                  ws.send(JSON.stringify({
                    type: 'ack_device',
                    id: crypto.randomUUID(),
                    ts: new Date().toISOString(),
                    payload: { message_id: m.message_id }
                  }));
                });
              }
              break;

            case 'presence':
              useAppStore.getState().setPresence(env.payload.device_id, env.payload.status);
              break;

            case 'device_joined':
              useAppStore.getState().addDeviceToDirectory(env.payload.device);
              break;

            case 'ping':
              ws.send(JSON.stringify({
                type: 'pong',
                id: env.id,
                ts: new Date().toISOString(),
                payload: {}
              }));
              break;
          }
        } catch (e) {
          console.error("Parse error", e);
        }
      };

      ws.onclose = () => {
        setConnected(false);
        wsRef.current = null;
        // Simple reconnect backoff could go here
      };
    };

    connect();

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
    };
  }, [isActive, deviceId]);

  const send = (to: string, text: string) => {
    const msgId = crypto.randomUUID();
    const isGroup = to === 'general';
    const convId = isGroup ? 'grp:general' : `dm:${[deviceId, to].sort().join(':')}`;

    const msgObj = {
      id: msgId,
      conversation_id: convId,
      from: deviceId,
      text,
      sent_at: new Date().toISOString(),
      state: 'pending' as const,
    };

    addMessage(deviceId, msgObj);

    if (connected && wsRef.current) {
      const env: Envelope = {
        type: 'send',
        id: msgId,
        ts: new Date().toISOString(),
        payload: {
          to: isGroup ? 'grp:general' : to,
          text: text,
        }
      };
      wsRef.current.send(JSON.stringify(env));
    }
  };

  const markRead = (messageId: string) => {
    // Optimistically update local state so the unread badge clears immediately
    updateMessageState(deviceId, messageId, 'read');

    if (connected && wsRef.current) {
      wsRef.current.send(JSON.stringify({
        type: 'read',
        id: crypto.randomUUID(),
        ts: new Date().toISOString(),
        payload: { message_id: messageId }
      }));
    }
  };

  return { send, connected, markRead };
}
