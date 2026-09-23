.PHONY: up down sync demo-hpa demo-policy pods

up:
    ./cluster/bootstrap.sh

down:
    kind delete cluster --name aegis

sync:
    kubectl apply -f gitops/root-app.yaml
    kubectl -n argocd get applications

pods:
    kubectl get nodes
    kubectl get pods -A

demo-policy:
    kubectl -n pulse run bad-latest --image=nginx:latest || true

demo-hpa:
    kubectl apply -f cluster/loadgen.yaml
    kubectl -n pulse get hpa gateway --watch
