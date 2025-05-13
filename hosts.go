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
	var results []nodeSummaryWrapper

	jsonObjects, err := convertJSON()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read JSON data"})
		return
	}

	for _, obj := range jsonObjects {
		portOpen, err := testHostPort(obj.Parent, obj.Port)
		if err != nil {
			log.Printf("Failed to check if the port %d for %s is open - %v", obj.Port, obj.Parent, err)
			continue
		}
		if portOpen {
			datacenterNodes, err := getDatacenterNodes(obj.Parent, obj.Port, obj.Token)
			if err != nil {
				log.Printf("Failed to check if the port %d for %s is open - %v", obj.Port, obj.Parent, err)
				continue
			}
			for _, xv := range datacenterNodes.Data {
				var temporary nodeSummaryWrapper
				//If the first object in the slice is empty / offline, then the struct will be limited to only show the fields that have values.
				if xv.NodeStatus != "online" {
					temporary.Parent = obj.Parent
					temporary.Node = xv.Node
					temporary.NodeStatus = xv.NodeStatus
					temporary.MaxCPU = 0
					temporary.MaxMemGb = 0
					temporary.MemGb = 0
					temporary.UptimeHours = 0
					temporary.Cpu = 0
					temporary.MaxRootDiskGb = 0
					temporary.RootDiskGb = 0
					results = append(results, temporary)
					continue
				} else {
					temporary.Parent = obj.Parent
					temporary.Node = xv.Node
					temporary.NodeStatus = xv.NodeStatus
					temporary.MaxCPU = xv.MaxCPU
					temporary.MaxMemGb = int(xv.MaxMem) / (1024 * 1024 * 1024)
					temporary.MemGb = int(xv.Mem) / (1024 * 1024 * 1024)
					temporary.UptimeHours = int(xv.UptimeHours) / (60 * 60)
					temporary.Cpu = math.Round(xv.Cpu * 100)
					temporary.MaxRootDiskGb = xv.Maxdisk / (1000 * 1000 * 1000)
					temporary.RootDiskGb = xv.Disk / (1000 * 1000 * 1000)
					results = append(results, temporary)
				}

			}
		} else {
			log.Printf("The server %s is not listening on %d", obj.Parent, obj.Port)
			continue
		}
	}
	c.JSON(http.StatusOK, results)
}

func getDatacenterNodes(parent string, port int, apiToken string) (PVENodesObject, error) {
	customUrl := fmt.Sprintf("https://%s:%d/api2/json/nodes", parent, port)
	req, err := http.NewRequest(http.MethodGet, customUrl, nil)
	if err != nil {
		log.Printf("Error creating the HTTP request for %s - %v", customUrl, err)
		return PVENodesObject{}, err
	}
	var datacenterNodes PVENodesObject
	if err := sendRequest(req, &datacenterNodes, apiToken); err != nil {
		log.Printf("Failed to process the request for %s - error %v", customUrl, err)
		return PVENodesObject{}, err
	}

	return datacenterNodes, nil
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
	hostName := c.Param("parent")
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

	for _, obj := range parentObjects {
		if obj.Parent == hostName {
			selectedObj = obj
			found = true
			break
		}
	}

	if found {
		clusterData, err := getDatacenterNodes(selectedObj.Parent, selectedObj.Port, selectedObj.Token)
		if err != nil {
			errors = append(errors, ApiError{
				Parent:  selectedObj.Parent,
				Action:  "getDatacenterNodes",
				Message: err.Error(),
			})
			log.Printf("Failed to get cluster nodes - %v", err)
			c.JSON(http.StatusInternalServerError, NodeDetailsResponse{
				Data:   results,
				Errors: errors,
			})
			return
		}

		clusterNodes := clusterData.Data

		for _, obj := range clusterNodes {
			var detailedHost NodeDetails

			if obj.NodeStatus != "online" {
				detailedHost.NodeInfo.Node = obj.Node
				detailedHost.NodeInfo.Parent = selectedObj.Parent
				detailedHost.NodeInfo.NodeStatus = obj.NodeStatus
				results = append(results, detailedHost)
				log.Printf("Skipping node %s - its offline", obj.Node)
				continue
			}

			nodeStatus, err := getNodeStatus(selectedObj.Parent, selectedObj.Port, selectedObj.Token, obj.Node)
			if err != nil {
				errors = append(errors, ApiError{
					Parent:  selectedObj.Parent,
					Node:    obj.Node,
					Action:  "getNodeStatus",
					Message: err.Error(),
				})
				log.Printf("Failed to get the node status for %s - %v", obj.Node, err)
				continue
			}

			NodeDnsObject, err := getNodeDnsObject(selectedObj.Parent, selectedObj.Port, selectedObj.Token, obj.Node)
			if err != nil {
				errors = append(errors, ApiError{
					Parent:  selectedObj.Parent,
					Node:    obj.Node,
					Action:  "getNodeDnsObject",
					Message: err.Error(),
				})
				log.Printf("Failed to get the node DNS for %s - %v", obj.Node, err)
				continue
			}

			NodeTimeObject, err := getNodeTimeObject(selectedObj.Parent, selectedObj.Port, selectedObj.Token, obj.Node)
			if err != nil {
				errors = append(errors, ApiError{
					Parent:  selectedObj.Parent,
					Node:    obj.Node,
					Action:  "getNodeTimeObject",
					Message: err.Error(),
				})
				log.Printf("Failed to get the time configuration for %s - %v", obj.Node, err)
				continue
			}

			detailedHost.NodeInfo.Node = obj.Node
			detailedHost.NodeInfo.Parent = selectedObj.Parent
			detailedHost.NodeInfo.NodeStatus = obj.NodeStatus
			detailedHost.Dns.Dns1 = NodeDnsObject.Data.Dns1
			detailedHost.Dns.Dns2 = NodeDnsObject.Data.Dns2
			detailedHost.Dns.Dns3 = NodeDnsObject.Data.Dns3
			detailedHost.Dns.Search = NodeDnsObject.Data.Search
			detailedHost.BootInfo.Mode = nodeStatus.Data.BootInfo.Mode
			detailedHost.BootInfo.Secureboot = nodeStatus.Data.BootInfo.Secureboot
			detailedHost.Cpuinfo.Cpus = nodeStatus.Data.Cpuinfo.Cpus
			detailedHost.Cpuinfo.Cores = nodeStatus.Data.Cpuinfo.Cores
			detailedHost.Cpuinfo.Model = nodeStatus.Data.Cpuinfo.Model
			detailedHost.Cpuinfo.Mhz = nodeStatus.Data.Cpuinfo.Mhz
			detailedHost.Cpuinfo.Sockets = nodeStatus.Data.Cpuinfo.Sockets
			detailedHost.Pveversion = nodeStatus.Data.Pveversion
			detailedHost.CurrentKernel.Release = nodeStatus.Data.CurrentKernel.Release
			detailedHost.CurrentKernel.Machine = nodeStatus.Data.CurrentKernel.Machine
			detailedHost.Time.Time = time.Unix(int64(NodeTimeObject.Data.Time), 0)
			detailedHost.Time.Timezone = NodeTimeObject.Data.Timezone
			results = append(results, detailedHost)

		}

		c.JSON(http.StatusOK, NodeDetailsResponse{
			Data:   results,
			Errors: errors,
		})
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("The parent entered (%s) is not present in the environment variable - please adjust.", hostName)})
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
	clusterObjects, err := convertJSON()
	if err != nil {
		log.Printf("Failed to convert the JSON objects - %v", err)
		return
	}

	var selectedObj PVEConnectionObject
	var found bool
	for _, clusterNode := range clusterObjects {
		if clusterNode.Parent == parentName {
			selectedObj = clusterNode
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

		clusterNodes, err := getDatacenterNodes(selectedObj.Parent, selectedObj.Port, selectedObj.Token)
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

		for _, clusterNode := range clusterNodes.Data {
			if clusterNode.NodeStatus != "online" {
				var temporary NodeStorageInfo
				temporary.Parent = selectedObj.Parent
				temporary.Node = clusterNode.Node
				temporary.NodeStatus = clusterNode.NodeStatus
				storageList.Data = append(storageList.Data, temporary)
				log.Printf("Skipping node %s - its offline", clusterNode.Node)
				continue
			}

			var nodeStorage hostStorageList
			nodeStorage, err := getNodeStorage(selectedObj.Parent, selectedObj.Port, selectedObj.Token, clusterNode.Node)
			if err != nil {
				errors = append(errors, ApiError{
					Parent:  selectedObj.Parent,
					Node:    clusterNode.Node,
					Action:  "getNodeStorage",
					Message: err.Error(),
				})
				log.Printf("Failed to obtain the storage for %s - %v", clusterNode.Node, err)
				continue
			}

			for _, v := range nodeStorage.Data {
				var temporary NodeStorageInfo
				temporary.Parent = selectedObj.Parent
				temporary.Node = clusterNode.Node
				temporary.NodeStatus = clusterNode.NodeStatus
				temporary.Active = v.Active
				temporary.Content = v.Content
				temporary.Enabled = v.Enabled
				temporary.Shared = v.Shared
				temporary.Type = v.Type
				temporary.Storage = v.Storage
				temporary.TotalGb = v.Total / (1024 * 1024 * 1024)
				temporary.AvailableGb = v.Available / (1024 * 1024 * 1024)
				temporary.UsedGb = v.Used / (1024 * 1024 * 1024)

				storageList.Data = append(storageList.Data, temporary)
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
	clusterObjects, err := convertJSON()
	if err != nil {
		log.Printf("Failed to convert the JSON objects - %v", err)
		return
	}

	var selectedObj PVEConnectionObject
	var found bool
	for _, clusterNode := range clusterObjects {
		if clusterNode.Parent == parentName {
			selectedObj = clusterNode
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
		clusterNodes, err := getDatacenterNodes(selectedObj.Parent, selectedObj.Port, selectedObj.Token)
		if err != nil {
			log.Printf(
				"Failed to obtain the cluster nodes for %s - %v", selectedObj.Parent, err)
			c.JSON(http.StatusInternalServerError, NodeDiskObject{
				Data: diskList.Data,
				Errors: append(errors, ApiError{
					Parent:  selectedObj.Parent,
					Action:  "getDatacenterNodes",
					Message: err.Error(),
				}),
			})
			return
		}

		ch := make(chan []NodeDiskInfo, len(clusterNodes.Data))
		errCh := make(chan ApiError, len(clusterNodes.Data)*2)
		for _, clusterNode := range clusterNodes.Data {
			node := clusterNode
			go func() {
				var disks []NodeDiskInfo
				if node.NodeStatus != "online" {
					errCh <- ApiError{
						Parent:  selectedObj.Parent,
						Node:    node.Node,
						Action:  "onlineStatus",
						Message: fmt.Sprintf("The node %s appears to be offline", node.Node),
					}
					log.Printf("Skipping node %s - its offline", node.Node)
					return
				}
				var nodeStorage hostDiskList
				nodeStorage, err := getNodeDisks(selectedObj.Parent, selectedObj.Port, selectedObj.Token, node.Node)
				if err != nil {
					errCh <- ApiError{
						Parent:  selectedObj.Parent,
						Node:    node.Node,
						Action:  "getNodeDisks",
						Message: err.Error(),
					}
					log.Printf("Failed to obtain the disks for %s - %v", node.Node, err)
					return
				}

				for _, v := range nodeStorage.Data {

					var temporary NodeDiskInfo

					switch val := v.Wearout.(type) {
					case float64:
						temporary.Wearout = int(val)
					case string:
						temporary.Wearout = int(100)
					}

					switch val := v.Rpm.(type) {
					case float64:
						temporary.Rpm = int(val)
					case string:
						i, err := strconv.Atoi(val)
						if err != nil {
							log.Printf("Failed to convert the RPM value for %s - %v", node.Node, err)
						}
						temporary.Rpm = i
					}

					temporary.Parent = selectedObj.Parent
					temporary.Node = node.Node
					temporary.NodeStatus = node.NodeStatus
					temporary.Gpt = v.Gpt
					temporary.Vendor = v.Vendor
					temporary.Devpath = v.Devpath
					temporary.Health = v.Health
					temporary.Type = v.Type
					temporary.Serial = v.Serial
					temporary.Model = v.Model
					temporary.SizeGb = v.Size / (1024 * 1024 * 1024)

					disks = append(disks, temporary)

				}

				ch <- disks

			}()
		}

		for range clusterNodes.Data {
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
