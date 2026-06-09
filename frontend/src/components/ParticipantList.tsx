import { useStore } from '../state/store';
import { ParticipantTile } from './ParticipantTile';

export function ParticipantList() {
  const participants = useStore((s) => s.participants);
  const list = Object.values(participants).sort((a, b) => a.name.localeCompare(b.name));

  return (
    <section className="participants">
      <h2>In voice — {list.length}</h2>
      <div className="tiles">
        {list.map((p) => (
          <ParticipantTile key={p.id} participant={p} />
        ))}
      </div>
    </section>
  );
}
