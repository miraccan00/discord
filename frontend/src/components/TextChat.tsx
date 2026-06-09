import { useEffect, useRef, useState, type FormEvent } from 'react';
import { MeshManager } from '../webrtc/MeshManager';
import { useStore } from '../state/store';

export function TextChat({ mesh }: { mesh: MeshManager | null }) {
  const chat = useStore((s) => s.chat);
  const selfId = useStore((s) => s.auth.selfId);
  const [text, setText] = useState('');
  const bottomRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [chat.length]);

  function onSubmit(e: FormEvent) {
    e.preventDefault();
    const trimmed = text.trim();
    if (!trimmed) return;
    mesh?.sendChat(trimmed);
    setText('');
  }

  return (
    <section className="chat">
      <h2>Chat</h2>
      <div className="messages">
        {chat.map((m, i) => (
          <div key={`${m.id}-${m.ts}-${i}`} className={`msg${m.id === selfId ? ' own' : ''}`}>
            <span className="msg-name">{m.name}</span>
            <span className="msg-text">{m.text}</span>
          </div>
        ))}
        <div ref={bottomRef} />
      </div>
      <form className="chat-input" onSubmit={onSubmit}>
        <input
          value={text}
          onChange={(e) => setText(e.target.value)}
          placeholder="Message #general"
          maxLength={500}
        />
        <button type="submit">Send</button>
      </form>
    </section>
  );
}
