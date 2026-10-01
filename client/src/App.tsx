import { useEffect } from 'react';
import { Plus, Smartphone } from 'lucide-react';
import { useAppStore } from './core/store';
import { PhoneMock } from './features/devices/PhoneMock';
import { MonitoringView } from './features/inspector/MonitoringView';

const AVATARS = ['🦊', '🐰', '🐼', '🐯', '🦁', '🐮', '🐷', '🐸', '🐙', '🦖'];
const NAMES = ['Alice', 'Bob', 'Charlie', 'Diana', 'Eve', 'Frank', 'Grace', 'Heidi', 'Ivan', 'Judy'];

function App() {
  const directory = useAppStore(state => state.directory);
  const setDirectory = useAppStore(state => state.setDirectory);
  const addDeviceToDirectory = useAppStore(state => state.addDeviceToDirectory);

  // Fetch initial devices on load
  useEffect(() => {
    fetch('/devices')
      .then(r => r.json())
      .then(data => {
        if (Array.isArray(data)) setDirectory(data);
      })
      .catch(console.error);
  }, [setDirectory]);

  const spawnDevice = async () => {
    const avatar = AVATARS[Math.floor(Math.random() * AVATARS.length)];
    const name = NAMES[Math.floor(Math.random() * NAMES.length)];
    try {
      const res = await fetch('/devices', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name, avatar })
      });
      const data = await res.json();
      if (data?.id) addDeviceToDirectory(data);
    } catch (e) {
      console.error('Failed to spawn device', e);
    }
  };

  const devices = Object.values(directory).filter(d => d?.id);

  return (
    <div className="flex h-screen w-screen overflow-hidden" style={{ background: '#f0f2f5' }}>

      {/* Phones Area */}
      <div className="flex-1 flex flex-col min-w-0">
        <header className="bg-white px-6 py-4 shadow-sm z-10 flex justify-between items-center shrink-0 border-b border-gray-200">
          <div>
            <h1 className="text-xl font-bold text-gray-900">System Design: Store and Forward</h1>
            <p className="text-sm text-gray-500 mt-0.5">Episode 1 · Core Mechanics</p>
          </div>
          <button
            onClick={spawnDevice}
            className="flex items-center gap-2 px-4 py-2 rounded-lg font-semibold text-sm text-white transition-colors"
            style={{ backgroundColor: '#075E54' }}
          >
            <Plus size={16} />
            Spawn Device
          </button>
        </header>

        <main className="flex-1 overflow-x-auto overflow-y-hidden">
          {devices.length === 0 ? (
            <div className="h-full flex flex-col items-center justify-center text-gray-400 gap-3">
              <Smartphone size={56} strokeWidth={1.5} />
              <p className="text-base font-medium">No devices yet</p>
              <p className="text-sm">Click "Spawn Device" to create a simulated phone</p>
            </div>
          ) : (
            <div className="flex gap-10 items-center h-full px-10 py-8 w-max">
              {devices.map(device => (
                <PhoneMock key={device.id} device={device} />
              ))}
            </div>
          )}
        </main>
      </div>

      {/* God View */}
      <MonitoringView />
    </div>
  );
}

export default App;
