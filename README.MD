# SAP AI Core Homelab — Task Checklist (Docker Desktop)

This document is a copy/paste task list to guide implementation of a Kubernetes + Helm + Observability homelab using Docker Desktop Kubernetes.

---

## 0) One-time prerequisites

- [ ] Install/verify tools:
  - [ ] Docker Desktop
  - [ ] Enable Kubernetes in Docker Desktop (Settings → Kubernetes → Enable)
  - [ ] `kubectl` installed and working
  - [ ] `helm` installed and working
  - [ ] (Optional) `hey` for load testing (`brew install hey` on macOS)

- [ ] Verify cluster is reachable:
  - [ ] `kubectl config use-context docker-desktop`
  - [ ] `kubectl get nodes` (should show at least 1 node)

---

## 1) Repo bootstrap

- [ ] Create baseline folders in repo:
  - [ ] `apps/`
  - [ ] `deploy/`
  - [ ] `docs/`
  - [ ] `scripts/`

- [ ] Add this file as: `docs/tasks.md`

- [ ] Create namespaces manifests:
  - [ ] `deploy/namespaces/apps.yaml`
  - [ ] `deploy/namespaces/observability.yaml`

- [ ] Apply namespaces:
  - [ ] `kubectl apply -f deploy/namespaces/apps.yaml`
  - [ ] `kubectl apply -f deploy/namespaces/observability.yaml`

---

## 2) Build the minimal "inference-api" app

**Goal:** a tiny HTTP service you can deploy, observe, and break.

### App requirements
- [ ] Expose `GET /healthz` → returns `200 ok`
- [ ] Expose `GET /predict` → returns JSON
- [ ] Env var: `LATENCY_MS` (simulate latency)
- [ ] Env var: `ERROR_RATE` (0.0–1.0, simulate failures)
- [ ] Log one line per request (structured preferred)

### Implement
- [ ] Create app folder: `apps/inference-api/`
- [ ] Implement service in **Go or Python** (your choice)
- [ ] Add Dockerfile
- [ ] Build image locally:
  - [ ] `docker build -t inference-api:local apps/inference-api`

### Local test (no Kubernetes)
- [ ] Run locally:
  - [ ] `docker run --rm -p 8080:8080 -e LATENCY_MS=50 -e ERROR_RATE=0.0 inference-api:local`
- [ ] Test endpoints:
  - [ ] `curl http://localhost:8080/healthz`
  - [ ] `curl http://localhost:8080/predict`

---

## 3) Deploy using plain Kubernetes YAML (before Helm)

**Goal:** learn K8s primitives and debugging first.

### Create manifests (under `deploy/k8s/inference-api/`)
- [ ] `deployment.yaml`
  - [ ] 2 replicas
  - [ ] resource requests/limits
  - [ ] readinessProbe on `/healthz`
  - [ ] livenessProbe on `/healthz`
  - [ ] env vars `LATENCY_MS`, `ERROR_RATE`
- [ ] `service.yaml`
  - [ ] ClusterIP service exposing HTTP
- [ ] (Optional) `ingress.yaml` (you can do this later, Step 4)

### Deploy
- [ ] `kubectl apply -n apps -f deploy/k8s/inference-api/`

### Verify & debug practice
- [ ] `kubectl -n apps get pods -w`
- [ ] `kubectl -n apps describe pod <pod>`
- [ ] `kubectl -n apps logs <pod> -f`

### Break & fix drills (must do)
- [ ] Break readiness probe path → observe no traffic → fix
- [ ] Set `ERROR_RATE=0.5` → observe HTTP 500s → fix
- [ ] Set `LATENCY_MS=1500` → observe slower responses → fix
- [ ] Set too-low memory limit → observe OOMKilled → fix

---

## 4) Install Ingress Controller + external access

**Goal:** simulate real traffic path: client → ingress → service → pod.

### Install NGINX Ingress Controller
- [ ] `helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx`
- [ ] `helm repo update`
- [ ] `helm install ingress-nginx ingress-nginx/ingress-nginx -n ingress-nginx --create-namespace`

### Add Ingress for the app (choose one)
- [ ] Option A: plain YAML `deploy/k8s/inference-api/ingress.yaml`
- [ ] Option B: later via Helm template

### Local host mapping
- [ ] Add to `/etc/hosts`:
  - [ ] `127.0.0.1 inference.local`

### Validate external access
- [ ] `curl http://inference.local/healthz`
- [ ] `curl http://inference.local/predict`

### Networking debug drills (must do)
- [ ] Break Service selector → traffic fails → debug/fix
- [ ] Break Ingress backend service name → traffic fails → debug/fix

---

## 5) Convert to Helm chart (Senior Helm expectations)

**Goal:** parameterize deployments safely and support environments.

### Create chart
- [ ] Create: `deploy/helm/inference-api/`
  - [ ] `Chart.yaml`
  - [ ] `values.yaml`
  - [ ] `values-dev.yaml`
  - [ ] `values-prod.yaml`
  - [ ] `templates/`
    - [ ] deployment
    - [ ] service
    - [ ] ingress
    - [ ] (optional) HPA
    - [ ] `_helpers.tpl`

### Move config to values
- [ ] image repo/tag/pullPolicy
- [ ] env vars
- [ ] resources
- [ ] replicas
- [ ] ingress host/path/class

### Install/upgrade
- [ ] `helm upgrade --install inference-api deploy/helm/inference-api -n apps -f deploy/helm/inference-api/values-dev.yaml`

### Helm drills (must do)
- [ ] Make a safe change (env var tweak
