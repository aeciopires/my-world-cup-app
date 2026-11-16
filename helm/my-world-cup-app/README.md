# my-world-cup-app

![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square)  ![Version: 2.0.0](https://img.shields.io/badge/Version-2.0.0-informational?style=flat-square)

A Helm chart for My World Cup App, a Go web application displaying FIFA World Cup 2026 standings, knockout results, and statistics.

# Prerequisites

- **Helm 4** (Helm 3 is no longer supported; rendering with it fails with an explicit error).
- kubectl and access to a Kubernetes cluster.
- helm-docs (only to regenerate this README).

All of them are pinned in the repository's `mise.toml`; run `make install-deps` from the repository root to install them.

# Installing the Chart

- Access a Kubernetes cluster.

- Change the values according to the need of the environment in ``my-world-cup-app/values.yaml`` file. The [Parameters](#parameters) section lists the parameters that can be configured during installation.

- Test the installation with command:

```bash
helm upgrade --install my-world-cup-app -f my-world-cup-app/values.yaml my-world-cup-app/ -n my-world-cup-app --create-namespace --dry-run=client
```

- To install/upgrade the chart with the release name `my-world-cup-app`:

```bash
helm upgrade --install my-world-cup-app -f my-world-cup-app/values.yaml my-world-cup-app/ -n my-world-cup-app --create-namespace
```

Create a port-forward with the follow command:

```bash
kubectl port-forward svc/my-world-cup-app 8080:8080 -n my-world-cup-app
```

Open the browser and access the URL: http://localhost:8080

## Uninstalling the Chart

To uninstall/delete the `my-world-cup-app` deployment:

```bash
helm uninstall my-world-cup-app -n my-world-cup-app
```

The command removes all the Kubernetes components associated with the chart and deletes the release.

## Parameters

The following tables lists the configurable parameters of the chart and their default values.

Change the values according to the need of the environment in ``my-world-cup-app/values.yaml`` file.

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Affinity rules for pod scheduling |
| autoscaling.enabled | bool | `false` | Enable the HorizontalPodAutoscaler (disables replicaCount) |
| autoscaling.maxReplicas | int | `3` | Maximum number of replicas |
| autoscaling.minReplicas | int | `1` | Minimum number of replicas |
| autoscaling.targetCPUUtilizationPercentage | int | `80` | Target average CPU utilization percentage |
| containerPort | int | `8080` | Port the application listens on inside the container |
| env | object | `{"LOG_FORMAT":"json","PORT":"8080","SHUTDOWN_DRAIN_DELAY":"5s"}` | Environment variables passed through to the container (see the app's README "Configuration" table). PORT must match service.port and containerPort if overridden. SHUTDOWN_DRAIN_DELAY keeps the pod serving while /readyz reports 503 after SIGTERM, so it leaves the Service's endpoints before connections close; keep it below terminationGracePeriodSeconds minus ~10s. Uncomment the OTEL_* entries to export traces to an OpenTelemetry collector (e.g. Grafana Alloy or the OpenTelemetry Collector). |
| fullnameOverride | string | `""` | Override the full generated resource name |
| image.pullPolicy | string | `"IfNotPresent"` | Container image pull policy |
| image.repository | string | `"my-world-cup-app"` | Container image repository |
| image.tag | string | `""` | Container image tag (defaults to .Chart.AppVersion when empty) |
| imagePullSecrets | list | `[]` | References to secrets used to pull the image from a private registry |
| ingress.annotations | object | `{}` | Annotations to add to the Ingress resource |
| ingress.className | string | `""` | IngressClass name (leave empty to use the cluster default) |
| ingress.enabled | bool | `false` | Enable the Ingress resource |
| ingress.hosts | list | `[{"host":"worldcup.example.com","paths":[{"path":"/","pathType":"Prefix"}]}]` | List of hosts and paths to route to the Service |
| ingress.tls | list | `[]` | TLS configuration for the Ingress resource |
| livenessProbe.failureThreshold | int | `3` | Consecutive failures before the container is restarted |
| livenessProbe.httpGet.path | string | `"/healthz"` | Liveness probe HTTP path |
| livenessProbe.httpGet.port | string | `"http"` | Liveness probe port (named container port) |
| livenessProbe.periodSeconds | int | `15` | Seconds between liveness probes |
| livenessProbe.timeoutSeconds | int | `2` | Seconds before a liveness probe times out |
| nameOverride | string | `""` | Override the chart name used in generated resource names |
| nodeSelector | object | `{}` | Node selector for pod scheduling |
| podAnnotations."prometheus.io/path" | string | `"/metrics"` | Path a cluster Prometheus should scrape for metrics |
| podAnnotations."prometheus.io/port" | string | `"8080"` | Port a cluster Prometheus should scrape for metrics |
| podAnnotations."prometheus.io/scrape" | string | `"true"` | Enables Prometheus scrape discovery for this pod |
| podSecurityContext.runAsGroup | int | `65532` | GID to run as |
| podSecurityContext.runAsNonRoot | bool | `true` | Require the pod to run as a non-root user |
| podSecurityContext.runAsUser | int | `65532` | UID to run as (matches the distroless "nonroot" user baked into the image) |
| podSecurityContext.seccompProfile.type | string | `"RuntimeDefault"` | Seccomp profile applied to the pod |
| readinessProbe.failureThreshold | int | `1` | Consecutive failures before the pod is removed from Service endpoints |
| readinessProbe.httpGet.path | string | `"/readyz"` | Readiness probe HTTP path (503 while draining on shutdown) |
| readinessProbe.httpGet.port | string | `"http"` | Readiness probe port (named container port) |
| readinessProbe.periodSeconds | int | `5` | Seconds between readiness probes (keep it below env.SHUTDOWN_DRAIN_DELAY) |
| readinessProbe.timeoutSeconds | int | `2` | Seconds before a readiness probe times out |
| replicaCount | int | `1` | Number of pod replicas (ignored when autoscaling.enabled is true) |
| resources.limits.cpu | string | `"250m"` | CPU limit |
| resources.limits.memory | string | `"128Mi"` | Memory limit |
| resources.requests.cpu | string | `"50m"` | CPU request |
| resources.requests.memory | string | `"32Mi"` | Memory request |
| securityContext.allowPrivilegeEscalation | bool | `false` | Disallow privilege escalation for the container |
| securityContext.capabilities.drop | list | `["ALL"]` | Linux capabilities to drop from the container |
| securityContext.readOnlyRootFilesystem | bool | `true` | Mount the container's root filesystem as read-only |
| service.port | int | `8080` | Service port (also used as the Prometheus scrape port annotation) |
| service.type | string | `"ClusterIP"` | Kubernetes Service type |
| serviceAccount.annotations | object | `{}` | Annotations to add to the ServiceAccount |
| serviceAccount.automountToken | bool | `false` | Mount the ServiceAccount API token into the pod. The app never talks to the Kubernetes API, so this is off by default. |
| serviceAccount.create | bool | `true` | Whether to create a ServiceAccount for the pod |
| serviceAccount.name | string | `""` | Name of the ServiceAccount to use (generated if empty and create=true) |
| startupProbe.failureThreshold | int | `15` | Failed startup probes tolerated (periodSeconds x failureThreshold = max startup time) |
| startupProbe.httpGet.path | string | `"/healthz"` | Startup probe HTTP path |
| startupProbe.httpGet.port | string | `"http"` | Startup probe port (named container port) |
| startupProbe.periodSeconds | int | `2` | Seconds between startup probes |
| startupProbe.timeoutSeconds | int | `2` | Seconds before a startup probe times out |
| terminationGracePeriodSeconds | int | `30` | Seconds Kubernetes waits after SIGTERM before killing the pod. Must cover env.SHUTDOWN_DRAIN_DELAY plus the app's 10s graceful-shutdown window. |
| tolerations | list | `[]` | Tolerations for pod scheduling |