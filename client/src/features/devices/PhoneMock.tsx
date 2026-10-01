import { useState, useRef, useEffect } from 'react';
import { Clock, Check, CheckCheck, Send, Power, Trash2 } from 'lucide-react';
import { useAppStore, type Device } from '../../core/store';
import { useDeviceSocket } from '../../core/socket/useDeviceSocket';
import { format } from 'date-fns';
import { AnimatePresence, motion } from 'framer-motion';

const EMPTY_MSGS: any[] = [];

export function PhoneMock({ device }: { device: Device }) {
  const [activeChat, setActiveChat] = useState<string>('grp:general');
  const [text, setText] = useState('');
  const chatRef = useRef<HTMLDivElement>(null);

  const { send, connected, markRead } = useDeviceSocket(device.id);

  const isActive    = useAppStore(s => s.activeDevices[device.id] ?? false);
  const toggleDevice  = useAppStore(s => s.toggleDevice);
  const deleteDevice  = useAppStore(s => s.deleteDevice);
  const directory   = useAppStore(s => s.directory);
  const presence    = useAppStore(s => s.presence);
  const messages    = useAppStore(s => s.messages[device.id] ?? EMPTY_MSGS);

  const [notification, setNotification] = useState<{ id: string; title: string; body: string } | null>(null);

  const chatMessages = messages.filter(m => m.conversation_id === activeChat);

  // Auto-scroll
  useEffect(() => {
    chatRef.current?.scrollTo({ top: chatRef.current.scrollHeight, behavior: 'smooth' });
  }, [chatMessages.length]);

  // Mark messages read when chat is open
  useEffect(() => {
    if (!isActive || !connected) return;
    chatMessages.forEach(m => {
      if (m.from !== device.id && m.state === 'delivered') markRead(m.id);
    });
  }, [chatMessages, isActive, connected]);

  // Notification banner for background conversations
  const lastMsg = messages[messages.length - 1];
  useEffect(() => {
    if (!lastMsg || !isActive) return;
    if (lastMsg.from === device.id) return;
    if (lastMsg.state !== 'delivered') return;
    if (lastMsg.conversation_id === activeChat) return;

    const sender = directory[lastMsg.from];
    setNotification({
      id: lastMsg.id,
      title: lastMsg.conversation_id === 'grp:general'
        ? `General · ${sender?.name ?? 'Unknown'}`
        : sender?.name ?? 'Unknown',
      body: lastMsg.text,
    });
    const t = setTimeout(() => setNotification(null), 4000);
    return () => clearTimeout(t);
  }, [lastMsg?.id]);

  const handleSend = (e: React.FormEvent) => {
    e.preventDefault();
    if (!text.trim()) return;
    const to = activeChat === 'grp:general'
      ? 'general'
      : activeChat.replace('dm:', '').split(':').find(id => id !== device.id) ?? '';
    send(to, text.trim());
    setText('');
  };

  const handleDelete = async () => {
    try {
      await fetch(`/devices/${device.id}`, { method: 'DELETE' });
    } catch {/* ignore */}
    deleteDevice(device.id);
  };

  const getChatName = (convId: string) => {
    if (convId === 'grp:general') return 'General';
    const otherId = convId.replace('dm:', '').split(':').find(id => id !== device.id) ?? '';
    return directory[otherId]?.name ?? 'Unknown';
  };

  const peers = Object.values(directory).filter(d => d.id !== device.id);

  return (
    <div
      className="flex flex-col shrink-0 overflow-hidden relative"
      style={{
        width: 300,
        height: 620,
        borderRadius: 36,
        border: '10px solid #1a1a1a',
        boxShadow: '0 25px 60px rgba(0,0,0,0.35), inset 0 0 0 2px #333',
        background: '#E5DDD5',
        fontFamily: '-apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif',
      }}
    >
      {/* Notification banner */}
      <AnimatePresence>
        {notification && (
          <motion.div
            initial={{ y: -80, opacity: 0 }}
            animate={{ y: 8, opacity: 1 }}
            exit={{ y: -80, opacity: 0 }}
            transition={{ type: 'spring', stiffness: 300, damping: 28 }}
            onClick={() => setNotification(null)}
            className="absolute left-2 right-2 z-50 cursor-pointer"
            style={{
              background: 'rgba(255,255,255,0.96)',
              backdropFilter: 'blur(12px)',
              borderRadius: 14,
              padding: '10px 12px',
              boxShadow: '0 4px 20px rgba(0,0,0,0.15)',
            }}
          >
            <div style={{ fontSize: 11, fontWeight: 700, color: '#075E54', marginBottom: 2 }}>
              {notification.title}
            </div>
            <div style={{ fontSize: 12, color: '#333', overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
              {notification.body}
            </div>
          </motion.div>
        )}
      </AnimatePresence>

      {/* Header */}
      <div style={{ background: '#075E54', padding: '10px 12px', display: 'flex', alignItems: 'center', gap: 10, flexShrink: 0 }}>
        <div style={{
          width: 36, height: 36, borderRadius: '50%',
          background: 'rgba(255,255,255,0.2)',
          display: 'flex', alignItems: 'center', justifyContent: 'center',
          fontSize: 18, flexShrink: 0,
        }}>
          {device.avatar}
        </div>
        <div style={{ flex: 1, minWidth: 0 }}>
          <div style={{ color: '#fff', fontWeight: 600, fontSize: 14, lineHeight: 1.2, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
            {device.name}
          </div>
          <div style={{ color: 'rgba(255,255,255,0.65)', fontSize: 11 }}>
            {isActive ? (connected ? '🟢 Online' : '⟳ Connecting…') : '⚫ Offline'}
          </div>
        </div>
        <button
          onClick={() => toggleDevice(device.id)}
          title={isActive ? 'Go Offline' : 'Go Online'}
          style={{
            background: isActive ? 'rgba(239,68,68,0.25)' : 'rgba(34,197,94,0.25)',
            border: 'none', borderRadius: 8, padding: '5px 7px',
            color: isActive ? '#fca5a5' : '#86efac',
            cursor: 'pointer', display: 'flex', alignItems: 'center',
          }}
        >
          <Power size={15} />
        </button>
        <button
          onClick={handleDelete}
          title="Delete device"
          style={{
            background: 'rgba(239,68,68,0.15)',
            border: 'none', borderRadius: 8, padding: '5px 7px',
            color: '#fca5a5', cursor: 'pointer', display: 'flex', alignItems: 'center',
          }}
        >
          <Trash2 size={15} />
        </button>
      </div>

      {/* People strip */}
      <div style={{ background: '#fff', padding: '8px 10px', display: 'flex', gap: 10, overflowX: 'auto', flexShrink: 0, borderBottom: '1px solid #e9edef' }}>
        {/* General */}
        <button
          onClick={() => setActiveChat('grp:general')}
          style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 3, background: 'none', border: 'none', cursor: 'pointer', opacity: activeChat === 'grp:general' ? 1 : 0.5, minWidth: 44 }}
        >
          <div style={{ width: 40, height: 40, borderRadius: '50%', background: '#e9edef', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 18 }}>🌍</div>
          <span style={{ fontSize: 10, color: '#3b4a54', fontWeight: 500 }}>General</span>
        </button>

        {peers.map(peer => {
          const convId = `dm:${[device.id, peer.id].sort().join(':')}`;
          const isOnline = presence[peer.id] === 'online';
          const unread = messages.filter(m => m.conversation_id === convId && m.from !== device.id && m.state === 'delivered').length;
          return (
            <button
              key={peer.id}
              onClick={() => setActiveChat(convId)}
              style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', gap: 3, background: 'none', border: 'none', cursor: 'pointer', opacity: activeChat === convId ? 1 : 0.55, minWidth: 44, position: 'relative' }}
            >
              <div style={{ position: 'relative' }}>
                <div style={{ width: 40, height: 40, borderRadius: '50%', background: '#e9edef', display: 'flex', alignItems: 'center', justifyContent: 'center', fontSize: 18 }}>
                  {peer.avatar}
                </div>
                <div style={{ position: 'absolute', bottom: 1, right: 0, width: 11, height: 11, borderRadius: '50%', background: isOnline ? '#25D366' : '#8696a0', border: '2px solid #fff' }} />
                {unread > 0 && (
                  <div style={{ position: 'absolute', top: -4, right: -4, background: '#25D366', color: '#fff', fontSize: 9, fontWeight: 700, borderRadius: '50%', width: 16, height: 16, display: 'flex', alignItems: 'center', justifyContent: 'center', border: '1.5px solid #fff' }}>
                    {unread}
                  </div>
                )}
              </div>
              <span style={{ fontSize: 10, color: '#3b4a54', fontWeight: 500, maxWidth: 44, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap' }}>
                {peer.name.split(' ')[0]}
              </span>
            </button>
          );
        })}
      </div>

      {/* Chat name bar */}
      <div style={{ background: '#f0f2f5', padding: '6px 14px', fontSize: 12, fontWeight: 600, color: '#54656f', flexShrink: 0 }}>
        {getChatName(activeChat)}
      </div>

      {/* Messages */}
      <div
        ref={chatRef}
        style={{ flex: 1, overflowY: 'auto', padding: '8px 10px', display: 'flex', flexDirection: 'column', gap: 4 }}
      >
        {chatMessages.length === 0 && (
          <div style={{ margin: 'auto', color: '#8696a0', fontSize: 12, textAlign: 'center', padding: 20 }}>
            No messages yet.<br />Say hello! 👋
          </div>
        )}

        {chatMessages.map(msg => {
          const isMe = msg.from === device.id;
          const sender = directory[msg.from];

          return (
            <div key={msg.id} style={{ display: 'flex', flexDirection: 'column', alignItems: isMe ? 'flex-end' : 'flex-start' }}>
              {!isMe && activeChat === 'grp:general' && (
                <span style={{ fontSize: 10, color: '#075E54', fontWeight: 600, marginLeft: 10, marginBottom: 2 }}>
                  {sender?.name ?? 'Unknown'}
                </span>
              )}
              <div style={{
                maxWidth: '82%',
                background: isMe ? '#DCF8C6' : '#fff',
                borderRadius: isMe ? '12px 0 12px 12px' : '0 12px 12px 12px',
                padding: '6px 10px 16px 10px',
                position: 'relative',
                boxShadow: '0 1px 2px rgba(0,0,0,0.08)',
                minWidth: 60,
              }}>
                <p style={{ margin: 0, fontSize: 13, color: '#111b21', lineHeight: 1.4, wordBreak: 'break-word' }}>
                  {msg.text}
                </p>
                <div style={{ position: 'absolute', bottom: 3, right: 7, display: 'flex', alignItems: 'center', gap: 2 }}>
                  <span style={{ fontSize: 10, color: '#8696a0' }}>
                    {format(new Date(msg.sent_at), 'HH:mm')}
                  </span>
                  {isMe && (
                    <span style={{ color: msg.state === 'read' ? '#53bdeb' : '#8696a0' }}>
                      {msg.state === 'pending'   && <Clock size={11} />}
                      {msg.state === 'sent'      && <Check size={11} />}
                      {(msg.state === 'delivered' || msg.state === 'read') && <CheckCheck size={11} />}
                    </span>
                  )}
                </div>
              </div>
            </div>
          );
        })}
      </div>

      {/* Composer */}
      <form onSubmit={handleSend} style={{ background: '#f0f2f5', padding: '6px 8px', display: 'flex', gap: 6, alignItems: 'center', flexShrink: 0 }}>
        <input
          type="text"
          value={text}
          onChange={e => setText(e.target.value)}
          placeholder={isActive ? 'Type a message' : 'Offline — messages queued'}
          style={{
            flex: 1, borderRadius: 24, border: 'none', padding: '8px 14px',
            fontSize: 13, background: '#fff', outline: 'none',
            color: '#111b21',
          }}
        />
        <button
          type="submit"
          disabled={!text.trim()}
          style={{
            width: 36, height: 36, borderRadius: '50%', border: 'none',
            background: text.trim() ? '#075E54' : '#8696a0',
            color: '#fff', cursor: text.trim() ? 'pointer' : 'default',
            display: 'flex', alignItems: 'center', justifyContent: 'center',
            transition: 'background 0.15s', flexShrink: 0,
          }}
        >
          <Send size={16} style={{ marginLeft: 2 }} />
        </button>
      </form>

      {/* Offline dimmer */}
      {!isActive && (
        <div style={{ position: 'absolute', inset: 0, background: 'rgba(0,0,0,0.04)', pointerEvents: 'none', borderRadius: 26 }} />
      )}
    </div>
  );
}
