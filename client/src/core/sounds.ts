/**
 * Lightweight Web Audio API sound effects — no external files required.
 * All sounds are synthesised on the fly.
 */

let ctx: AudioContext | null = null;

function getCtx(): AudioContext {
  if (!ctx) ctx = new AudioContext();
  // Resume in case the browser suspended it (autoplay policy)
  if (ctx.state === 'suspended') ctx.resume();
  return ctx;
}

/** WhatsApp-style incoming message "pop" — two short tones */
export function playReceive() {
  try {
    const ac = getCtx();
    const now = ac.currentTime;

    const freqs = [880, 1100];
    freqs.forEach((freq, i) => {
      const osc = ac.createOscillator();
      const gain = ac.createGain();

      osc.connect(gain);
      gain.connect(ac.destination);

      osc.type = 'sine';
      osc.frequency.setValueAtTime(freq, now + i * 0.09);

      gain.gain.setValueAtTime(0, now + i * 0.09);
      gain.gain.linearRampToValueAtTime(0.18, now + i * 0.09 + 0.01);
      gain.gain.exponentialRampToValueAtTime(0.0001, now + i * 0.09 + 0.12);

      osc.start(now + i * 0.09);
      osc.stop(now + i * 0.09 + 0.13);
    });
  } catch {
    // Silently ignore if audio is not available
  }
}

/** Soft keyboard "tick" for typing */
export function playTyping() {
  try {
    const ac = getCtx();
    const now = ac.currentTime;

    const bufferSize = Math.floor(ac.sampleRate * 0.04);
    const buffer = ac.createBuffer(1, bufferSize, ac.sampleRate);
    const data = buffer.getChannelData(0);
    for (let i = 0; i < bufferSize; i++) {
      data[i] = (Math.random() * 2 - 1) * (1 - i / bufferSize) * 0.25;
    }

    const source = ac.createBufferSource();
    const gain = ac.createGain();
    const filter = ac.createBiquadFilter();

    source.buffer = buffer;
    filter.type = 'bandpass';
    filter.frequency.value = 3200;
    filter.Q.value = 0.8;

    source.connect(filter);
    filter.connect(gain);
    gain.connect(ac.destination);

    gain.gain.setValueAtTime(0.22, now);
    gain.gain.exponentialRampToValueAtTime(0.0001, now + 0.04);

    source.start(now);
    source.stop(now + 0.05);
  } catch {
    // Silently ignore if audio is not available
  }
}

/** Notification chime for background-conversation banner */
export function playNotification() {
  try {
    const ac = getCtx();
    const now = ac.currentTime;

    const notes = [1047, 1319, 1568]; // C6, E6, G6 — a gentle major arpeggio
    notes.forEach((freq, i) => {
      const osc = ac.createOscillator();
      const gain = ac.createGain();

      osc.connect(gain);
      gain.connect(ac.destination);

      osc.type = 'triangle';
      osc.frequency.setValueAtTime(freq, now + i * 0.1);

      gain.gain.setValueAtTime(0, now + i * 0.1);
      gain.gain.linearRampToValueAtTime(0.12, now + i * 0.1 + 0.01);
      gain.gain.exponentialRampToValueAtTime(0.0001, now + i * 0.1 + 0.35);

      osc.start(now + i * 0.1);
      osc.stop(now + i * 0.1 + 0.36);
    });
  } catch {
    // Silently ignore if audio is not available
  }
}
