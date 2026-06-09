# Discord Voice

A web-based, **voice-first** chat app optimized for low latency. Voice runs over
**WebRTC in a peer-to-peer mesh** — the Go backend only does signaling (room
membership + SDP/ICE relay), so audio never passes through the server. Includes
simple text chat in the room.

> **Why Go (not Django)?** Browser voice is WebRTC (Opus/UDP/SRTP). Go handles
> many concurrent WebSocket/UDP connections with low GC overhead; a sync-first
> Python stack is a poor fit for real-time signaling.

## Architecture

```
 Browser A  ───────────  audio (P2P, Opus)  ───────────  Browser B
     │  \                                              /  │
     │   \── WS signaling (offer/answer/ICE) ─┐  ┌────/   │
     │                                        ▼  ▼        │
     └──────────────  Go signaling hub (:8080)  ──────────┘
                       (room state, relay only)
```

- **backend/** — Go 1.25. WebSocket signaling hub (`coder/websocket`), JWT login
  with two hardcoded users. No database, no media.
- **frontend/** — React + TypeScript + Vite. `MeshManager` creates one
  `RTCPeerConnection` per peer using the perfect-negotiation pattern; speaking
  indicators via the Web Audio API.
- **deploy/helm/discord/** — one Helm chart that deploys both components.
- **.github/workflows/** — CI (lint/test/build) and CD (push images to Docker Hub).

## Demo users

| Username | Password   |
|----------|------------|
| `alice`  | `alice123` |
| `bob`    | `bob123`   |

---

## Local development

Two terminals:

```bash
# 1) backend  → http://localhost:8080
make backend-run

# 2) frontend → http://localhost:5173  (proxies /api and /ws to the backend)
cd frontend && npm install && npm run dev
```

Open **two browser windows** at <http://localhost:5173> (localhost is a secure
context, so the microphone works without TLS). Log in as `alice` and `bob`, click
**Join voice**, and grant mic permission. Use headphones to avoid echo.

Verify in `chrome://webrtc-internals` that a P2P connection is established with
Opus and STUN (`srflx`) candidates — confirming no media flows through the
backend.

---

## Kubernetes deployment

### Prerequisites

| Tool | Minimum version | Purpose |
|---|---|---|
| `kubectl` | 1.28 | cluster access |
| `helm` | 3.14 | chart install |
| `openssl` | any | secret generation in pre-install script |
| Traefik | v2 or v3 | ingress controller |

### Step 1 — GitHub repository

Create a **public** repo at `github.com/miraccan00/discord` and push:

```bash
git init
git add .
git commit -m "initial: discord voice app"
git remote add origin https://github.com/miraccan00/discord.git
git push -u origin master
```

### Step 2 — GitHub Actions secrets

Go to **Settings → Secrets and variables → Actions** in the GitHub repo and add:

| Secret name | Value |
|---|---|
| `DOCKERHUB_USERNAME` | `miraccan` |
| `DOCKERHUB_PASSWORD` | Docker Hub hesap şifresi |

CI runs automatically on every push to `master`. To publish images:

```bash
git tag v0.1.0 && git push origin v0.1.0
```

This triggers the CD workflow and pushes `miraccan/discord-backend:0.1.0` and
`miraccan/discord-frontend:0.1.0` to Docker Hub.

### Step 3 — Pre-install secret

The JWT signing secret must exist in the cluster **before** `helm install`.
The script generates a 256-bit random secret and creates the Kubernetes Secret
so the value never appears in Helm release history or git:

```bash
chmod +x scripts/create-secrets.sh

# Default namespace is "discord". Override with NAMESPACE=my-ns.
./scripts/create-secrets.sh
```

The script outputs the exact `helm install` command to run next.

### Step 4 — Helm install

```bash
helm install discord deploy/helm/discord \
  --namespace discord \
  --set backend.existingSecret=discord-backend \
  --set ingress.host=discord.example.com \
  --set ingress.tls.enabled=true \
  --set ingress.tls.secretName=discord-tls
```

> **`ALLOWED_ORIGINS` is derived automatically** from `ingress.host` and
> `ingress.tls.enabled`. You do not need to set it separately — changing
> `ingress.host` is enough to keep backend CORS in sync.

> **Microphone needs HTTPS off-localhost.** `getUserMedia` only works in a secure
> context. For any real deployment set `ingress.tls.enabled=true` with a valid
> certificate.
>
> **Backend is single-replica by design** — room state is in-memory. Scaling out
> would require shared pub/sub (e.g. Redis).

### ConfigMap / Secret auto-reload

The backend Deployment carries `checksum/config` and `checksum/secret` annotations
computed from the rendered ConfigMap and Secret. When either value changes (e.g.
`helm upgrade --set backend.env.jwtTtlMinutes=120`), the hash rotates, mutating
the pod template, which triggers a rolling restart automatically — no manual
`kubectl rollout restart` needed.

### Step 5 — ArgoCD (optional, GitOps)

If you use ArgoCD, apply the Application definition after running the pre-install
script:

```bash
# Edit deploy/argocd/application.yaml first:
#   ingress.host → your real hostname
kubectl apply -f deploy/argocd/application.yaml
```

ArgoCD pulls the chart directly from `github.com/miraccan00/discord` at `HEAD`
(or pin `targetRevision` to a tag for production). It references the pre-created
`discord-backend` secret via `backend.existingSecret` — no plain-text credentials
in git.

### Helm values reference

| Key | Default | Description |
|---|---|---|
| `backend.existingSecret` | `""` | Pre-created K8s secret name. Takes priority over `jwtSecret`. |
| `backend.jwtSecret` | `change-me-in-prod` | Only used when `existingSecret` is empty (dev only). |
| `backend.env.allowedOrigins` | `""` | CORS origin. Derived from `ingress.host` + TLS when empty. |
| `backend.env.jwtTtlMinutes` | `"60"` | JWT expiry in minutes. |
| `ingress.className` | `traefik` | IngressClass name. |
| `ingress.host` | `discord.local` | Public hostname. |
| `ingress.tls.enabled` | `false` | Enable TLS on the ingress. |
| `ingress.tls.secretName` | `""` | K8s TLS secret name (cert-manager or manual). |
| `backend.image.tag` | *(appVersion)* | Image tag. Defaults to `Chart.AppVersion`. |
| `frontend.image.tag` | *(appVersion)* | Image tag. |

---

## CI/CD

- **CI** (`push`/`PR` to `master`): backend `golangci-lint` → `go test -race` →
  build; frontend typecheck/lint/build; Docker build smoke test for both images.
- **CD** (tag `v*`): build & push both images to Docker Hub (semver + `sha-` tags),
  then `helm lint --strict`.

Required GitHub secrets: **`DOCKERHUB_USERNAME`**, **`DOCKERHUB_PASSWORD`**.

---

## Signaling protocol

JSON envelope `{ type, from, to, payload }` over `/ws?token=<jwt>`:

| Direction | Types |
|-----------|-------|
| client → server | `join`, `offer`, `answer`, `ice-candidate`, `chat-message`, `mute-state` |
| server → client | `room-state`, `peer-joined`, `peer-left`, `offer`, `answer`, `ice-candidate`, `chat-message`, `mute-state`, `error` |

The server stamps `from` authoritatively and relays directed messages
(`offer`/`answer`/`ice-candidate`) only to the addressed peer.
