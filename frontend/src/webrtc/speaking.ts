// SpeakingDetector watches one or more audio MediaStreams and reports when each
// crosses a loudness threshold, driving the "speaking" ring on participant
// tiles. It uses a single AudioContext and a single rAF loop for all streams.

interface Entry {
  analyser: AnalyserNode;
  source: MediaStreamAudioSourceNode;
  data: Uint8Array<ArrayBuffer>;
  speaking: boolean;
  // Timestamp until which the tile stays "speaking" after audio drops, to avoid
  // flicker between words (hangover).
  hangoverUntil: number;
}

const SPEAKING_RMS = 0.04; // empirical threshold on normalized RMS [0,1]
const HANGOVER_MS = 250;

export class SpeakingDetector {
  private ctx: AudioContext;
  private entries = new Map<string, Entry>();
  private raf = 0;
  private readonly onChange: (id: string, speaking: boolean) => void;

  constructor(ctx: AudioContext, onChange: (id: string, speaking: boolean) => void) {
    this.ctx = ctx;
    this.onChange = onChange;
  }

  add(id: string, stream: MediaStream): void {
    if (this.entries.has(id)) this.remove(id);
    const source = this.ctx.createMediaStreamSource(stream);
    const analyser = this.ctx.createAnalyser();
    analyser.fftSize = 512;
    source.connect(analyser);
    this.entries.set(id, {
      analyser,
      source,
      data: new Uint8Array(analyser.fftSize),
      speaking: false,
      hangoverUntil: 0,
    });
    if (this.raf === 0) this.loop();
  }

  remove(id: string): void {
    const entry = this.entries.get(id);
    if (!entry) return;
    entry.source.disconnect();
    entry.analyser.disconnect();
    this.entries.delete(id);
  }

  stop(): void {
    if (this.raf !== 0) cancelAnimationFrame(this.raf);
    this.raf = 0;
    for (const id of [...this.entries.keys()]) this.remove(id);
  }

  private loop = (): void => {
    const now = performance.now();
    for (const [id, entry] of this.entries) {
      entry.analyser.getByteTimeDomainData(entry.data);
      let sumSquares = 0;
      for (const sample of entry.data) {
        const centered = (sample - 128) / 128;
        sumSquares += centered * centered;
      }
      const rms = Math.sqrt(sumSquares / entry.data.length);

      if (rms > SPEAKING_RMS) entry.hangoverUntil = now + HANGOVER_MS;
      const speaking = now < entry.hangoverUntil;

      if (speaking !== entry.speaking) {
        entry.speaking = speaking;
        this.onChange(id, speaking);
      }
    }
    this.raf = requestAnimationFrame(this.loop);
  };
}
