# Aegis

Local Kubernetes platform lab on kind.

It runs a small app (Pulse), an ingress, autoscaling, Kyverno policies, Grafana, and Argo CD GitOps. The cluster lives in Docker on your laptop. App requests stay on localhost. Outbound use is image pulls, Helm chart repos, and Argo CD cloning this GitHub repository.

Repo: https://github.com/Demosprojectstests/aegis

## What it demonstrates

- kind cluster with ingress-nginx
- Pulse: gateway, orders, worker, Redis
- Worker calls Redis (`redis-cli -h redis ping` → `PONG`)
- Horizontal Pod Autoscaler on gateway (metrics-server)
- Kyverno ValidatingPolicies scoped to namespace `pulse`
- kube-prometheus-stack / Grafana
- Argo CD app-of-apps synced from this repo

## What it is not

- Not a production cluster
- Not exposed as a public SaaS
- No tenant operator
- No custom `services/` application code
- Gateway and orders still use podinfo as HTTP stubs
- GitHub has source files only, not live cluster data or host public IPs

## Layout

```text
aegis/
├── Makefile
├── README.md
├── cluster/
│   ├── kind.yaml          # 1 control-plane + 2 workers, host ports 80/443
│   ├── bootstrap.sh       # idempotent platform install
│   └── loadgen.yaml       # HPA demo pod
├── charts/pulse/          # Helm chart Argo syncs
│   ├── Chart.yaml
│   ├── values.yaml
│   └── templates/
│       ├── gateway.yaml
│       ├── orders.yaml
│       ├── worker.yaml    # redis-cli loop
│       ├── redis.yaml
│       ├── ingress.yaml
│       └── hpa.yaml
└── gitops/
    ├── root-app.yaml      # Argo Application -> path gitops
    ├── kustomization.yaml
    ├── apps/apps.yaml     # Argo Application -> charts/pulse
    └── platform/policies/
        ├── disallow-latest.yaml
        └── require-resources.yaml
