# NodeQuotaSync Plugin for HNS

The NodeQuotaSync plugin enables syncing the root subnamespace and secondary subnamespaces with the node's allocatable resources in the cluster. It provides support for resources multiplier for over-commit and reserved resources mechanism, making it easier to remove nodes from the cluster temporarily without affecting the subnamespaces wallets.

## Features

- Auto-sync root subnamespace and secondary subnamespaces with the matching nodes' allocatable resources.
- Configurable resources multiplier for over-commit.
- Reserved resources mechanism for removing nodes in a safe way.
- Select what type of resource to control
- Config CRD

## Installation

To install the NodeQuotaSync plugin, follow these steps:

1. Clone the repository or download the plugin code.
2. Build the plugin using the provided build script.
3. Deploy

## Configuration 

The NodeQuotaSync plugin can be configured by modifying the HNS configuration file. The configuration options for the plugin are as follows:

```
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

subnamespaceRoots defines the cluster's hierarchy, the `name` field represents the name of the `root` namespace and the secondaryRoots are the direct children of the `root` namespaces with their corresponding node's labelSelector and multipliers.

## About the ReservedResources 

The ReservedResources mechanism is a way to ensure we don't encounter resources shortage and `HNS` breakdowns when dealing with 
`Nodes` maintenance. It works by giving the cluster's admins time to return the node to the cluster without recalculating the cluster's resources and only removes the node's resources in a controlled way after a number of hours that can be configured in the Config CR with the `ReservedHoursToLive` field.

When we remove one of the nodes from the cluster a `ReservedResources` will be added to the CRD status:
```
  reservedResources:
    - Timestamp: '2023-07-09T07:04:27Z'
      nodeGroup: cpu-workloads
      resources:
        cpu: 3500m
        memory: '61847027712'
        pods: '250'
```

The cluster's resources won't update until we either return the node's resources or the number of hours in the `ReservedHoursTolive` will pass from the `Timestamp` and then the resources will be removed.

## Usage

1. Deploy the controller with the image `danateam/nodequotasync:tagname`
2. Create the `NodeQuotaConfig` CR
 
## License

Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

