import { create } from 'zustand';
import type { ChatMessage, Participant } from '../types';

// The store holds only serializable UI state. Non-serializable WebRTC objects
// (RTCPeerConnection, MediaStream, AudioContext) live in MeshManager.

interface AuthState {
  token: string | null;
  username: string | null;
  selfId: string | null;
}

interface AppState {
  auth: AuthState;
  participants: Record<string, Participant>;
  chat: ChatMessage[];
  muted: boolean;
  deafened: boolean;

  setAuth: (token: string, username: string) => void;
  clearAuth: () => void;
  setSelfId: (id: string) => void;

  upsertParticipant: (p: Participant) => void;
  removeParticipant: (id: string) => void;
  patchParticipant: (id: string, patch: Partial<Participant>) => void;
  resetParticipants: () => void;

  addChat: (m: ChatMessage) => void;
  setMuted: (muted: boolean) => void;
  setDeafened: (deafened: boolean) => void;
}

const TOKEN_KEY = 'discord.token';
const USER_KEY = 'discord.username';

export const useStore = create<AppState>((set) => ({
  auth: {
    token: localStorage.getItem(TOKEN_KEY),
    username: localStorage.getItem(USER_KEY),
    selfId: null,
  },
  participants: {},
  chat: [],
  muted: false,
  deafened: false,

  setAuth: (token, username) => {
    localStorage.setItem(TOKEN_KEY, token);
    localStorage.setItem(USER_KEY, username);
    set((s) => ({ auth: { ...s.auth, token, username } }));
  },
  clearAuth: () => {
    localStorage.removeItem(TOKEN_KEY);
    localStorage.removeItem(USER_KEY);
    set({ auth: { token: null, username: null, selfId: null }, participants: {}, chat: [] });
  },
  setSelfId: (id) => set((s) => ({ auth: { ...s.auth, selfId: id } })),

  upsertParticipant: (p) =>
    set((s) => ({ participants: { ...s.participants, [p.id]: p } })),
  removeParticipant: (id) =>
    set((s) => {
      const next = { ...s.participants };
      delete next[id];
      return { participants: next };
    }),
  patchParticipant: (id, patch) =>
    set((s) => {
      const existing = s.participants[id];
      if (!existing) return {};
      return { participants: { ...s.participants, [id]: { ...existing, ...patch } } };
    }),
  resetParticipants: () => set({ participants: {} }),

  addChat: (m) => set((s) => ({ chat: [...s.chat, m] })),
  setMuted: (muted) => set({ muted }),
  setDeafened: (deafened) => set({ deafened }),
}));
