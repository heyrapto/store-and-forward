import { useEffect, useState } from 'react';
import { useAppStore } from '../../core/store';
import { Activity, Server, Smartphone, Mailbox } from 'lucide-react';
import { format } from 'date-fns';

interface LogEvent {
  id: string;
  type: string;
  time: string;
  data: any;
}

export function GodView() {
  const directory = useAppStore(state => state.directory);
  const presence = useAppStore(state => state.presence);
  const outbox = useAppStore(state => state.outbox);
  
  const [logs, setLogs] = useState<LogEvent[]>([]);

  useEffect(() => {
    const sse = new EventSource('/admin/events');
    
    const handleEvent = (type: string) => (e: MessageEvent) => {
      setLogs(prev => [{
        id: crypto.randomUUID(),
        type,
        time: new Date().toISOString(),
        data: JSON.parse(e.data)
      }, ...prev].slice(0, 50));
    };

    sse.addEventListener('device.online', handleEvent('device.online'));
    sse.addEventListener('device.offline', handleEvent('device.offline'));
    sse.addEventListener('message.queued', handleEvent('message.queued'));
    sse.addEventListener('delivery.acked', handleEvent('delivery.acked'));
    sse.addEventListener('delivery.read', handleEvent('delivery.read'));
    sse.addEventListener('device.joined', handleEvent('device.joined'));
    sse.addEventListener('mailbox.drained', handleEvent('mailbox.drained'));

    return () => sse.close();
  }, []);

  return (
    <div className="w-[450px] bg-slate-900 text-slate-300 h-full flex flex-col shrink-0 font-mono text-sm border-l border-slate-700 shadow-xl overflow-hidden">
      <div className="p-4 bg-slate-800 border-b border-slate-700 flex items-center gap-2 text-slate-100 font-bold tracking-wider uppercase text-xs">
        <Activity size={16} className="text-wa-light" />
        God View Dashboard
      </div>

      <div className="flex-1 overflow-y-auto p-4 space-y-6">
        
        {/* Gateways Panel (Mocked for Ep1) */}
        <section>
          <h3 className="text-slate-500 font-semibold mb-2 flex items-center gap-2 text-xs">
            <Server size={14} /> Gateways (Episode 1)
          </h3>
          <div className="bg-slate-800 rounded p-3 border border-slate-700 flex justify-between items-center">
            <div>
              <div className="text-green-400 font-bold">gw-local-1</div>
              <div className="text-xs text-slate-500">Standalone (Memory Bus)</div>
            </div>
            <div className="text-right">
              <div className="text-xl font-light text-slate-200">
                {Object.values(presence).filter(p => p === 'online').length}
              </div>
              <div className="text-[10px] text-slate-500 uppercase">Connections</div>
            </div>
          </div>
        </section>

        {/* Sessions & Mailboxes Panel */}
        <section>
          <h3 className="text-slate-500 font-semibold mb-2 flex items-center gap-2 text-xs">
            <Smartphone size={14} /> Sessions & Mailbox Depth
          </h3>
          <div className="bg-slate-800 rounded border border-slate-700 divide-y divide-slate-700">
            {Object.values(directory).length === 0 && (
              <div className="p-3 text-slate-500 italic text-xs text-center">No devices registered</div>
            )}
            {Object.values(directory).map((d, i) => {
              if (!d?.id) return null;
              const isOnline = presence[d.id] === 'online';
              const pendingCount = (outbox[d.id] || []).filter(m => m.state === 'pending').length;
              
              return (
                <div key={d.id || i} className="p-2 px-3 flex items-center justify-between">
                  <div className="flex items-center gap-2">
                    <div className={`w-2 h-2 rounded-full ${isOnline ? 'bg-green-500 shadow-[0_0_8px_rgba(34,197,94,0.6)]' : 'bg-slate-600'}`} />
                    <span className="text-slate-200">{d.name || 'Unknown'}</span>
                  </div>
                  <div className="flex items-center gap-4 text-xs">
                    <span className="text-slate-500">{d.id.substring(d.id.length - 6)}</span>
                    <div className="flex items-center gap-1 w-12 justify-end" title="Pending items in Outbox (Local) or Server Mailbox">
                      <Mailbox size={12} className={pendingCount > 0 ? 'text-amber-400' : 'text-slate-600'} />
                      <span className={pendingCount > 0 ? 'text-amber-400 font-bold' : 'text-slate-600'}>
                        {pendingCount}
                      </span>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
        </section>

        {/* Event Log */}
        <section>
          <h3 className="text-slate-500 font-semibold mb-2 flex items-center gap-2 text-xs">
            <Activity size={14} /> Live Event Log
          </h3>
          <div className="bg-slate-950 rounded border border-slate-800 p-2 h-[300px] overflow-y-auto space-y-1">
            {logs.length === 0 && (
              <div className="text-slate-600 italic text-xs p-2">Waiting for events...</div>
            )}
            {logs.map(log => (
              <div key={log.id} className="text-[10px] leading-tight font-mono">
                <span className="text-slate-500">[{format(new Date(log.time), 'HH:mm:ss.SSS')}]</span>{' '}
                <span className={
                  log.type.includes('online') ? 'text-green-400' :
                  log.type.includes('offline') ? 'text-red-400' :
                  log.type.includes('queued') ? 'text-blue-400' :
                  log.type.includes('acked') ? 'text-emerald-400' :
                  'text-purple-400'
                }>{log.type}</span>{' '}
                <span className="text-slate-400">{JSON.stringify(log.data)}</span>
              </div>
            ))}
          </div>
        </section>

      </div>
    </div>
  );
}
