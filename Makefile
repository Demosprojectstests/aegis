.PHONY: up down sync demo-hpa demo-policy pods

.PHONY: images pulse-images operator-images operator-deploy

pulse-images:
    docker build -f services/gateway/Dockerfile -t pulse-gateway:dev services
    docker build -f services/orders/Dockerfile  -t pulse-orders:dev  services
    docker build -f services/worker/Dockerfile  -t pulse-worker:dev  services
    kind load docker-image pulse-gateway:dev --name aegis
    kind load docker-image pulse-orders:dev  --name aegis
    kind load docker-image pulse-worker:dev  --name aegis

operator-images:
    $(MAKE) -C operator docker-build IMG=aegis-operator:dev
    kind load docker-image aegis-operator:dev --name aegis

images: pulse-images operator-images

operator-deploy:
    $(MAKE) -C operator deploy IMG=aegis-operator:dev
    kubectl -n operator-system patch deploy operator-controller-manager --type json -p='[{"op":"replace","path":"/spec/template/spec/containers/0/imagePullPolicy","value":"Never"}]' || true

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
