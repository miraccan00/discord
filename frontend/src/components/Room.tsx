import { useEffect, useRef, useState } from 'react';
import { MeshManager } from '../webrtc/MeshManager';
import { useStore } from '../state/store';
import { ParticipantList } from './ParticipantList';
import { Controls } from './Controls';
import { TextChat } from './TextChat';

export function Room() {
  const token = useStore((s) => s.auth.token);
  const username = useStore((s) => s.auth.username);
  const clearAuth = useStore((s) => s.clearAuth);
  const resetParticipants = useStore((s) => s.resetParticipants);

  const meshRef = useRef<MeshManager | null>(null);
  const [joined, setJoined] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Tear down the mesh if the component unmounts while connected.
  useEffect(() => {
    return () => {
      meshRef.current?.stop();
      meshRef.current = null;
    };
  }, []);

  // Joining requires a user gesture so getUserMedia and the AudioContext are
  // allowed by the browser's autoplay policy.
  async function join() {
    if (!token) return;
    setError(null);
    try {
      const mesh = new MeshManager();
      await mesh.start(token);
      meshRef.current = mesh;
      setJoined(true);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'could not access microphone');
    }
  }

  function leave() {
    meshRef.current?.stop();
    meshRef.current = null;
    resetParticipants();
    setJoined(false);
  }

  function logout() {
    leave();
    clearAuth();
  }

  return (
    <div className="room">
      <header className="room-header">
        <span className="room-title"># general</span>
        <span className="spacer" />
        <span className="who">{username}</span>
        <button className="ghost" onClick={logout}>
          Log out
        </button>
      </header>

      {!joined ? (
        <div className="join-gate">
          <button className="primary" onClick={join}>
            🎙️ Join voice
          </button>
          {error && <p className="error">{error}</p>}
          <p className="hint">Use headphones to avoid echo when testing two tabs.</p>
        </div>
      ) : (
        <div className="room-body">
          <ParticipantList />
          <TextChat mesh={meshRef.current} />
          <Controls mesh={meshRef.current} onLeave={leave} />
        </div>
      )}
    </div>
  );
}
