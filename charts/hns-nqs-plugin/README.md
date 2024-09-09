# hns-nqs-plugin

![Version: 0.0.0](https://img.shields.io/badge/Version-0.0.0-informational?style=flat-square) ![Type: application](https://img.shields.io/badge/Type-application-informational?style=flat-square) ![AppVersion: latest](https://img.shields.io/badge/AppVersion-latest-informational?style=flat-square)

A Helm chart for the hns-nqs-plugin.

## Values

| Key | Type | Default | Description |
|-----|------|---------|-------------|
| affinity | object | `{}` | Node affinity rules for scheduling pods. Allows you to specify advanced node selection constraints. |
| fullnameOverride | string | `""` |  |
| image.kubeRbacProxy.pullPolicy | string | `"IfNotPresent"` | The pull policy for the image. |
| image.kubeRbacProxy.repository | string | `"gcr.io/kubebuilder/kube-rbac-proxy"` | The repository of the kube-rbac-proxy container image. |
| image.kubeRbacProxy.tag | string | `"v0.16.0"` | The tag of the kube-rbac-proxy container image. |
| image.manager.pullPolicy | string | `"IfNotPresent"` | The pull policy for the image. |
| image.manager.repository | string | `"ghcr.io/dana-team/hns-nqs-plugin"` | The repository of the manager container image. |
| image.manager.tag | string | `""` | The tag of the manager container image. |
| kubeRbacProxy | object | `{"args":["--secure-listen-address=0.0.0.0:8443","--upstream=http://127.0.0.1:8080/","--logtostderr=true","--v=0"],"ports":{"https":{"containerPort":8443,"name":"https","protocol":"TCP"}},"resources":{"limits":{"cpu":"500m","memory":"128Mi"},"requests":{"cpu":"5m","memory":"64Mi"}},"securityContext":{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]}}}` | Configuration for the kube-rbac-proxy container. |
| kubeRbacProxy.args | list | `["--secure-listen-address=0.0.0.0:8443","--upstream=http://127.0.0.1:8080/","--logtostderr=true","--v=0"]` | Command-line arguments passed to the kube-rbac-proxy container. |
| kubeRbacProxy.ports.https.containerPort | int | `8443` | The port for the HTTPS endpoint. |
| kubeRbacProxy.ports.https.name | string | `"https"` | The name of the HTTPS port. |
| kubeRbacProxy.ports.https.protocol | string | `"TCP"` | The protocol used by the HTTPS endpoint. |
| kubeRbacProxy.resources | object | `{"limits":{"cpu":"500m","memory":"128Mi"},"requests":{"cpu":"5m","memory":"64Mi"}}` | Resource requests and limits for the kube-rbac-proxy container. |
| kubeRbacProxy.securityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]}}` | Security settings for the kube-rbac-proxy container. |
| livenessProbe | object | `{"initialDelaySeconds":15,"periodSeconds":20}` | Configuration for the liveness probe. |
| livenessProbe.initialDelaySeconds | int | `15` | The initial delay before the liveness probe is initiated. |
| livenessProbe.periodSeconds | int | `20` | The frequency (in seconds) with which the probe will be performed. |
| manager | object | `{"args":["--leader-elect","--health-probe-bind-address=:8081","--metrics-bind-address=127.0.0.1:8080"],"command":["/manager"],"ports":{"health":{"containerPort":8081,"name":"health","protocol":"TCP"}},"resources":{"limits":{"cpu":"500m","memory":"128Mi"},"requests":{"cpu":"10m","memory":"64Mi"}},"securityContext":{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]}}}` | Configuration for the manager container. |
| manager.args | list | `["--leader-elect","--health-probe-bind-address=:8081","--metrics-bind-address=127.0.0.1:8080"]` | Command-line arguments passed to the manager container. |
| manager.command | list | `["/manager"]` | Command-line commands passed to the manager container. |
| manager.ports.health.containerPort | int | `8081` | The port for the health check endpoint. |
| manager.ports.health.name | string | `"health"` | The name of the health check port. |
| manager.ports.health.protocol | string | `"TCP"` | The protocol used by the health check endpoint. |
| manager.resources | object | `{"limits":{"cpu":"500m","memory":"128Mi"},"requests":{"cpu":"10m","memory":"64Mi"}}` | Resource requests and limits for the manager container. |
| manager.securityContext | object | `{"allowPrivilegeEscalation":false,"capabilities":{"drop":["ALL"]}}` | Security settings for the manager container. |
| nameOverride | string | `""` |  |
| nodeQuotaConfig.controlledResources | list | `["cpu","memory","pods"]` | Defines which node resources are controlled. |
| nodeQuotaConfig.enabled | bool | `false` |  |
| nodeQuotaConfig.name | string | `"cluster-nodequotaconfig"` | The name of the NodeQuotaConfig resource. |
| nodeQuotaConfig.reservedHoursToLive | int | `48` | Defines how many hours the ReservedResources can live until they are removed from the cluster resources. |
| nodeQuotaConfig.subnamespacesRoots | list | `[{"rootNamespace":"cluster-root","secondaryRoots":[{"labelSelector":{"app":"gpu"},"multipliers":{"cpu":"1","memory":"1"},"name":"gpu"},{"labelSelector":{"app":"cpu-workloads"},"multipliers":{"memory":"1"},"name":"cpu-workloads"}]}]` | The cluster's hierarchy (root namespace and secondary roots). |
| nodeQuotaConfig.subnamespacesRoots[0] | object | `{"rootNamespace":"cluster-root","secondaryRoots":[{"labelSelector":{"app":"gpu"},"multipliers":{"cpu":"1","memory":"1"},"name":"gpu"},{"labelSelector":{"app":"cpu-workloads"},"multipliers":{"memory":"1"},"name":"cpu-workloads"}]}` | The name of the root namespace. |
| nodeQuotaConfig.subnamespacesRoots[0].secondaryRoots | list | `[{"labelSelector":{"app":"gpu"},"multipliers":{"cpu":"1","memory":"1"},"name":"gpu"},{"labelSelector":{"app":"cpu-workloads"},"multipliers":{"memory":"1"},"name":"cpu-workloads"}]` | The hierarchy of the secondary roots. |
| nodeQuotaConfig.subnamespacesRoots[0].secondaryRoots[0].multipliers | object | `{"cpu":"1","memory":"1"}` | The overcommit multipliers. |
| nodeSelector | object | `{}` | Node selector for scheduling pods. Allows you to specify node labels for pod assignment. |
| readinessProbe | object | `{"initialDelaySeconds":5,"periodSeconds":10}` | Configuration for the readiness probe. |
| readinessProbe.initialDelaySeconds | int | `5` | The initial delay before the readiness probe is initiated. |
| readinessProbe.periodSeconds | int | `10` | The frequency (in seconds) with which the probe will be performed. |
| replicaCount | int | `1` | The number of replicas for the deployment. |
| securityContext | object | `{}` | Pod-level security context for the entire pod. |
| service | object | `{"httpsPort":8443,"protocol":"TCP","targetPort":"https","type":"ClusterIP"}` | Configuration for the metrics service. |
| service.httpsPort | int | `8443` | The port for the HTTPS endpoint. |
| service.protocol | string | `"TCP"` | The protocol used by the HTTPS endpoint. |
| service.targetPort | string | `"https"` | The name of the target port. |
| service.type | string | `"ClusterIP"` | The type of the service. |
| tolerations | list | `[]` | Node tolerations for scheduling pods. Allows the pods to be scheduled on nodes with matching taints. |

