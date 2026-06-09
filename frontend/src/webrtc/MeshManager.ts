import { SignalingClient } from '../signaling/SignalingClient';
import { useStore } from '../state/store';
import { iceConfig } from './iceConfig';
import { SpeakingDetector } from './speaking';
import type {
  ChatPayload,
  IceCandidatePayload,
  OfferAnswerPayload,
  PeerInfo,
  RoomStatePayload,
  SignalMessage,
} from '../types';

interface PeerConn {
  pc: RTCPeerConnection;
  polite: boolean;
  makingOffer: boolean;
  ignoreOffer: boolean;
}

const ROOM = 'general';

// MeshManager owns all non-serializable WebRTC state and keeps the zustand store
// in sync. One RTCPeerConnection is maintained per remote peer in a full mesh;
// audio flows browser-to-browser and never touches the backend.
export class MeshManager {
  private signaling = new SignalingClient();
  private peers = new Map<string, PeerConn>();
  private remoteAudio = new Map<string, HTMLAudioElement>();
  private localStream: MediaStream | null = null;
  private audioCtx: AudioContext | null = null;
  private speaking: SpeakingDetector | null = null;
  private selfId = '';
  private deafened = false;

  // start must be called from a user gesture (the Join button) so getUserMedia
  // and the AudioContext are permitted by autoplay policy.
  async start(token: string): Promise<void> {
    this.localStream = await navigator.mediaDevices.getUserMedia({
      audio: { echoCancellation: true, noiseSuppression: true, autoGainControl: true },
      video: false,
    });

    this.audioCtx = new AudioContext();
    if (this.audioCtx.state === 'suspended') await this.audioCtx.resume();
    this.speaking = new SpeakingDetector(this.audioCtx, (id, isSpeaking) => {
      useStore.getState().patchParticipant(id, { speaking: isSpeaking });
    });

    this.registerHandlers();
    this.signaling.connect(token, {
      onOpen: () => this.signaling.send('join', { room: ROOM }),
      onClose: () => this.cleanupPeers(),
    });
  }

  stop(): void {
    this.signaling.close();
    this.cleanupPeers();
    this.localStream?.getTracks().forEach((t) => t.stop());
    this.speaking?.stop();
    void this.audioCtx?.close();
    this.localStream = null;
    this.audioCtx = null;
    this.speaking = null;
  }

  setMuted(muted: boolean): void {
    const track = this.localStream?.getAudioTracks()[0];
    if (track) track.enabled = !muted;
    useStore.getState().setMuted(muted);
    this.broadcastMuteState();
  }

  // Deafen silences all remote audio and, mirroring Discord, also self-mutes.
  setDeafened(deafened: boolean): void {
    this.deafened = deafened;
    for (const audio of this.remoteAudio.values()) audio.muted = deafened;
    const store = useStore.getState();
    store.setDeafened(deafened);
    if (deafened) this.setMuted(true);
    else this.broadcastMuteState();
  }

  sendChat(text: string): void {
    this.signaling.send('chat-message', { text });
  }

  private broadcastMuteState(): void {
    const { muted, deafened } = useStore.getState();
    this.signaling.send('mute-state', { muted, deafened });
    if (this.selfId) useStore.getState().patchParticipant(this.selfId, { muted, deafened });
  }

  private registerHandlers(): void {
    this.signaling.on('room-state', (m) => this.onRoomState(m as SignalMessage<RoomStatePayload>));
    this.signaling.on('peer-joined', (m) => this.onPeerJoined(m as SignalMessage<PeerInfo>));
    this.signaling.on('peer-left', (m) => this.onPeerLeft(m as SignalMessage<{ id: string }>));
    this.signaling.on('offer', (m) => void this.onDescription(m as SignalMessage<OfferAnswerPayload>));
    this.signaling.on('answer', (m) => void this.onDescription(m as SignalMessage<OfferAnswerPayload>));
    this.signaling.on('ice-candidate', (m) =>
      void this.onCandidate(m as SignalMessage<IceCandidatePayload>),
    );
    this.signaling.on('chat-message', (m) => this.onChat(m as SignalMessage<ChatPayload>));
    this.signaling.on('mute-state', (m) => this.onMuteState(m as SignalMessage<PeerInfo>));
    this.signaling.on('error', (m) => console.warn('signaling error', m.payload));
  }

  private onRoomState(m: SignalMessage<RoomStatePayload>): void {
    const state = m.payload;
    if (!state) return;
    this.selfId = state.selfId;
    const store = useStore.getState();
    store.setSelfId(state.selfId);
    store.upsertParticipant({
      id: state.selfId,
      name: state.selfName,
      muted: false,
      deafened: false,
      speaking: false,
      isSelf: true,
    });
    if (this.localStream && this.speaking) this.speaking.add(state.selfId, this.localStream);

    for (const peer of state.peers) {
      this.addParticipant(peer);
      this.createPeer(peer.id);
    }
  }

  private onPeerJoined(m: SignalMessage<PeerInfo>): void {
    if (!m.payload) return;
    this.addParticipant(m.payload);
    this.createPeer(m.payload.id);
  }

  private onPeerLeft(m: SignalMessage<{ id: string }>): void {
    const id = m.payload?.id;
    if (!id) return;
    this.closePeer(id);
    useStore.getState().removeParticipant(id);
  }

  private async onDescription(m: SignalMessage<OfferAnswerPayload>): Promise<void> {
    const from = m.from;
    const description = m.payload?.sdp;
    if (!from || !description) return;

    const peer = this.createPeer(from);
    const { pc } = peer;
    const offerCollision =
      description.type === 'offer' && (peer.makingOffer || pc.signalingState !== 'stable');
    peer.ignoreOffer = !peer.polite && offerCollision;
    if (peer.ignoreOffer) return;

    // Modern browsers perform an implicit rollback for the polite peer here.
    await pc.setRemoteDescription(description);
    if (description.type === 'offer') {
      await pc.setLocalDescription();
      this.signaling.send('answer', { sdp: pc.localDescription }, from);
    }
  }

  private async onCandidate(m: SignalMessage<IceCandidatePayload>): Promise<void> {
    const from = m.from;
    const candidate = m.payload?.candidate;
    if (!from || !candidate) return;
    const peer = this.peers.get(from);
    if (!peer) return;
    try {
      await peer.pc.addIceCandidate(candidate);
    } catch (err) {
      if (!peer.ignoreOffer) console.warn('addIceCandidate failed', err);
    }
  }

  private onChat(m: SignalMessage<ChatPayload>): void {
    if (m.payload) useStore.getState().addChat(m.payload);
  }

  private onMuteState(m: SignalMessage<PeerInfo>): void {
    if (m.from && m.payload) {
      useStore.getState().patchParticipant(m.from, {
        muted: m.payload.muted,
        deafened: m.payload.deafened,
      });
    }
  }

  private addParticipant(peer: PeerInfo): void {
    useStore.getState().upsertParticipant({
      id: peer.id,
      name: peer.name,
      muted: peer.muted,
      deafened: peer.deafened,
      speaking: false,
      isSelf: false,
    });
  }

  // createPeer returns the existing connection or builds a new one wired for
  // perfect negotiation. Politeness is derived deterministically from the ids so
  // exactly one side is polite.
  private createPeer(peerId: string): PeerConn {
    const existing = this.peers.get(peerId);
    if (existing) return existing;

    const pc = new RTCPeerConnection(iceConfig);
    const peer: PeerConn = { pc, polite: this.selfId < peerId, makingOffer: false, ignoreOffer: false };
    this.peers.set(peerId, peer);

    const track = this.localStream?.getAudioTracks()[0];
    if (track && this.localStream) pc.addTrack(track, this.localStream);

    pc.onnegotiationneeded = async () => {
      try {
        peer.makingOffer = true;
        await pc.setLocalDescription();
        this.signaling.send('offer', { sdp: pc.localDescription }, peerId);
      } catch (err) {
        console.warn('negotiation failed', err);
      } finally {
        peer.makingOffer = false;
      }
    };
    pc.onicecandidate = ({ candidate }) => {
      if (candidate) this.signaling.send('ice-candidate', { candidate }, peerId);
    };
    pc.ontrack = ({ streams }) => {
      const [stream] = streams;
      if (stream) this.attachRemoteAudio(peerId, stream);
    };
    pc.onconnectionstatechange = () => {
      if (pc.connectionState === 'failed') pc.restartIce();
    };
    return peer;
  }

  private attachRemoteAudio(peerId: string, stream: MediaStream): void {
    let audio = this.remoteAudio.get(peerId);
    if (!audio) {
      audio = new Audio();
      audio.autoplay = true;
      this.remoteAudio.set(peerId, audio);
    }
    audio.srcObject = stream;
    audio.muted = this.deafened;
    void audio.play().catch(() => undefined);
    this.speaking?.add(peerId, stream);
  }

  private closePeer(id: string): void {
    this.peers.get(id)?.pc.close();
    this.peers.delete(id);
    const audio = this.remoteAudio.get(id);
    if (audio) {
      audio.srcObject = null;
      this.remoteAudio.delete(id);
    }
    this.speaking?.remove(id);
  }

  private cleanupPeers(): void {
    for (const id of [...this.peers.keys()]) this.closePeer(id);
  }
}
