# Kubernetes & Observability Homelab

A hands-on learning path for Kubernetes, Helm, Observability, and Infrastructure as Code.
Designed to build skills for a Senior DevOps/SRE role working with AI platforms.

---

## Learning Goals

By completing this homelab, you will be able to:
- Deploy and debug applications on Kubernetes
- Create and manage Helm charts
- Build a complete observability stack (metrics, logs, traces)
- Instrument applications with Prometheus metrics and OpenTelemetry
- Understand GitOps-style infrastructure with Crossplane

---

## Prerequisites

### Tools to Install
- [ ] **Docker Desktop** with Kubernetes enabled (Settings → Kubernetes → Enable)
- [ ] **kubectl** - Kubernetes CLI
- [ ] **helm** - Package manager for Kubernetes
- [ ] **k9s** - Terminal UI for Kubernetes (optional but highly recommended: `brew install k9s`)
- [ ] **hey** - HTTP load generator (`brew install hey`)

### Verify Setup
```bash
kubectl config use-context docker-desktop
kubectl get nodes  # Should show 1 node (docker-desktop)
helm version       # Should show v3.x
```

---

# Part 1: Docker & Container Fundamentals

**Why:** Everything in Kubernetes runs in containers. You need to understand images, layers, and how containers work.

## 1.1) Understand Docker Basics

### Concepts to Learn
- [ ] What is a container vs a VM?
- [ ] What is an image? What are layers?
- [ ] What is a Dockerfile?
- [ ] What is a multi-stage build and why use it?

### Hands-on Tasks
- [ ] Read your existing `apps/api/Dockerfile` - understand each line
- [ ] Build the image: `docker build -t api:local apps/api`
- [ ] Run it: `docker run --rm -p 8080:8080 api:local`
- [ ] Test it: `curl http://localhost:8080/api/v1/healthz`
- [ ] Inspect the image: `docker inspect api:local`
- [ ] Check image size: `docker images api:local`

### Exploration Exercises
- [ ] Run `docker history api:local` - what do you see?
- [ ] Try building without multi-stage - how much bigger is the image?
- [ ] What happens if you use `golang:1.24` as final image instead of `scratch`?

---

## 1.2) Build a Better API

**Why:** You need an app that can generate metrics, logs, and traces to observe.

### Current State
Your `apps/api/` has a basic Go API. Let's enhance it with endpoints useful for load testing.

### Endpoints to Implement
| Endpoint | Method | Purpose |
|----------|--------|---------|
| `/api/v1/healthz` | GET | Health check (already exists) |
| `/api/v1/echo` | POST | Echoes back request body - good for testing payloads |
| `/api/v1/work` | GET | Simulates CPU work - useful for testing resource limits |
| `/metrics` | GET | Prometheus metrics (Part 4) |

### Environment Variables
- [ ] `LATENCY_MS` - Artificial delay added to `/work` endpoint (default: 0)
- [ ] `FAIL_RATE` - Probability of 500 error on `/work` (0.0-1.0, default: 0.0)
- [ ] `MEMORY_MB` - Memory to allocate on `/work` calls (default: 0) - useful for OOM testing

### Implementation Tasks
- [ ] Add `POST /api/v1/echo` - returns the request body as JSON response
- [ ] Add `GET /api/v1/work` - simulates work with configurable latency/errors
- [ ] Add env var parsing with sensible defaults
- [ ] Add structured logging (use `slog` package in Go)
- [ ] Log each request: method, path, status, duration

### Test Your Changes
```bash
docker build -t api:local apps/api
docker run --rm -p 8080:8080 -e LATENCY_MS=100 -e FAIL_RATE=0.1 api:local

# In another terminal:
curl http://localhost:8080/api/v1/healthz
curl -X POST -d '{"test": "data"}' http://localhost:8080/api/v1/echo
curl http://localhost:8080/api/v1/work  # ~90% success, 100ms delay

# Load test with hey
hey -z 30s -c 5 http://localhost:8080/api/v1/work
```

---

# Part 2: Kubernetes Core Concepts

**Why:** Kubernetes is the foundation. You need to understand Pods, Deployments, Services before adding complexity.

## 2.1) Kubernetes Architecture

### Concepts to Learn
- [ ] What is a Pod? Why not just "container"?
- [ ] What is a Deployment? Why not create Pods directly?
- [ ] What is a Service? How does service discovery work?
- [ ] What is a Namespace?
- [ ] What are Labels and Selectors?

### Reading (15 min each)
- [ ] [Kubernetes Components](https://kubernetes.io/docs/concepts/overview/components/)
- [ ] [Pods](https://kubernetes.io/docs/concepts/workloads/pods/)
- [ ] [Deployments](https://kubernetes.io/docs/concepts/workloads/controllers/deployment/)
- [ ] [Services](https://kubernetes.io/docs/concepts/services-networking/service/)

---

## 2.2) Deploy Your API with Plain YAML

**Why:** Before using Helm, understand what Helm abstracts away.

### Setup
- [ ] Create folder structure:
  ```
  deploy/
    namespaces/
      apps.yaml
      observability.yaml
    k8s/
      api/
        deployment.yaml
        service.yaml
  ```

### Create Namespace
```yaml
# deploy/namespaces/apps.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: apps
```
- [ ] Apply: `kubectl apply -f deploy/namespaces/`

### Create Deployment
- [ ] Create `deploy/k8s/api/deployment.yaml`:
  - [ ] 2 replicas
  - [ ] Container image: `api:local`
  - [ ] `imagePullPolicy: Never` (use local image)
  - [ ] Environment variables: `LATENCY_MS`, `FAIL_RATE`, `MEMORY_MB`
  - [ ] Resource requests: `cpu: 50m, memory: 64Mi`
  - [ ] Resource limits: `cpu: 200m, memory: 128Mi`
  - [ ] Readiness probe: `GET /api/v1/healthz`
  - [ ] Liveness probe: `GET /api/v1/healthz`

### Create Service
- [ ] Create `deploy/k8s/api/service.yaml`:
  - [ ] Type: ClusterIP
  - [ ] Port: 80 → targetPort: 8080
  - [ ] Selector matching your deployment labels

### Deploy & Verify
```bash
kubectl apply -f deploy/k8s/api/
kubectl -n apps get pods -w          # Watch pods come up
kubectl -n apps get svc              # See service
kubectl -n apps describe pod <name>  # Detailed pod info
kubectl -n apps logs <pod> -f        # Stream logs
```

### Test from Inside Cluster
```bash
# Start a debug pod
kubectl -n apps run curl --image=curlimages/curl -it --rm -- sh

# Inside the pod:
curl http://api/api/v1/healthz
curl http://api/api/v1/work
curl -X POST -d '{"hello":"world"}' http://api/api/v1/echo
```

---

## 2.3) Break Things! (Critical Learning)

**Why:** You'll spend 50% of your job debugging. Learn what failures look like.

### Exercise 1: Break the Readiness Probe
- [ ] Change readiness probe path to `/api/v1/wrong`
- [ ] Apply and observe: `kubectl -n apps get pods -w`
- [ ] What do you see? Pod stays `0/1 Running`
- [ ] Check events: `kubectl -n apps describe pod <name>`
- [ ] Fix it and verify traffic flows again

### Exercise 2: Simulate Errors
- [ ] Set `FAIL_RATE=0.5` in deployment
- [ ] Generate traffic: `kubectl -n apps run curl --image=curlimages/curl -it --rm -- sh -c "while true; do curl -s http://api/api/v1/work; sleep 0.5; done"`
- [ ] Observe ~50% failures
- [ ] Question: How would you detect this in production? (Answer: metrics!)

### Exercise 3: Cause OOMKilled
- [ ] Set memory limit to `16Mi`
- [ ] Apply and watch pod get OOMKilled
- [ ] Check: `kubectl -n apps describe pod <name>` - look for "OOMKilled"
- [ ] Fix by increasing memory limit

### Exercise 4: Break Service Discovery
- [ ] Change service selector to not match deployment labels
- [ ] Observe: service has no endpoints
- [ ] Check: `kubectl -n apps get endpoints api`
- [ ] Debug: `kubectl -n apps describe svc api`
- [ ] Fix selector and verify endpoints appear

---

## 2.4) Ingress Controller

**Why:** In production, external traffic enters through an Ingress. Learn the traffic flow.

### Install NGINX Ingress Controller
```bash
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx
helm repo update
helm install ingress-nginx ingress-nginx/ingress-nginx \
  -n ingress-nginx --create-namespace
```

### Create Ingress Resource
- [ ] Create `deploy/k8s/api/ingress.yaml`:
  - [ ] Host: `api.local`
  - [ ] Path: `/`
  - [ ] Backend: your api service

### Local DNS Setup
- [ ] Add to `/etc/hosts`: `127.0.0.1 api.local`

### Test External Access
```bash
curl http://api.local/api/v1/healthz
curl http://api.local/api/v1/work
curl -X POST -d '{"test":"payload"}' http://api.local/api/v1/echo
```

### Debug Exercise
- [ ] Break ingress backend service name → observe 503 errors
- [ ] Check ingress controller logs: `kubectl -n ingress-nginx logs -l app.kubernetes.io/name=ingress-nginx`
- [ ] Fix and verify

---

# Part 3: Helm Charts

**Why:** Helm is how you'll deploy everything in production. It's templating + package management for Kubernetes.

## 3.1) Helm Concepts

### Concepts to Learn
- [ ] What is a Helm chart?
- [ ] What is `values.yaml`?
- [ ] What is `Chart.yaml`?
- [ ] What are Helm templates and Go templating?
- [ ] What is `helm upgrade --install`?
- [ ] What is a Helm release?

### Reading
- [ ] [Helm Quickstart](https://helm.sh/docs/intro/quickstart/)
- [ ] [Chart Template Guide](https://helm.sh/docs/chart_template_guide/)

---

## 3.2) Create Your First Helm Chart

### Generate Chart Skeleton
```bash
mkdir -p deploy/helm
helm create deploy/helm/api
```

### Understand the Structure
- [ ] Read generated `Chart.yaml`
- [ ] Read generated `values.yaml`
- [ ] Look at `templates/deployment.yaml` - notice `{{ .Values.xxx }}`
- [ ] Look at `templates/_helpers.tpl` - reusable template functions

### Customize for Your API
- [ ] Update `values.yaml`:
  ```yaml
  image:
    repository: api
    tag: local
    pullPolicy: Never
  
  env:
    LATENCY_MS: "0"
    FAIL_RATE: "0.0"
    MEMORY_MB: "0"
  
  resources:
    requests:
      cpu: 50m
      memory: 64Mi
    limits:
      cpu: 200m
      memory: 128Mi
  ```

- [ ] Update `templates/deployment.yaml` to use `{{ .Values.env }}`

### Create Environment Overrides
- [ ] Create `values-dev.yaml`:
  ```yaml
  replicaCount: 1
  env:
    LATENCY_MS: "50"
    FAIL_RATE: "0.0"
    MEMORY_MB: "0"
  ```

- [ ] Create `values-prod.yaml`:
  ```yaml
  replicaCount: 3
  env:
    LATENCY_MS: "0"
    FAIL_RATE: "0.0"
    MEMORY_MB: "0"
  resources:
    requests:
      cpu: 100m
      memory: 128Mi
  ```

### Deploy with Helm
```bash
# Dry-run first - see what would be created
helm upgrade --install api deploy/helm/api -n apps --dry-run

# Actually deploy
helm upgrade --install api deploy/helm/api -n apps -f deploy/helm/api/values-dev.yaml

# Check release
helm list -n apps
helm history api -n apps
```

### Helm Exercises
- [ ] Change an env var and run `helm upgrade` - observe rolling update
- [ ] Rollback: `helm rollback api 1 -n apps`
- [ ] Template debug: `helm template api deploy/helm/api -f deploy/helm/api/values-dev.yaml`

---

# Part 4: Observability - Metrics

**Why:** The job requires VictoriaMetrics and Grafana. Metrics tell you what's happening in your system.

## 4.1) Metrics Concepts

### Concepts to Learn
- [ ] What is Prometheus? What is VictoriaMetrics?
- [ ] What is a metric? (counter, gauge, histogram, summary)
- [ ] What is the Prometheus exposition format?
- [ ] What is scraping?
- [ ] What is PromQL?
- [ ] What are the RED metrics? (Rate, Errors, Duration)

### Reading
- [ ] [Prometheus Metric Types](https://prometheus.io/docs/concepts/metric_types/)
- [ ] [VictoriaMetrics Docs](https://docs.victoriametrics.com/)

---

## 4.2) Install VictoriaMetrics Stack

**Why:** VictoriaMetrics is a Prometheus-compatible, more efficient metrics backend used at SAP.

### Install via Helm
```bash
helm repo add vm https://victoriametrics.github.io/helm-charts/
helm repo update

helm install victoria-metrics vm/victoria-metrics-k8s-stack \
  -n observability --create-namespace \
  -f deploy/helm/observability/victoria-metrics-values.yaml
```

### Create Values File
- [ ] Create `deploy/helm/observability/victoria-metrics-values.yaml`:
  ```yaml
  # Minimal setup for homelab
  victoria-metrics-single:
    enabled: true
  
  vmagent:
    enabled: true
  
  grafana:
    enabled: true
    adminPassword: admin
    service:
      type: NodePort
      nodePort: 30000
  
  # Disable components we don't need yet
  alertmanager:
    enabled: false
  ```

### Access Grafana
- [ ] Get Grafana URL: `kubectl -n observability get svc`
- [ ] Open http://localhost:30000 (or use port-forward)
- [ ] Login: admin / admin
- [ ] Explore pre-built dashboards

### Verify Metrics Collection
- [ ] In Grafana, go to Explore
- [ ] Select VictoriaMetrics datasource
- [ ] Query: `up` - shows all scrape targets
- [ ] Query: `container_cpu_usage_seconds_total` - container metrics

---

## 4.3) Instrument Your API

**Why:** Your app needs to expose metrics for the observability stack to collect.

### Add Prometheus Metrics to Go API
- [ ] Add dependency: `github.com/prometheus/client_golang`
- [ ] Create metrics:
  ```go
  var (
    requestsTotal = prometheus.NewCounterVec(
      prometheus.CounterOpts{
        Name: "http_requests_total",
        Help: "Total HTTP requests",
      },
      []string{"method", "path", "status"},
    )
    requestDuration = prometheus.NewHistogramVec(
      prometheus.HistogramOpts{
        Name: "http_request_duration_seconds",
        Help: "HTTP request duration",
        Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
      },
      []string{"method", "path"},
    )
  )
  ```
- [ ] Add `/metrics` endpoint using `promhttp.Handler()`
- [ ] Instrument handlers to record metrics

### Add ServiceMonitor for Auto-Discovery
- [ ] Create `deploy/helm/api/templates/servicemonitor.yaml`:
  ```yaml
  apiVersion: monitoring.coreos.com/v1
  kind: ServiceMonitor
  metadata:
    name: {{ include "api.fullname" . }}
  spec:
    selector:
      matchLabels:
        {{- include "api.selectorLabels" . | nindent 6 }}
    endpoints:
      - port: http
        path: /metrics
  ```

### Verify Metrics
```bash
# Port-forward to your pod
kubectl -n apps port-forward svc/api 8080:80

# Check metrics endpoint
curl http://localhost:8080/metrics | grep http_requests_total
```

### Build RED Dashboard in Grafana
- [ ] Create new dashboard
- [ ] Panel 1: Request Rate - `rate(http_requests_total[5m])`
- [ ] Panel 2: Error Rate - `rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m])`
- [ ] Panel 3: Latency (p99) - `histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))`

---

## 4.4) Load Test & Observe

**Why:** Generate traffic to see your metrics in action.

### Generate Load
```bash
# Install hey if needed: brew install hey

# Sustained load for 2 minutes
hey -z 2m -c 10 http://api.local/api/v1/work

# Test echo endpoint with POST body
hey -z 1m -c 5 -m POST -d '{"data":"test"}' http://api.local/api/v1/echo
```

### Observe in Grafana
- [ ] Watch request rate increase
- [ ] Watch latency histogram populate
- [ ] Set `FAIL_RATE=0.2` and watch error rate spike
- [ ] Set `LATENCY_MS=500` and watch latency increase
- [ ] Set `MEMORY_MB=50` and watch memory usage grow

---

# Part 5: Observability - Logging

**Why:** The job requires Loki. Logs tell you why something happened.

## 5.1) Logging Concepts

### Concepts to Learn
- [ ] What is structured logging vs unstructured?
- [ ] What is Loki? How is it different from Elasticsearch?
- [ ] What is Promtail?
- [ ] What is LogQL?

### Reading
- [ ] [Grafana Loki Overview](https://grafana.com/docs/loki/latest/)

---

## 5.2) Install Loki Stack

### Install via Helm
```bash
helm repo add grafana https://grafana.github.io/helm-charts
helm repo update

helm install loki grafana/loki-stack \
  -n observability \
  -f deploy/helm/observability/loki-values.yaml
```

### Create Values File
- [ ] Create `deploy/helm/observability/loki-values.yaml`:
  ```yaml
  loki:
    enabled: true
    persistence:
      enabled: false  # For homelab simplicity
  
  promtail:
    enabled: true
  
  grafana:
    enabled: false  # We already have Grafana from VM stack
  ```

### Configure Grafana Datasource
- [ ] In Grafana → Configuration → Data Sources
- [ ] Add Loki datasource
- [ ] URL: `http://loki:3100`

---

## 5.3) Add Structured Logging to Your API

### Update Your Go API
- [ ] Use `slog` (Go 1.21+) for structured JSON logging:
  ```go
  logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
  
  // In your handler:
  logger.Info("request processed",
    "method", c.Request.Method,
    "path", c.Request.URL.Path,
    "status", status,
    "duration_ms", duration.Milliseconds(),
    "trace_id", traceID,
  )
  ```

### Query Logs in Grafana
- [ ] Go to Explore → Select Loki
- [ ] Query: `{namespace="apps", app="api"}`
- [ ] Filter errors: `{namespace="apps"} |= "error"`
- [ ] Parse JSON: `{namespace="apps"} | json | status >= 500`

### Log Exploration Exercises
- [ ] Find all requests that took > 100ms
- [ ] Find all 500 errors in the last hour
- [ ] Correlate a spike in error metrics with log messages

---

# Part 6: Observability - Tracing

**Why:** The job mentions Jaeger. Traces show you the path of a request through your system.

## 6.1) Tracing Concepts

### Concepts to Learn
- [ ] What is distributed tracing?
- [ ] What is a trace? What is a span?
- [ ] What is OpenTelemetry?
- [ ] What is Jaeger? What is Tempo?
- [ ] What is trace context propagation?

### Reading
- [ ] [OpenTelemetry Concepts](https://opentelemetry.io/docs/concepts/)
- [ ] [Jaeger Introduction](https://www.jaegertracing.io/docs/latest/)

---

## 6.2) Install Jaeger

### Install via Helm
```bash
helm repo add jaegertracing https://jaegertracing.github.io/helm-charts
helm repo update

helm install jaeger jaegertracing/jaeger \
  -n observability \
  -f deploy/helm/observability/jaeger-values.yaml
```

### Create Values File
- [ ] Create `deploy/helm/observability/jaeger-values.yaml`:
  ```yaml
  provisionDataStore:
    cassandra: false
  
  allInOne:
    enabled: true
  
  storage:
    type: memory  # For homelab; use Cassandra/Elasticsearch in prod
  
  agent:
    enabled: true
  
  collector:
    enabled: false  # allInOne handles this
  
  query:
    enabled: false  # allInOne handles this
  ```

### Access Jaeger UI
```bash
kubectl -n observability port-forward svc/jaeger-query 16686:16686
# Open http://localhost:16686
```

---

## 6.3) Instrument Your API with OpenTelemetry

### Add OpenTelemetry to Go API
- [ ] Add dependencies:
  ```
  go.opentelemetry.io/otel
  go.opentelemetry.io/otel/exporters/jaeger
  go.opentelemetry.io/otel/sdk/trace
  go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin
  ```

- [ ] Initialize tracer provider pointing to Jaeger
- [ ] Add middleware: `r.Use(otelgin.Middleware("api"))`
- [ ] Add trace_id to your structured logs (correlate traces with logs!)

### View Traces in Jaeger
- [ ] Generate some traffic
- [ ] Open Jaeger UI
- [ ] Select service "api"
- [ ] View traces - see request flow and timing

### Trace Exercises
- [ ] Find the slowest requests
- [ ] Correlate a trace_id from logs to Jaeger
- [ ] Add a custom span for the "prediction" logic

---

# Part 7: Crossplane (Infrastructure as Data)

**Why:** The job mentions Crossplane as beneficial. It's GitOps for infrastructure.

## 7.1) Crossplane Concepts

### Concepts to Learn
- [ ] What is Crossplane?
- [ ] What is a Provider?
- [ ] What is a Managed Resource?
- [ ] What is a Composition?
- [ ] What is an XRD (CompositeResourceDefinition)?
- [ ] How is Crossplane different from Terraform?

### Reading
- [ ] [Crossplane Concepts](https://docs.crossplane.io/latest/concepts/)
- [ ] [Crossplane vs Terraform](https://docs.crossplane.io/latest/concepts/crossplane-vs-terraform/)

---

## 7.2) Install Crossplane

```bash
helm repo add crossplane-stable https://charts.crossplane.io/stable
helm repo update

helm install crossplane crossplane-stable/crossplane \
  -n crossplane-system --create-namespace
```

### Verify Installation
```bash
kubectl get pods -n crossplane-system
kubectl api-resources | grep crossplane
```

---

## 7.3) Create a Simple Composition

**Why:** Compositions let you define abstractions - e.g., "give me a database" without knowing the cloud details.

### Install Kubernetes Provider (for local practice)
```yaml
# deploy/crossplane/provider-kubernetes.yaml
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: provider-kubernetes
spec:
  package: xpkg.upbound.io/crossplane-contrib/provider-kubernetes:v0.9.0
```

### Create a Simple XRD
- [ ] Create an XRD that defines a "MyApp" abstraction
- [ ] Create a Composition that creates Deployment + Service + ConfigMap
- [ ] Create a Claim that requests a "MyApp"

### Exercise
- [ ] Deploy an app using only a Crossplane Claim
- [ ] Modify the Composition and watch resources update
- [ ] Delete the Claim and watch resources get cleaned up

---

# Part 8: Putting It All Together

## 8.1) Full Stack Exercise

Create a complete deployment that demonstrates:
- [ ] Helm chart with environment-specific values
- [ ] Prometheus metrics exposed and scraped
- [ ] Structured JSON logs collected by Loki
- [ ] Distributed traces sent to Jaeger
- [ ] Grafana dashboard showing RED metrics
- [ ] Ability to correlate metrics → logs → traces

## 8.2) Chaos Engineering Exercise

- [ ] Use your `FAIL_RATE`, `LATENCY_MS`, and `MEMORY_MB` to simulate incidents
- [ ] Practice the debugging workflow:
  1. Alert fires (metric threshold breached)
  2. Check dashboard (see the spike)
  3. Query logs (find error messages)
  4. Check traces (see where slowness is)
  5. Fix and verify

## 8.3) Documentation Exercise

- [ ] Write runbook for "API high error rate"
- [ ] Document your observability stack architecture
- [ ] Create architecture diagram

---

# Project Structure

```
k8s-homelab/
├── README.md
├── apps/
│   └── api/
│       ├── api.go
│       ├── api_test.go
│       ├── Dockerfile
│       └── go.mod
├── deploy/
│   ├── namespaces/
│   │   ├── apps.yaml
│   │   └── observability.yaml
│   ├── k8s/
│   │   └── api/
│   │       ├── deployment.yaml
│   │       ├── service.yaml
│   │       └── ingress.yaml
│   ├── helm/
│   │   ├── api/
│   │   │   ├── Chart.yaml
│   │   │   ├── values.yaml
│   │   │   ├── values-dev.yaml
│   │   │   ├── values-prod.yaml
│   │   │   └── templates/
│   │   └── observability/
│   │       ├── victoria-metrics-values.yaml
│   │       ├── loki-values.yaml
│   │       └── jaeger-values.yaml
│   └── crossplane/
│       └── provider-kubernetes.yaml
└── scripts/
    └── load-test.sh
```

---

# Quick Reference Commands

```bash
# Kubernetes
kubectl get pods -n apps -w              # Watch pods
kubectl describe pod <name> -n apps      # Debug pod
kubectl logs <pod> -n apps -f            # Stream logs
kubectl exec -it <pod> -n apps -- sh     # Shell into pod
kubectl port-forward svc/api 8080:80 -n apps  # Local access

# Helm
helm list -n apps                        # List releases
helm history <release> -n apps           # Release history
helm rollback <release> <revision> -n apps  # Rollback
helm template <chart-path>               # Debug templates

# Debugging
kubectl get events -n apps --sort-by='.lastTimestamp'
kubectl top pods -n apps                 # Resource usage
kubectl get endpoints -n apps            # Service endpoints

# Observability
kubectl -n observability port-forward svc/grafana 3000:80
kubectl -n observability port-forward svc/jaeger-query 16686:16686
```

---

# Resources

### Kubernetes
- [Kubernetes Documentation](https://kubernetes.io/docs/home/)
- [kubectl Cheat Sheet](https://kubernetes.io/docs/reference/kubectl/cheatsheet/)

### Helm
- [Helm Documentation](https://helm.sh/docs/)
- [Artifact Hub](https://artifacthub.io/) - Find Helm charts

### Observability
- [VictoriaMetrics Docs](https://docs.victoriametrics.com/)
- [Grafana Loki Docs](https://grafana.com/docs/loki/latest/)
- [Jaeger Docs](https://www.jaegertracing.io/docs/)
- [OpenTelemetry Docs](https://opentelemetry.io/docs/)

### Crossplane
- [Crossplane Docs](https://docs.crossplane.io/)
- [Upbound Marketplace](https://marketplace.upbound.io/) - Find providers
