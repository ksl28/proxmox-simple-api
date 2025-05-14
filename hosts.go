package main

import (
	"fmt"
	"log"
	"math"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func testHostPort(host string, port int) (bool, error) {
	hostPort := fmt.Sprintf("%s:%d", host, port)
	_, err := net.DialTimeout("tcp", hostPort, time.Duration(600)*time.Millisecond)
	if err != nil {
		return false, err
	} else {
		return true, nil
	}
}

func quickHostOverview(c *gin.Context) {
	var (
		results []nodeSummaryWrapper
		errors  []ApiError
	)

	parentObjects, err := convertJSON()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read JSON data"})
		return
	}

	for _, host := range parentObjects {
		portOpen, err := testHostPort(host.Parent, host.Port)
		if err != nil {
			errors = append(errors, ApiError{
				Parent:  host.Parent,
				Action:  "testHostPort",
				Message: err.Error(),
			})
			log.Printf("Failed to check if the port %d for %s is open - %v", host.Port, host.Parent, err)
			continue
		}
		if portOpen {
			parentNodes, err := getParentNodes(host.Parent, host.Port, host.Token)
			if err != nil {
				errors = append(errors, ApiError{
					Parent:  host.Parent,
					Action:  "getParentNodes",
					Message: err.Error(),
				})
				log.Printf("Failed to obtain the nodes under the parent %s - %v", host.Parent, err)
				continue
			}
			for _, node := range parentNodes.Data {
				var summary nodeSummaryWrapper
				//If the first object in the slice is empty / offline, then the struct will be limited to only show the fields that have values.
				if node.NodeStatus != "online" {
					errors = append(errors, ApiError{
						Parent:  host.Parent,
						Node:    node.Node,
						Action:  "onlineStatus",
						Message: fmt.Sprintf("The node %s is offline", node.Node),
					})
					summary.Parent = host.Parent
					summary.Node = node.Node
					summary.NodeStatus = node.NodeStatus
					summary.MaxCPU = 0
					summary.MaxMemGb = 0
					summary.MemGb = 0
					summary.UptimeHours = 0
					summary.Cpu = 0
					summary.MaxRootDiskGb = 0
					summary.RootDiskGb = 0
					results = append(results, summary)
					continue
				} else {
					summary.Parent = host.Parent
					summary.Node = node.Node
					summary.NodeStatus = node.NodeStatus
					summary.MaxCPU = node.MaxCPU
					summary.MaxMemGb = int(node.MaxMem) / (1024 * 1024 * 1024)
					summary.MemGb = int(node.Mem) / (1024 * 1024 * 1024)
					summary.UptimeHours = int(node.UptimeHours) / (60 * 60)
					summary.Cpu = math.Round(node.Cpu * 100)
					summary.MaxRootDiskGb = node.Maxdisk / (1000 * 1000 * 1000)
					summary.RootDiskGb = node.Disk / (1000 * 1000 * 1000)
					results = append(results, summary)
				}

			}
		} else {
			errors = append(errors, ApiError{
				Parent:  host.Parent,
				Action:  "testHostPort",
				Message: fmt.Sprintf("The check to see if the port was open executed with success, but the parent appears to be offline or not listening on %d", host.Port),
			})
			log.Printf("The server %s is not listening on %d", host.Parent, host.Port)
			continue
		}
	}
	c.JSON(http.StatusOK, NodeSummaryResponse{
		Data:   results,
		Errors: errors,
	})
}

func getParentNodes(parent string, port int, apiToken string) (PVENodesObject, error) {
	customUrl := fmt.Sprintf("https://%s:%d/api2/json/nodes", parent, port)
	req, err := http.NewRequest(http.MethodGet, customUrl, nil)
	if err != nil {
		log.Printf("Error creating the HTTP request for %s - %v", customUrl, err)
		return PVENodesObject{}, err
	}
	var parentNodes PVENodesObject
	if err := sendRequest(req, &parentNodes, apiToken); err != nil {
		log.Printf("Failed to process the request for %s - error %v", customUrl, err)
		return PVENodesObject{}, err
	}

	return parentNodes, nil
}

func nodeGuestsOverview(guestType string, parent string, port int, node string, apiToken string) (NodeGuestOverview, error) {

	var customUrl string
	if guestType == "lxc" {
		customUrl = fmt.Sprintf("https://%s:%d/api2/json/nodes/%s/lxc", parent, port, node)
	} else {
		customUrl = fmt.Sprintf("https://%s:%d/api2/json/nodes/%s/qemu", parent, port, node)
	}

	req, err := http.NewRequest(http.MethodGet, customUrl, nil)
	if err != nil {
		log.Printf("Failed to create the HTTP request for %s - %v", parent, err)
		return NodeGuestOverview{}, err
	}

	var qemuList NodeGuestOverview
	if err := sendRequest(req, &qemuList, apiToken); err != nil {
		log.Printf("Failed to process the request for %s - error %v", customUrl, err)
		return NodeGuestOverview{}, err
	}

	for i := range qemuList.Data {
		qemuList.Data[i].Node = node
		qemuList.Data[i].Parent = parent
	}

	return qemuList, nil

}

func detailedHostOverview(c *gin.Context) {
	parentName := c.Param("parent")
	parentObjects, err := convertJSON()
	if err != nil {
		log.Printf("Failed to convert the JSON values - %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to obtain the JSON data from the environment variable - %v", err)})
		return
	}
	var selectedObj PVEConnectionObject
	var found bool

	var (
		results []NodeDetails
		errors  []ApiError
	)

	for _, host := range parentObjects {
		if host.Parent == parentName {
			selectedObj = host
			found = true
			break
		}
	}

	if found {
		parentNodes, err := getParentNodes(selectedObj.Parent, selectedObj.Port, selectedObj.Token)
		if err != nil {
			errors = append(errors, ApiError{
				Parent:  selectedObj.Parent,
				Action:  "getParentNodes",
				Message: err.Error(),
			})
			log.Printf("Failed to get nodes for the parent %s - %v", selectedObj.Parent, err)
			c.JSON(http.StatusInternalServerError, NodeDetailsResponse{
				Data:   results,
				Errors: errors,
			})
			return
		}

		for _, node := range parentNodes.Data {
			var details NodeDetails

			if node.NodeStatus != "online" {
				errors = append(errors, ApiError{
					Parent:  selectedObj.Parent,
					Node:    node.Node,
					Action:  "onlineStatus",
					Message: fmt.Sprintf("The node %s appears to be offline", node.Node),
				})
				log.Printf("Skipping node %s - its offline", node.Node)
				continue
			}

			nodeStatus, err := getNodeStatus(selectedObj.Parent, selectedObj.Port, selectedObj.Token, node.Node)
			if err != nil {
				errors = append(errors, ApiError{
					Parent:  selectedObj.Parent,
					Node:    node.Node,
					Action:  "getNodeStatus",
					Message: err.Error(),
				})
				log.Printf("Failed to get the node status for %s - %v", node.Node, err)
				continue
			}

			NodeDnsObject, err := getNodeDnsObject(selectedObj.Parent, selectedObj.Port, selectedObj.Token, node.Node)
			if err != nil {
				errors = append(errors, ApiError{
					Parent:  selectedObj.Parent,
					Node:    node.Node,
					Action:  "getNodeDnsObject",
					Message: err.Error(),
				})
				log.Printf("Failed to get the node DNS for %s - %v", node.Node, err)
				continue
			}

			NodeTimeObject, err := getNodeTimeObject(selectedObj.Parent, selectedObj.Port, selectedObj.Token, node.Node)
			if err != nil {
				errors = append(errors, ApiError{
					Parent:  selectedObj.Parent,
					Node:    node.Node,
					Action:  "getNodeTimeObject",
					Message: err.Error(),
				})
				log.Printf("Failed to get the time configuration for %s - %v", node.Node, err)
				continue
			}

			details.NodeInfo.Node = node.Node
			details.NodeInfo.Parent = selectedObj.Parent
			details.NodeInfo.NodeStatus = node.NodeStatus
			details.Dns.Dns1 = NodeDnsObject.Data.Dns1
			details.Dns.Dns2 = NodeDnsObject.Data.Dns2
			details.Dns.Dns3 = NodeDnsObject.Data.Dns3
			details.Dns.Search = NodeDnsObject.Data.Search
			details.BootInfo.Mode = nodeStatus.Data.BootInfo.Mode
			details.BootInfo.Secureboot = nodeStatus.Data.BootInfo.Secureboot
			details.Cpuinfo.Cpus = nodeStatus.Data.Cpuinfo.Cpus
			details.Cpuinfo.Cores = nodeStatus.Data.Cpuinfo.Cores
			details.Cpuinfo.Model = nodeStatus.Data.Cpuinfo.Model
			details.Cpuinfo.Mhz = nodeStatus.Data.Cpuinfo.Mhz
			details.Cpuinfo.Sockets = nodeStatus.Data.Cpuinfo.Sockets
			details.Pveversion = nodeStatus.Data.Pveversion
			details.CurrentKernel.Release = nodeStatus.Data.CurrentKernel.Release
			details.CurrentKernel.Machine = nodeStatus.Data.CurrentKernel.Machine
			details.Time.Time = time.Unix(int64(NodeTimeObject.Data.Time), 0)
			details.Time.Timezone = NodeTimeObject.Data.Timezone
			results = append(results, details)

		}

		c.JSON(http.StatusOK, NodeDetailsResponse{
			Data:   results,
			Errors: errors,
		})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("The parent entered (%s) is not present in the environment variable - please adjust.", parentName)})
	}
}

func getNodeStatus(parent string, port int, apiToken string, node string) (NodeStatusObject, error) {
	customUrl := fmt.Sprintf("https://%s:%d/api2/json/nodes/%s/status", parent, port, node)
	req, err := http.NewRequest(http.MethodGet, customUrl, nil)
	if err != nil {
		log.Printf("Failed to create the HTTP request for getNodeStatus - %v", err)
		return NodeStatusObject{}, err
	}

	var jsonObject NodeStatusObject
	if err := sendRequest(req, &jsonObject, apiToken); err != nil {
		log.Printf("Failed to process the request for %s - error %v", customUrl, err)
		return NodeStatusObject{}, err
	}

	return jsonObject, nil
}

func getNodeDnsObject(parent string, port int, apiToken string, node string) (NodeDnsObject, error) {
	customUrl := fmt.Sprintf("https://%s:%d/api2/json/nodes/%s/dns", parent, port, node)
	req, err := http.NewRequest(http.MethodGet, customUrl, nil)
	if err != nil {
		log.Printf("Failed to create the HTTP request for the getNodeDnsObject - %v", err)
		return NodeDnsObject{}, err
	}

	var jsonObject NodeDnsObject
	if err := sendRequest(req, &jsonObject, apiToken); err != nil {
		log.Printf("Failed to process the request for %s - error %v", customUrl, err)
		return NodeDnsObject{}, err
	}

	return jsonObject, nil
}

func getNodeTimeObject(parent string, port int, apiToken string, node string) (NodeTimeObject, error) {
	customUrl := fmt.Sprintf("https://%s:%d/api2/json/nodes/%s/time", parent, port, node)
	req, err := http.NewRequest(http.MethodGet, customUrl, nil)
	if err != nil {
		log.Printf("Failed to create the HTTP request for the getNodeTimeObject - %v", err)
		return NodeTimeObject{}, err
	}
	var jsonObject NodeTimeObject
	if err := sendRequest(req, &jsonObject, apiToken); err != nil {
		log.Printf("Failed to process the request for %s - error %v", customUrl, err)
		return NodeTimeObject{}, err
	}

	return jsonObject, nil

}

func getNodeStorageOverview(c *gin.Context) {
	parentName := c.Param("parent")
	parentObjects, err := convertJSON()
	if err != nil {
		log.Printf("Failed to convert the JSON objects - %v", err)
		return
	}

	var selectedObj PVEConnectionObject
	var found bool
	for _, host := range parentObjects {
		if host.Parent == parentName {
			selectedObj = host
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("The parent entered (%s) is not present in the environment variable - please adjust.", parentName)})
		return
	}

	var (
		storageList NodeStorageResponse
		errors      []ApiError
	)

	portOpen, err := testHostPort(selectedObj.Parent, selectedObj.Port)
	if err != nil {
		errors = append(errors, ApiError{
			Parent:  selectedObj.Parent,
			Action:  "testHostPort",
			Message: err.Error(),
		})
		log.Printf("Failed to check if the port was open on %s:%d - %v", selectedObj.Parent, selectedObj.Port, err)
		c.JSON(http.StatusInternalServerError, NodeStorageResponse{
			Data:   storageList.Data,
			Errors: errors,
		})
		return
	}

	if portOpen {

		parentNodes, err := getParentNodes(selectedObj.Parent, selectedObj.Port, selectedObj.Token)
		if err != nil {
			errors = append(errors, ApiError{
				Parent:  selectedObj.Parent,
				Action:  "testHostPort",
				Message: err.Error(),
			})
			log.Printf("Failed to obtain the cluster nodes for %s - %v", selectedObj.Parent, err)
			c.JSON(http.StatusInternalServerError, NodeStorageResponse{
				Data:   storageList.Data,
				Errors: errors,
			})
			return
		}

		for _, node := range parentNodes.Data {
			if node.NodeStatus != "online" {
				errors = append(errors, ApiError{
					Parent:  selectedObj.Parent,
					Node:    node.Node,
					Action:  "onlineStatus",
					Message: fmt.Sprintf("The node %s appears to be offline", node.Node),
				})
				log.Printf("Skipping node %s - its offline", node.Node)
				continue
			}

			nodeStorage, err := getNodeStorage(selectedObj.Parent, selectedObj.Port, selectedObj.Token, node.Node)
			if err != nil {
				errors = append(errors, ApiError{
					Parent:  selectedObj.Parent,
					Node:    node.Node,
					Action:  "getNodeStorage",
					Message: err.Error(),
				})
				log.Printf("Failed to obtain the storage for %s - %v", node.Node, err)
				continue
			}

			for _, storage := range nodeStorage.Data {
				var details NodeStorageInfo
				details.Parent = selectedObj.Parent
				details.Node = node.Node
				details.NodeStatus = node.NodeStatus
				details.Active = storage.Active
				details.Content = storage.Content
				details.Enabled = storage.Enabled
				details.Shared = storage.Shared
				details.Type = storage.Type
				details.Storage = storage.Storage
				details.TotalGb = storage.Total / (1024 * 1024 * 1024)
				details.AvailableGb = storage.Available / (1024 * 1024 * 1024)
				details.UsedGb = storage.Used / (1024 * 1024 * 1024)

				storageList.Data = append(storageList.Data, details)
			}

		}
	} else {
		c.JSON(http.StatusBadGateway, NodeStorageResponse{
			Data: storageList.Data,
			Errors: append(errors, ApiError{
				Parent:  selectedObj.Parent,
				Action:  "testHostPort",
				Message: fmt.Sprintf("port %d closed", selectedObj.Port),
			}),
		})
		return
	}

	c.JSON(http.StatusOK, NodeStorageResponse{
		Data:   storageList.Data,
		Errors: errors,
	})
}

func getNodeStorage(parent string, port int, apiToken string, node string) (hostStorageList, error) {
	customUrl := fmt.Sprintf("https://%s:%d/api2/json/nodes/%s/storage", parent, port, node)

	var nodeStorageObj hostStorageList
	req, err := http.NewRequest(http.MethodGet, customUrl, nil)
	if err != nil {
		log.Printf("Failed to create HTTP request for getNodeStorage on %s - %v", customUrl, err)
		return hostStorageList{}, err
	}

	if err := sendRequest(req, &nodeStorageObj, apiToken); err != nil {
		log.Printf("Failed to process the request for %s - error %v", customUrl, err)
		return hostStorageList{}, err
	}
	return nodeStorageObj, nil
}

func getNodeDisks(parent string, port int, apiToken string, node string) (hostDiskList, error) {
	customUrl := fmt.Sprintf("https://%s:%d/api2/json/nodes/%s/disks/list", parent, port, node)

	var nodeStorageObj hostDiskList

	req, err := http.NewRequest(http.MethodGet, customUrl, nil)
	if err != nil {
		log.Printf("Failed to create HTTP request for getNodeDisks on %s - %v", customUrl, err)
		return hostDiskList{}, err
	}

	if err := sendRequest(req, &nodeStorageObj, apiToken); err != nil {
		log.Printf("Failed to process the request for %s - error %v", customUrl, err)
		return hostDiskList{}, err
	}

	return nodeStorageObj, nil
}

func getNodeDiskOverview(c *gin.Context) {
	parentName := c.Param("parent")
	parentObjects, err := convertJSON()
	if err != nil {
		log.Printf("Failed to convert the JSON objects - %v", err)
		return
	}

	var selectedObj PVEConnectionObject
	var found bool
	for _, host := range parentObjects {
		if host.Parent == parentName {
			selectedObj = host
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("The parent entered (%s) is not present in the environment variable - please adjust.", parentName)})
		return
	}

	var (
		diskList NodeDiskObject
		errors   []ApiError
	)

	portOpen, err := testHostPort(selectedObj.Parent, selectedObj.Port)
	if err != nil {
		errors = append(errors, ApiError{
			Parent:  selectedObj.Parent,
			Action:  "testHostPort",
			Message: err.Error(),
		})
		log.Printf("Failed to check if the port was open on %s:%d - %v", selectedObj.Parent, selectedObj.Port, err)
		c.JSON(http.StatusBadGateway, NodeDiskObject{
			Data:   diskList.Data,
			Errors: errors,
		})
		return
	}

	if portOpen {
		parentNodes, err := getParentNodes(selectedObj.Parent, selectedObj.Port, selectedObj.Token)
		if err != nil {
			log.Printf(
				"Failed to obtain the nodes for the parent %s - %v", selectedObj.Parent, err)
			c.JSON(http.StatusInternalServerError, NodeDiskObject{
				Data: diskList.Data,
				Errors: append(errors, ApiError{
					Parent:  selectedObj.Parent,
					Action:  "getParentNodes",
					Message: err.Error(),
				}),
			})
			return
		}

		ch := make(chan []NodeDiskInfo, len(parentNodes.Data))
		errCh := make(chan ApiError, len(parentNodes.Data)*2)
		for _, node := range parentNodes.Data {
			n := node
			go func() {
				var disks []NodeDiskInfo
				if n.NodeStatus != "online" {
					errCh <- ApiError{
						Parent:  selectedObj.Parent,
						Node:    n.Node,
						Action:  "onlineStatus",
						Message: fmt.Sprintf("The node %s appears to be offline", n.Node),
					}
					// Sends an empty response, so the the channel is not blocked
					ch <- disks
					log.Printf("Skipping node %s - its offline", n.Node)
					return
				}
				var nodeStorage hostDiskList
				nodeStorage, err := getNodeDisks(selectedObj.Parent, selectedObj.Port, selectedObj.Token, n.Node)
				if err != nil {
					errCh <- ApiError{
						Parent:  selectedObj.Parent,
						Node:    n.Node,
						Action:  "getNodeDisks",
						Message: err.Error(),
					}
					log.Printf("Failed to obtain the disks for %s - %v", n.Node, err)
					return
				}

				for _, disk := range nodeStorage.Data {

					var details NodeDiskInfo

					switch val := disk.Wearout.(type) {
					case float64:
						details.Wearout = int(val)
					case string:
						details.Wearout = int(100)
					}

					switch val := disk.Rpm.(type) {
					case float64:
						details.Rpm = int(val)
					case string:
						i, err := strconv.Atoi(val)
						if err != nil {
							log.Printf("Failed to convert the RPM value for %s - %v", n.Node, err)
						}
						details.Rpm = i
					}

					details.Parent = selectedObj.Parent
					details.Node = n.Node
					details.NodeStatus = n.NodeStatus
					details.Gpt = disk.Gpt
					details.Vendor = disk.Vendor
					details.Devpath = disk.Devpath
					details.Health = disk.Health
					details.Type = disk.Type
					details.Serial = disk.Serial
					details.Model = disk.Model
					details.SizeGb = disk.Size / (1024 * 1024 * 1024)

					disks = append(disks, details)

				}

				ch <- disks

			}()
		}

		for range parentNodes.Data {
			batch := <-ch
			diskList.Data = append(diskList.Data, batch...)
		}

		for len(errCh) > 0 {
			errors = append(errors, <-errCh)
		}

		c.JSON(http.StatusOK, NodeDiskObject{
			Data:   diskList.Data,
			Errors: errors,
		})
	} else {
		c.JSON(http.StatusBadGateway, NodeDiskObject{
			Data: diskList.Data,
			Errors: append(errors, ApiError{
				Parent:  selectedObj.Parent,
				Action:  "testHostPort",
				Message: fmt.Sprintf("port %d closed", selectedObj.Port),
			}),
		})
		return
	}

}
