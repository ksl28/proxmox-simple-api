# proxmox-simple-api
Simple API written in Go, that allows you to collect info from multiple Proxmox instances, without the need for instances to be in the same datacenter or cluster.


## Features
- A quick overview of all nodes, that are configured in the API.
- A detailed overview of all the nodes within a "parent", that are configured in the API.
- A quick overview of all the VMs that are running on the nodes configured.
- A detailed insight into a VM, based on the "parent" cluster and the VM ID in Proxmox.
- A quick overview of all the LXC containers that are running on the nodes configured.
- A detailed overview of the storage classes for all the nodes, that are configured in the API.
- A detailed overview of the disks for all the nodes, that are configured in the API. 

## Requirements
- Go 1.23+
- An API key with the correct permissions per parent
- The server or pod running the program, must be able to reach the parent

## Security / Disclaimer

> ⚠️ **No built-in authentication**  
> This API does **not** include any authentication or authorization mechanisms.  
> You must secure access (e.g. via network policies, API gateway, or reverse proxy with auth) before exposing it in production.



## Key Concepts

- **Parent**  
  A Proxmox node in the cluster that serves as the entry point for API calls.

- **API Token**  
  The `Token` should be formatted as `PVEAPIToken=<user>@<realm>!<tokenid>=<secret>`

## Endpoints

- **`GET /api/v1/infrastructure/nodes/summary`**  
  Returns a summary of CPU, memory, and disk usage for all configured parents and their nodes.

- **`GET /api/v1/infrastructure/nodes/detailed/:name`**  
  Returns detailed information (CPU, memory, network, kernel, etc.) for the node identified by `:name`.

- **`GET /api/v1/infrastructure/nodes/detailed/:name/storage`**  
  Returns a per-volume storage overview for the node `:name`, including total, used, and available space.

- **`GET /api/v1/infrastructure/nodes/detailed/:name/disks`**  
  Returns a per-disk health and usage overview for the node `:name`, including vendor, serial, and capacity.
  > **Note:** The API used for getting the disks can take several seconds to finish - with several nodes, this will likely cause this API to become slow.

- **`GET /api/v1/virtualization/vm/summary`**  
  Returns a brief list of all QEMU virtual machines across the cluster(s), with parent node, ID, and name.

- **`GET /api/v1/virtualization/vm/detailed/:parent/:id`**  
  Returns full details for the QEMU VM with ID `:id` on node `:parent`, including status, hostname, OS info, and networking.

- **`GET /api/v1/virtualization/lxc/summary`**  
  Returns a summary list of all LXC containers across the configured clusters, with parent node, ID, and status.

> **Note:** The API listens on `0.0.0.0:${APIPORT}` (default: 8080) as configured via the `apiport` environment variable.

### Example:

To run the program from Powershell:
```Powershell
$objects = @(
    @{
        Type  = "cluster"
        Name  = "parent01.domain.tld"
        Token = "<API Token>"
        Port  = 8006
    },
    @{
        Type  = "node"
        Name  = "parent02.domain.tld"
        Token = "<API Token>"
        Port  = 8006
    }
)
$json = $objects | ConvertTo-Json -Compress -Depth 10
$env:OBJECTS_JSON = $json
```

From Kubernetes (simple example)
```yaml 
apiVersion: apps/v1
kind: Deployment
metadata:
  name: pve-api
spec:
  replicas: 1
  selector:
    matchLabels:
      app: pve-api
  template:
    metadata:
      labels:
        app: pve-api
    spec:
      containers:
        - name: pve-api
          image: your-registry/pve-api:latest  
          ports:
            - containerPort: 8080
          env:
            - name: OBJECTS_JSON
              value: |
                [
                  {"Type":"cluster","Name":"parent01.domain.tld","Token":"<API Token>","Port":8006},
                  {"Type":"node","Name":"parent02.domain.tld","Token":"<API Token>","Port":8006}
                ]

```