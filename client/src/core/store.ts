import { create } from 'zustand';


export type MessageState = 'pending' | 'sent' | 'delivered' | 'read';

export interface Message {
  id: string;
  conversation_id: string;
  from: string;
  text: string;
  sent_at: string;
  server_at?: string;
  state: MessageState;
  seq?: number;
}

export interface Device {
  id: string;
  name: string;
  avatar: string;
  created_at: string;
  last_seen_at: string;
}

interface AppState {
  // Global directory
  directory: Record<string, Device>;
  presence: Record<string, 'online' | 'offline'>;
  
  // Mocks active state
  activeDevices: Record<string, boolean>;

  // Per-device localized databases
  messages: Record<string, Message[]>; // deviceId -> Messages
  outbox: Record<string, Message[]>;   // deviceId -> Outbox

  // Actions
  setDirectory: (devices: Device[]) => void;
  addDeviceToDirectory: (device: Device) => void;
  deleteDevice: (deviceId: string) => void;
  setPresence: (deviceId: string, status: 'online' | 'offline') => void;
  toggleDevice: (deviceId: string) => void;
  
  // Localized Actions
  addMessage: (ownerId: string, msg: Message) => void;
  updateMessageState: (ownerId: string, messageId: string, state: MessageState) => void;
  clearOutbox: (ownerId: string, messageId: string) => void;
}

export const useAppStore = create<AppState>((set) => ({
  directory: {},
  presence: {},
  activeDevices: {},
  messages: {},
  outbox: {},

  setDirectory: (devices) =>
    set(() => {
      const dir: Record<string, Device> = {};
      devices.forEach((d) => (dir[d.id] = d));
      return { directory: dir };
    }),

  addDeviceToDirectory: (device) =>
    set((state) => ({
      directory: { ...state.directory, [device.id]: device },
    })),

  deleteDevice: (deviceId) =>
    set((state) => {
      const { [deviceId]: _, ...restDir } = state.directory;
      const { [deviceId]: _p, ...restPresence } = state.presence;
      const { [deviceId]: _a, ...restActive } = state.activeDevices;
      const { [deviceId]: _m, ...restMessages } = state.messages;
      const { [deviceId]: _o, ...restOutbox } = state.outbox;
      return {
        directory: restDir,
        presence: restPresence,
        activeDevices: restActive,
        messages: restMessages,
        outbox: restOutbox,
      };
    }),

  setPresence: (deviceId, status) =>
    set((state) => ({
      presence: { ...state.presence, [deviceId]: status },
    })),

  toggleDevice: (deviceId) =>
    set((state) => {
      const current = state.activeDevices[deviceId] || false;
      return {
        activeDevices: { ...state.activeDevices, [deviceId]: !current },
      };
    }),

  addMessage: (ownerId, msg) =>
    set((state) => {
      const devMsgs = state.messages[ownerId] || [];
      // Dedupe by ID
      if (devMsgs.some(m => m.id === msg.id)) return state;

      const newMsgs = [...devMsgs, msg].sort(
        (a, b) => new Date(a.sent_at).getTime() - new Date(b.sent_at).getTime()
      );

      const isPending = msg.state === 'pending';
      const devOutbox = state.outbox[ownerId] || [];
      
      return {
        messages: { ...state.messages, [ownerId]: newMsgs },
        outbox: isPending
          ? { ...state.outbox, [ownerId]: [...devOutbox, msg] }
          : state.outbox,
      };
    }),

  updateMessageState: (ownerId, messageId, msgState) =>
    set((state) => {
      const devMsgs = state.messages[ownerId] || [];
      const updated = devMsgs.map((m) =>
        m.id === messageId && msgState !== m.state 
          ? { ...m, state: msgState as MessageState } 
          : m
      );
      return {
        messages: { ...state.messages, [ownerId]: updated },
      };
    }),

  clearOutbox: (ownerId, messageId) =>
    set((state) => {
      const devOutbox = state.outbox[ownerId] || [];
      return {
        outbox: {
          ...state.outbox,
          [ownerId]: devOutbox.filter((m) => m.id !== messageId),
        },
      };
    }),
}));
