# NodeQuotaSync Plugin for HNS

The `NodeQuotaSync` plugin is an extension for [HNS](https://github.com/dana-team/hns).

It enables syncing the `root subnamespace` quota object and `secondary subnamespaces` quota objects with the node's allocatable resources in the cluster. It provides support for resources multiplier for over-commit and a reserved resources mechanism, making it easier to remove nodes from the cluster temporarily without affecting the allocatable `subnamespaces` resources.

## Features

- Auto-syncs the quota objects of the `root subnamespace` and `secondary subnamespaces` with the matching nodes' allocatable resources.
- Configurable resources multiplier for over-commit.
- Reserved resources mechanism for removing nodes in a safe way.
- Select what type of resource to control
- Control through a Config CRD

## Install with Helm

Helm chart docs are available on `charts/hns-nqs-plugin` directory.

```bash
$ helm upgrade hns-nqs-plugin --install --namespace nodequotasync-system --create-namespace oci://ghcr.io/dana-team/helm-charts/hns-nqs-plugin --version <release>
```

## Configuration

The `NodeQuotaSync` plugin is configured using the `NodeQuotaConfig` CRD:

```yaml
apiVersion: dana.hns.io/v1alpha1
kind: NodeQuotaConfig
metadata:
  name: example-nodequotaconfig
spec:
  reservedHoursToLive: 24
  controlledResources: ["cpu","ephermal-storage","memory","pods","nvidia.com/gpu"]
  subnamespacesRoots:
    - rootNamespace: cluster-root
      secondaryRoots:
        - labelSelector:
            app: gpu
          name: gpu
          multipliers:  
            cpu: "2"
            memory: "2"
        - labelSelector:
            app: cpu-workloads
          name: cpu-workloads
          multipliers:
            memory: "4"
```

- `subnamespaceRoots` - defines the cluster's hierarchy;
- `rootNamespace` - represents the name of the `root` namespace;
- `secondaryRoots` - the direct children of the `root` namespaces with their corresponding node's `labelSelector` and multipliers.

## ReservedResources

The `ReservedResources` mechanism is a way to ensure there is no resources shortage when dealing with `Nodes` maintenance.

It works by giving the cluster's admins time to restore the node into the cluster without recalculating the cluster's resources and only removing the node's resources in a controlled way after a pre-set number of hours. This can be configured in the Config CR using the `ReservedHoursToLive` field.

When we remove one of the nodes from the cluster a `ReservedResources` field will be added to the CRD status:

```yaml
  reservedResources:
    - Timestamp: '2023-07-09T07:04:27Z'
      nodeGroup: cpu-workloads
      resources:
        cpu: 3500m
        memory: '61847027712'
        pods: '250'
```

The cluster's resources will not be updated until the number of hours in the `reservedHoursToLive` will pass from (starting from `Timestamp`); afterwards, the node's resources will be removed.
