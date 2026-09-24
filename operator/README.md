# Aegis Tenant operator

Kubebuilder controller for `tenants.platform.aegis.dev`.

A Tenant CR in `default` creates:

- Namespace `tenant-<name>`
- ResourceQuota `tenant-quota`
- NetworkPolicy `tenant-isolate`
- Status.ready / Status.namespace

`starter` quota: 1 CPU, 1Gi, 20 pods. `pro`: 4 CPU, 4Gi, 20 pods.

Quota and NetworkPolicy are owner-referenced to the tenant Namespace.
The Tenant finalizer deletes that Namespace on CR delete.

## Run in kind

    make docker-build IMG=aegis-operator:dev
    kind load docker-image aegis-operator:dev --name aegis
    make deploy IMG=aegis-operator:dev
    kubectl -n operator-system patch deploy operator-controller-manager \
      --type json -p='[{"op":"replace","path":"/spec/template/spec/containers/0/imagePullPolicy","value":"Never"}]'

    kubectl apply -f - <<END
    apiVersion: platform.aegis.dev/v1
    kind: Tenant
    metadata:
      name: acme
      namespace: default
    spec:
      plan: starter
    END

    kubectl get tenant
    kubectl get ns tenant-acme

Local debug: make install && make run
