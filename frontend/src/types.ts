// Wire protocol shared with the Go signaling backend. The envelope mirrors
// internal/signaling/message.go.

export type SignalType =
  | 'join'
  | 'offer'
  | 'answer'
  | 'ice-candidate'
  | 'chat-message'
  | 'mute-state'
  | 'room-state'
  | 'peer-joined'
  | 'peer-left'
  | 'error';

export interface SignalMessage<P = unknown> {
  type: SignalType;
  from?: string;
  to?: string;
  payload?: P;
}

export interface PeerInfo {
  id: string;
  name: string;
  muted: boolean;
  deafened: boolean;
}

export interface RoomStatePayload {
  selfId: string;
  selfName: string;
  peers: PeerInfo[];
}

export interface ChatPayload {
  id: string;
  name: string;
  text: string;
  ts: number;
}

export interface OfferAnswerPayload {
  sdp: RTCSessionDescriptionInit;
}

export interface IceCandidatePayload {
  candidate: RTCIceCandidateInit;
}

// UI-facing participant state (self included).
export interface Participant {
  id: string;
  name: string;
  muted: boolean;
  deafened: boolean;
  speaking: boolean;
  isSelf: boolean;
}

export interface ChatMessage {
  id: string;
  name: string;
  text: string;
  ts: number;
}
