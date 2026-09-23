# Aegis

Local Kubernetes platform lab: GitOps-ready cluster, a multi-service app, autoscaling, policy-as-code, and observability.

This runs on a laptop with kind. Nothing is exposed off the machine except image pulls and optional public DNS lookups.

## What it demonstrates

- kind cluster with ingress-nginx
- Argo CD installed
- Pulse app: gateway, orders, worker, redis
- Horizontal Pod Autoscaler with metrics-server
- Kyverno ValidatingPolicies
- kube-prometheus-stack / Grafana

## Layout

```text
aegis/
├── cluster/                 # kind config + bootstrap
├── charts/pulse/            # demo application
├── gitops/                  # Argo CD app-of-apps (next)
├── operator/                # Tenant operator (next)
├── services/                # custom app code (later)
└── docs/
