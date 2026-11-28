# Local Kubernetes with kind

Runs the whole platform on a 3-node cluster inside Docker, then serves it on http://localhost.

```bash
# 1. Cluster (each node is a Docker container)
kind create cluster --config deploy/kind/cluster.yaml

# 2. Add-ons: ingress controller + metrics for the HPA
helm repo add traefik https://traefik.github.io/charts
helm repo add metrics-server https://kubernetes-sigs.github.io/metrics-server/
helm upgrade --install traefik traefik/traefik --version 43.0.0 \
  -n traefik --create-namespace -f deploy/kind/traefik-values.yaml
helm upgrade --install metrics-server metrics-server/metrics-server --version 3.14.0 \
  -n kube-system -f deploy/kind/metrics-server-values.yaml

# 3. Images: build with compose, then copy them into the kind nodes (no registry needed)
docker compose build
for app in frontend todo-service user-service; do
  kind load docker-image todo-platform/$app:dev --name todo
done

# 4. The app
helm upgrade --install todo gitops/charts/todo-platform \
  -n todo-dev --create-namespace -f gitops/charts/todo-platform/values-dev.yaml --wait

# Clean up
kind delete cluster --name todo
```

After rebuilding an image, load it again and restart the pods:
`kubectl -n todo-dev rollout restart deployment`.
