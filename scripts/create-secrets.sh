#!/usr/bin/env bash
# Run this script ONCE before "helm install" to create the JWT secret.
# The secret is created (or updated) outside Helm so the value never
# appears in Helm release history or ArgoCD application manifests.
#
# Usage:
#   ./scripts/create-secrets.sh
#   NAMESPACE=my-ns ./scripts/create-secrets.sh
set -euo pipefail

NAMESPACE="${NAMESPACE:-discord}"
SECRET_NAME="discord-backend"

# ── 1. Ensure namespace exists ────────────────────────────────────────────────
echo "[1/3] Namespace: $NAMESPACE"
kubectl create namespace "$NAMESPACE" --dry-run=client -o yaml | kubectl apply -f -

# ── 2. Generate a 256-bit hex JWT secret ──────────────────────────────────────
echo "[2/3] Generating JWT secret …"
JWT_SECRET=$(openssl rand -hex 32)

# ── 3. Create (or replace) the K8s secret ─────────────────────────────────────
echo "[3/3] Applying secret '$SECRET_NAME' …"
kubectl create secret generic "$SECRET_NAME" \
  --namespace "$NAMESPACE" \
  --from-literal=JWT_SECRET="$JWT_SECRET" \
  --dry-run=client -o yaml | kubectl apply -f -

echo ""
echo "✓ Done. Now run:"
echo ""
echo "  helm install discord deploy/helm/discord \\"
echo "    --namespace $NAMESPACE \\"
echo "    --set backend.existingSecret=$SECRET_NAME \\"
echo "    --set ingress.host=<your-host>"
echo ""
echo "  # ArgoCD: set backend.existingSecret=$SECRET_NAME in application.yaml"
