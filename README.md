# NodeQuotaSync Plugin for HNS

The NodeQuotaSync plugin enables syncing the root quota and secondary subNamespace with the nodes' resources in the cluster. It provides support for resources multiplier for auto commit and reserved resources mechanism, making it easier to troubleshoot nodes without affecting the subnamespace wallets.

## Features

- Syncs root quota and secondary subNamespace with nodes' resources
- Resources multiplier for auto commit
- Reserved resources mechanism for troubleshooting nodes

## Installation

To install the NodeQuotaSync plugin, follow these steps:

1. Clone the repository or download the plugin code.
2. Build the plugin using the provided build script.
3. Copy the built binary to the desired location.
4. Configure the plugin settings according to your requirements.

## Usage

1. Start the HNS service with the NodeQuotaSync plugin enabled.
2. Configure the plugin settings in the HNS configuration file.
3. Monitor the syncing process and resource allocation using the HNS CLI or dashboard.

## Configuration

The NodeQuotaSync plugin can be configured by modifying the HNS configuration file. The configuration options for the plugin are as follows:


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

