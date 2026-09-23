#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

if ! kind get clusters | grep -qx aegis; then
  kind create cluster --config "${ROOT}/cluster/kind.yaml"
else
  echo "kind cluster 'aegis' already exists"
fi

kubectl apply -f https://kind.sigs.k8s.io/examples/ingress/deploy-ingress-nginx.yaml
echo "waiting for ingress-nginx..."
kubectl -n ingress-nginx wait --for=condition=ready pod \
  -l app.kubernetes.io/component=controller --timeout=180s

kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml
kubectl -n kube-system patch deploy metrics-server --type='json' -p='[
  {"op":"add","path":"/spec/template/spec/containers/0/args/-","value":"--kubelet-insecure-tls"}
]' || true

helm repo add argo https://argoproj.github.io/argo-helm
helm repo add kyverno https://kyverno.github.io/kyverno/
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm repo update

if helm status argocd -n argocd >/dev/null 2>&1; then
  echo "Argo CD already installed, skipping helm upgrade"
else
  helm upgrade --install argocd argo/argo-cd \
    --namespace argocd --create-namespace \
    --set server.service.type=ClusterIP \
    --set redisSecretInit.enabled=false \
    --timeout 10m \
    --wait
fi

if helm status kyverno -n kyverno >/dev/null 2>&1; then
  echo "Kyverno already installed"
else
  helm upgrade --install kyverno kyverno/kyverno \
    --namespace kyverno --create-namespace \
    --wait
fi

if helm status kube-prom-stack -n monitoring >/dev/null 2>&1; then
  echo "kube-prometheus-stack already installed"
else
  helm upgrade --install kube-prom-stack prometheus-community/kube-prometheus-stack \
    --namespace monitoring --create-namespace \
    --set grafana.adminPassword=aegis \
    --wait
fi

kubectl apply -f "${ROOT}/gitops/platform/policies/"
kubectl apply -f "${ROOT}/gitops/root-app.yaml" || true

echo
echo "Argo CD admin password:"
kubectl -n argocd get secret argocd-initial-admin-secret \
  -o jsonpath='{.data.password}' | base64 -d 2>/dev/null || echo "(secret already rotated)"
echo
echo "Grafana: kubectl -n monitoring port-forward svc/kube-prom-stack-grafana 3000:80"
echo "  user admin  password aegis"
echo "Pulse: curl -s -H 'Host: pulse.127.0.0.1.nip.io' http://127.0.0.1"
