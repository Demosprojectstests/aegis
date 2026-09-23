cat > ~/aegis/README.md <<'EOF'
# Aegis

Local Kubernetes platform lab on kind.

It includes a small Go order pipeline (Pulse), ingress, HPA, Kyverno, Grafana, Argo CD GitOps, and a Kubebuilder Tenant operator that provisions isolated namespaces.

The cluster runs on your laptop. App requests stay on localhost. Outbound use is image pulls, Helm repos, Go module/tool downloads, and Argo cloning this GitHub repository.

Repo: https://github.com/Demosprojectstests/aegis

## What it demonstrates

- kind cluster with ingress-nginx
- Pulse written in Go: gateway → orders → Redis list → worker
- HPA on gateway (metrics-server)
- Kyverno ValidatingPolicies in namespace `pulse`
- kube-prometheus-stack / Grafana
- Argo CD app-of-apps (root + Pulse)
- Tenant operator (`platform.aegis.dev/v1`) running in-cluster

cat > ~/aegis/README.md <<'ENDREADME'
# Aegis

Local Kubernetes platform lab on kind.

Pulse is a small Go order pipeline. The cluster also has ingress, HPA,
Kyverno, Grafana, Argo CD, and a Kubebuilder Tenant operator.

Everything runs on the laptop. App traffic stays on localhost.
Outbound use is image pulls, Helm/Go module downloads, and Argo
cloning this repo.

https://github.com/Demosprojectstests/aegis

## What is in this repo

- cluster/ — kind config, bootstrap, HPA loadgen
- charts/pulse/ — Helm chart (local images pulse-*:dev)
- services/ — Go gateway, orders, worker, Redis helper, Dockerfiles
- gitops/ — Argo root app, Pulse app, Kyverno policies
- operator/ — Tenant CRD + controller (Kubebuilder)

## Pipeline

    POST /orders
      -> ingress :80
      -> gateway :8080
      -> orders :8080
      -> LPUSH pulse:orders
      -> worker BRPOP
      -> log processed

Build and load images:

    cd services
    docker build -f gateway/Dockerfile -t pulse-gateway:dev .
    docker build -f orders/Dockerfile  -t pulse-orders:dev .
    docker build -f worker/Dockerfile  -t pulse-worker:dev .
    kind load docker-image pulse-gateway:dev --name aegis
    kind load docker-image pulse-orders:dev  --name aegis
    kind load docker-image pulse-worker:dev  --name aegis

Try it:

    curl -s -H "Host: pulse.127.0.0.1.nip.io" http://127.0.0.1/
    curl -s -H "Host: pulse.127.0.0.1.nip.io" -X POST http://127.0.0.1/orders -d '{"item":"widget"}'
    kubectl -n pulse logs deploy/worker --tail=10

Expected: {"service":"pulse-gateway"}, {"id":"ord-...","queued":true},
and worker log processed ...

Chart uses imagePullPolicy Never. If Argo syncs an old podinfo chart,
pause the Pulse Application before applying local changes.

## Platform

    make up

Installs kind aegis if needed, ingress-nginx, metrics-server, Argo CD,
Kyverno, kube-prometheus-stack, policies, and the Argo root app.

    make demo-policy
    make demo-hpa

Grafana:

    kubectl -n monitoring port-forward svc/kube-prom-stack-grafana 3000:80

Login admin / aegis.

VPN note: Nord can break kubectl to 127.0.0.1 with TLS handshake
timeout. Disconnect or whitelist 127.0.0.0/8 and 172.16.0.0/12.
Pulse on port 80 can still work.

## Tenant operator

    apiVersion: platform.aegis.dev/v1
    kind: Tenant
    metadata:
      name: acme
      namespace: default
    spec:
      plan: starter

Creates tenant-acme, ResourceQuota, NetworkPolicy.
Delete the CR to remove the namespace.

    cd operator
    make docker-build IMG=aegis-operator:dev
    kind load docker-image aegis-operator:dev --name aegis
    make deploy IMG=aegis-operator:dev
    kubectl -n operator-system get pods
    kubectl get tenant

In-cluster manager lives in operator-system.
Local debug: make install && make run.

## GitOps

Argo root watches gitops/. Pulse app watches charts/pulse.
Push only when Git matches what you want the cluster to run.

## Makefile

    make up
    make sync
    make pods
    make demo-policy
    make demo-hpa
    make down

## Privacy

Do not commit kubeconfig, secrets, or host public IPs.

## Teardown

    make down
ENDREADME

head -n 8 ~/aegis/README.md
## Layout

```text
aegis/
├── Makefile
├── README.md
├── cluster/                 # kind, bootstrap, loadgen
├── charts/pulse/            # Helm chart for Pulse
├── gitops/                  # Argo root app, Pulse app, Kyverno policies
├── services/                # Go source + Dockerfiles
│   ├── gateway/
│   ├── orders/
│   ├── worker/
│   └── internal/redisx/
└── operator/                # Kubebuilder Tenant operator
