// ICE configuration for peer connections. Public Google STUN servers are enough
// for most networks. Restrictive/symmetric NATs would additionally need a TURN
// relay (e.g. coturn) — left as a documented future addition.
export const iceConfig: RTCConfiguration = {
  iceServers: [
    {
      urls: [
        'stun:stun.l.google.com:19302',
        'stun:stun1.l.google.com:19302',
      ],
    },
    // Example TURN entry (disabled by default):
    // { urls: 'turn:turn.example.com:3478', username: 'user', credential: 'pass' },
  ],
};
