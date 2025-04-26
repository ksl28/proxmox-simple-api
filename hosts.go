package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func testHostPort(host string, port int) (bool, error) {
	hostPort := fmt.Sprintf("%v:%v", host, port)
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
		portOpen, err := testHostPort(obj.Name, obj.Port)
		if err != nil {
			log.Printf("Failed to check if the port %d for %s is open - %v", obj.Port, obj.Name, err)
			continue
		}
		if portOpen {
			datacenterNodes, err := getDatacenterNodes(obj.Name, obj.Port, obj.Token)
			if err != nil {
				log.Printf("Failed to check if the port %d for %s is open - %v", obj.Port, obj.Name, err)
				continue
			}
			for _, xv := range datacenterNodes.Data {
				var temporary nodeSummaryWrapper
				// Needs review - omitempty is used, so might not need to have this
				if xv.Status != "online" {
					temporary.Parent = obj.Name
					temporary.Node = xv.Node
					temporary.Status = xv.Status
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
					temporary.Parent = obj.Name
					temporary.Node = xv.Node
					temporary.Status = xv.Status
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
			log.Printf("The server %s is not listening on %d", obj.Name, obj.Port)
			continue
		}
	}
	c.JSON(http.StatusOK, results)
}

func getDatacenterNodes(parent string, port int, apiToken string) (nodesStruct, error) {
	requestUrl := fmt.Sprintf("https://%s:%d/api2/json/nodes", parent, port)
	req, err := http.NewRequest(http.MethodGet, requestUrl, nil)
	if err != nil {
		log.Printf("Error creating the HTTP request for %s - %v", requestUrl, err)
		return nodesStruct{}, err
	}
	req.Header.Add("Authorization", apiToken)
	res, err := httpClient.Do(req)
	if err != nil {
		log.Printf("Failed to perform the HTTP request for %s - %v", parent, err)
		return nodesStruct{}, err
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Failed to read the http body for %s - %v", parent, err)
		return nodesStruct{}, err
	}

	var datacenterNodes nodesStruct
	err = json.Unmarshal(raw, &datacenterNodes)
	if err != nil {
		log.Printf("Failed to unmarshal the http body for %s - %v", parent, err)
		return nodesStruct{}, err
	}
	return datacenterNodes, nil
}

func nodeGuestsOverview(guestType string, parent string, port int, node string, apiToken string) (NodeGuestId, error) {

	// Check this later - should be used in the vmSummary instead?

	var requestUrl string
	if guestType == "lxc" {
		requestUrl = fmt.Sprintf("https://%s:%d/api2/json/nodes/%s/lxc", parent, port, node)
	} else {
		requestUrl = fmt.Sprintf("https://%s:%d/api2/json/nodes/%s/qemu", parent, port, node)
	}

	req, err := http.NewRequest(http.MethodGet, requestUrl, nil)
	if err != nil {
		log.Printf("Failed to create the HTTP request for %s - %v", parent, err)
		return NodeGuestId{}, err
	}
	req.Header.Add("Authorization", apiToken)
	res, err := httpClient.Do(req)
	if err != nil {
		log.Printf("Failed to perform the HTTP request for %s - %v", parent, err)
		return NodeGuestId{}, err
	}
	defer res.Body.Close()

	raw, err := io.ReadAll(res.Body)
	if err != nil {
		fmt.Printf("Failed - %v", err)
	}

	var qemuList NodeGuestId
	err = json.Unmarshal(raw, &qemuList)
	if err != nil {
		fmt.Printf("JSON failed - %v", err)
	}

	for i, _ := range qemuList.Data {
		qemuList.Data[i].Node = node
		qemuList.Data[i].Parent = parent
	}

	return qemuList, nil

}

func detailedHostOverview(c *gin.Context) {
	hostName := c.Param("name")
	rawObj, err := convertJSON()
	if err != nil {
		log.Printf("Failed to convert the JSON values - %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Failed to obtain the JSON data from the environment variable - %v", err)})
		return
	}
	var selectedObj PVEObject
	var found bool
	for _, obj := range rawObj {
		if obj.Name == hostName {
			selectedObj = obj
			found = true
			break
		}
	}

	if found {
		clusterData, err := getDatacenterNodes(selectedObj.Name, selectedObj.Port, selectedObj.Token)
		if err != nil {
			log.Printf("Failed to get cluster nodes - %v", err)
			return
		}

		clusterNodes := clusterData.Data

		var results []DataObject
		for _, obj := range clusterNodes {

			if obj.Status != "online" {
				log.Printf("Skipping node %s - its offline", obj.Node)
				continue
			}

			var detailedHost DataObject

			nodeStatus, err := getNodeStatus(selectedObj.Name, selectedObj.Port, selectedObj.Token, obj.Node)
			if err != nil {
				log.Printf("Failed to get the node status for %s - %v", obj.Node, err)
				continue
			}

			nodeDns, err := getNodeDns(selectedObj.Name, selectedObj.Port, selectedObj.Token, obj.Node)
			if err != nil {
				log.Printf("Failed to get the node DNS for %s - %v", obj.Node, err)
				continue
			}

			nodeTime, err := getNodeTime(selectedObj.Name, selectedObj.Port, selectedObj.Token, obj.Node)
			if err != nil {
				log.Printf("Failed to get the time configuration for %s - %v", obj.Node, err)
				continue
			}

			detailedHost.Data.NodeInfo.Node = obj.Node
			detailedHost.Data.NodeInfo.Parent = selectedObj.Name
			detailedHost.Data.Dns.Dns1 = nodeDns.Data.Dns1
			detailedHost.Data.Dns.Dns2 = nodeDns.Data.Dns2
			detailedHost.Data.Dns.Dns3 = nodeDns.Data.Dns3
			detailedHost.Data.Dns.Search = nodeDns.Data.Search
			detailedHost.Data.BootInfo.Mode = nodeStatus.Data.BootInfo.Mode
			detailedHost.Data.BootInfo.Secureboot = nodeStatus.Data.BootInfo.Secureboot
			detailedHost.Data.Cpuinfo.Cpus = nodeStatus.Data.Cpuinfo.Cpus
			detailedHost.Data.Cpuinfo.Cores = nodeStatus.Data.Cpuinfo.Cores
			detailedHost.Data.Cpuinfo.Model = nodeStatus.Data.Cpuinfo.Model
			detailedHost.Data.Cpuinfo.Mhz = nodeStatus.Data.Cpuinfo.Mhz
			detailedHost.Data.Cpuinfo.Sockets = nodeStatus.Data.Cpuinfo.Sockets
			detailedHost.Data.Pveversion = nodeStatus.Data.Pveversion
			detailedHost.Data.CurrentKernel.Release = nodeStatus.Data.CurrentKernel.Release
			detailedHost.Data.CurrentKernel.Machine = nodeStatus.Data.CurrentKernel.Machine
			detailedHost.Data.Time.Time = time.Unix(int64(nodeTime.Data.Time), 0)
			detailedHost.Data.Time.Timezone = nodeTime.Data.Timezone
			results = append(results, detailedHost)

		}

		c.JSON(http.StatusOK, results)
	} else {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("The parent entered (%s) is not present in the environment variable - please adjust.", hostName)})
	}
}

func getNodeStatus(parent string, port int, apiToken string, node string) (nodeStatus, error) {
	customUrl := fmt.Sprintf("https://%s:%d/api2/json/nodes/%s/status", parent, port, node)
	req, err := http.NewRequest(http.MethodGet, customUrl, nil)
	if err != nil {
		log.Printf("Failed to create the HTTP request for getNodeStatus - %v", err)
		return nodeStatus{}, err
	}
	req.Header.Add("Authorization", apiToken)

	res, err := httpClient.Do(req)
	if err != nil {
		log.Printf("Failed to perform the HTTP request for getNodeStatus - %s - %v", customUrl, err)
		return nodeStatus{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Failed to read the body for getNodeStatus - %s - %v", customUrl, err)
		return nodeStatus{}, err
	}

	var jsonObject nodeStatus
	if err := json.Unmarshal(body, &jsonObject); err != nil {
		log.Printf("Failed to convert the body for getNodeStatus - %s - %v", customUrl, err)
		return nodeStatus{}, err
	}

	return jsonObject, nil
}

func getNodeDns(parent string, port int, apiToken string, node string) (nodeDns, error) {
	customUrl := fmt.Sprintf("https://%s:%d/api2/json/nodes/%s/dns", parent, port, node)
	req, err := http.NewRequest(http.MethodGet, customUrl, nil)
	if err != nil {
		log.Printf("Failed to create the HTTP request for the getNodeDns - %v", err)
		return nodeDns{}, err
	}
	req.Header.Add("Authorization", apiToken)
	res, err := httpClient.Do(req)
	if err != nil {
		log.Printf("Failed to perform the HTTP request for getNodeDns - %s - %v", customUrl, err)
		return nodeDns{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Failed to read the body for getNodeDns - %s - %v", customUrl, err)
		return nodeDns{}, err
	}

	var jsonObject nodeDns
	if err := json.Unmarshal(body, &jsonObject); err != nil {
		log.Printf("Failed to convert the body for getNodeDns - %s - %v", customUrl, err)
		return nodeDns{}, err
	}

	return jsonObject, nil
}

func getNodeTime(parent string, port int, apiToken string, node string) (nodeTime, error) {
	customUrl := fmt.Sprintf("https://%s:%d/api2/json/nodes/%s/time", parent, port, node)
	req, err := http.NewRequest(http.MethodGet, customUrl, nil)
	if err != nil {
		log.Printf("Failed to create the HTTP request for the getNodeTime - %v", err)
		return nodeTime{}, err
	}
	req.Header.Add("Authorization", apiToken)
	res, err := httpClient.Do(req)
	if err != nil {
		log.Printf("Failed to perform the HTTP request for getNodeTime - %s - %v", customUrl, err)
		return nodeTime{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Failed to read the body for getNodeTime - %s - %v", customUrl, err)
		return nodeTime{}, err
	}

	var jsonObject nodeTime
	if err := json.Unmarshal(body, &jsonObject); err != nil {
		log.Printf("Failed to convert the body for getNodeTime - %s - %v", customUrl, err)
		return nodeTime{}, err
	}

	return jsonObject, nil
}

func getNodeStorageOverview(c *gin.Context) {
	hostName := c.Param("name")
	clusterObjects, err := convertJSON()
	if err != nil {
		log.Printf("Failed to convert the JSON objects - %v", err)
		return
	}

	var selectedObj PVEObject
	var found bool
	for _, clusterNode := range clusterObjects {
		if clusterNode.Name == hostName {
			selectedObj = clusterNode
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("The parent entered (%s) is not present in the environment variable - please adjust.", hostName)})
		return
	}
	var storageList hostStorageListWrapper

	portOpen, err := testHostPort(selectedObj.Name, selectedObj.Port)
	if err != nil {
		log.Printf("Failed to check if the port was open on %s:%d - %v", selectedObj.Name, selectedObj.Port, err)
		return
	}

	if portOpen {

		clusterNodes, err := getDatacenterNodes(selectedObj.Name, selectedObj.Port, selectedObj.Token)
		if err != nil {
			log.Printf("Failed to obtain the cluster nodes for %s - %v", selectedObj.Name, err)
			return
		}

		for _, clusterNode := range clusterNodes.Data {
			if clusterNode.Status != "online" {
				log.Printf("Skipping node %s - its offline", clusterNode.Node)
				continue
			}

			var nodeStorage hostStorageList
			nodeStorage, err := getNodeStorage(selectedObj.Name, selectedObj.Port, selectedObj.Token, clusterNode.Node)
			if err != nil {
				log.Printf("Failed to obtain the storage for %s - %v", clusterNode.Node, err)
				continue
			}

			for _, v := range nodeStorage.Data {
				var temporary hostStorageEntry
				temporary.Parent = selectedObj.Name
				temporary.Node = clusterNode.Node
				temporary.Active = v.Active
				temporary.Content = v.Content
				temporary.Enabled = v.Enabled
				temporary.Shared = v.Shared
				temporary.Storage = v.Storage
				temporary.TotalGb = v.Total / (1024 * 1024 * 1024)
				temporary.AvailableGb = v.Available / (1024 * 1024 * 1024)
				temporary.UsedGb = v.Used / (1024 * 1024 * 1024)

				storageList.Data = append(storageList.Data, temporary)
			}

		}
	}

	c.JSON(http.StatusOK, storageList)
}

func getNodeStorage(parent string, port int, apiToken string, node string) (hostStorageList, error) {
	customUrl := fmt.Sprintf("https://%s:%d/api2/json/nodes/%s/storage", parent, port, node)

	var nodeStorageObj hostStorageList

	req, err := http.NewRequest(http.MethodGet, customUrl, nil)
	if err != nil {
		log.Printf("Failed to create HTTP request for getNodeStorage on %s - %v", customUrl, err)
		return hostStorageList{}, err
	}
	req.Header.Add("Authorization", apiToken)

	res, err := httpClient.Do(req)
	if err != nil {
		log.Printf("Failed to perform HTTP request for getNodeStorage on %s - %v", customUrl, err)
		return hostStorageList{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Failed to read response body for getNodeStorage on %s - %v", customUrl, err)
		return hostStorageList{}, err
	}

	if err := json.Unmarshal(body, &nodeStorageObj); err != nil {
		log.Printf("Failed to unmarshal JSON response for getNodeStorage on %s - %v", customUrl, err)
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
	req.Header.Add("Authorization", apiToken)

	res, err := httpClient.Do(req)
	if err != nil {
		log.Printf("Failed to perform HTTP request for getNodeDisks on %s - %v", customUrl, err)
		return hostDiskList{}, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Failed to read response body for getNodeDisks on %s - %v", customUrl, err)
		return hostDiskList{}, err
	}

	if err := json.Unmarshal(body, &nodeStorageObj); err != nil {
		log.Printf("Failed to unmarshal JSON response for getNodeDisks on %s - %v", customUrl, err)
		return hostDiskList{}, err
	}

	return nodeStorageObj, nil
}

func getNodeDiskOverview(c *gin.Context) {
	hostName := c.Param("name")

	clusterObjects, err := convertJSON()
	if err != nil {
		log.Printf("Failed to convert the JSON objects - %v", err)
		return
	}

	var selectedObj PVEObject
	var found bool
	for _, clusterNode := range clusterObjects {
		if clusterNode.Name == hostName {
			selectedObj = clusterNode
			found = true
			break
		}
	}

	if !found {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("The parent entered (%s) is not present in the environment variable - please adjust.", hostName)})
		return
	}

	var diskList hostDiskEntryWrapper

	portOpen, err := testHostPort(selectedObj.Name, selectedObj.Port)
	if err != nil {
		log.Printf("Failed to check if the port was open on %s:%d - %v", selectedObj.Name, selectedObj.Port, err)
		return
	}

	if portOpen {

		clusterNodes, err := getDatacenterNodes(selectedObj.Name, selectedObj.Port, selectedObj.Token)
		if err != nil {
			log.Printf("Failed to obtain the cluster nodes for %s - %v", selectedObj.Name, err)
			return
		}

		for _, clusterNode := range clusterNodes.Data {
			if err != nil {
				log.Printf("Failed to obtain the cluster nodes for %s - %v", selectedObj.Name, err)
				return
			}
			var nodeStorage hostDiskList
			nodeStorage, err := getNodeDisks(selectedObj.Name, selectedObj.Port, selectedObj.Token, clusterNode.Node)
			if err != nil {
				log.Printf("Failed to obtain the disks for %s - %v", clusterNode.Node, err)
				continue
			}

			for _, v := range nodeStorage.Data {
				var temporary hostDiskEntry

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
						log.Printf("Failed to convert the RPM value for %s - %v", clusterNode.Node, err)
					}
					temporary.Rpm = i
				}

				temporary.Parent = selectedObj.Name
				temporary.Node = clusterNode.Node
				temporary.Gpt = v.Gpt
				temporary.Vendor = v.Vendor
				temporary.Devpath = v.Devpath
				temporary.Health = v.Health
				temporary.Type = v.Type
				temporary.Serial = v.Serial
				temporary.Model = v.Model
				temporary.UsedGb = v.Size / (1024 * 1024 * 1024)

				diskList.Data = append(diskList.Data, temporary)
			}
		}
	}

	c.JSON(http.StatusOK, diskList)
}
