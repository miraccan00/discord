import type { SignalMessage, SignalType } from '../types';

type Handler = (msg: SignalMessage) => void;

// SignalingClient is a thin typed wrapper around the native WebSocket. It
// connects to /ws with the JWT as a query parameter (browsers cannot set
// WebSocket headers) and dispatches inbound messages by type.
export class SignalingClient {
  private ws: WebSocket | null = null;
  private handlers = new Map<SignalType, Handler>();
  private onOpen?: () => void;
  private onClose?: () => void;

  connect(token: string, callbacks: { onOpen?: () => void; onClose?: () => void } = {}): void {
    this.onOpen = callbacks.onOpen;
    this.onClose = callbacks.onClose;

    const proto = window.location.protocol === 'https:' ? 'wss' : 'ws';
    const url = `${proto}://${window.location.host}/ws?token=${encodeURIComponent(token)}`;
    const ws = new WebSocket(url);
    this.ws = ws;

    ws.onopen = () => this.onOpen?.();
    ws.onclose = () => this.onClose?.();
    ws.onmessage = (ev: MessageEvent<string>) => {
      let msg: SignalMessage;
      try {
        msg = JSON.parse(ev.data) as SignalMessage;
      } catch {
        return;
      }
      this.handlers.get(msg.type)?.(msg);
    };
  }

  on(type: SignalType, handler: Handler): void {
    this.handlers.set(type, handler);
  }

  send<P>(type: SignalType, payload?: P, to?: string): void {
    if (this.ws?.readyState !== WebSocket.OPEN) return;
    const msg: SignalMessage<P> = { type, payload, to };
    this.ws.send(JSON.stringify(msg));
  }

  close(): void {
    this.ws?.close();
    this.ws = null;
  }
}
