import type { Participant } from '../types';

export function ParticipantTile({ participant }: { participant: Participant }) {
  const { name, speaking, muted, deafened, isSelf } = participant;
  const initial = name.charAt(0).toUpperCase();

  return (
    <div className={`tile${speaking ? ' speaking' : ''}`}>
      <div className="avatar">{initial}</div>
      <div className="tile-name">
        {name}
        {isSelf && <span className="you"> (you)</span>}
      </div>
      <div className="tile-icons">
        {deafened ? <span title="Deafened">🔇</span> : muted ? <span title="Muted">🚫</span> : null}
      </div>
    </div>
  );
}
