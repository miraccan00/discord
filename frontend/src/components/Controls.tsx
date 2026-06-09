import { MeshManager } from '../webrtc/MeshManager';
import { useStore } from '../state/store';

interface Props {
  mesh: MeshManager | null;
  onLeave: () => void;
}

export function Controls({ mesh, onLeave }: Props) {
  const muted = useStore((s) => s.muted);
  const deafened = useStore((s) => s.deafened);

  return (
    <div className="controls">
      <button
        className={muted ? 'active' : ''}
        onClick={() => mesh?.setMuted(!muted)}
        disabled={deafened}
        title="Toggle microphone"
      >
        {muted ? '🚫 Unmute' : '🎙️ Mute'}
      </button>
      <button
        className={deafened ? 'active' : ''}
        onClick={() => mesh?.setDeafened(!deafened)}
        title="Toggle incoming audio"
      >
        {deafened ? '🔇 Undeafen' : '🎧 Deafen'}
      </button>
      <button className="danger" onClick={onLeave} title="Leave voice">
        📞 Leave
      </button>
    </div>
  );
}
