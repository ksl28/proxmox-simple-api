package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func lxcSummary(c *gin.Context) {
	parentObjects, err := convertJSON()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to convert the JSON data - %v", err)})
		return
	}

	var (
		allLxc []LxcInfo
		errors []ApiError
	)

	for _, obj := range parentObjects {
		portOpen, err := testHostPort(obj.Parent, obj.Port)
		if err != nil {
			errors = append(errors, ApiError{
				Parent:  obj.Parent,
				Action:  "testHostPort",
				Message: err.Error(),
			})
			log.Printf("Failed to check if the port %d for %s is open - %v", obj.Port, obj.Parent, err)
			continue
		}
		if portOpen {
			datacenterNodes, err := getDatacenterNodes(obj.Parent, obj.Port, obj.Token)
			if err != nil {
				errors = append(errors, ApiError{
					Parent:  obj.Parent,
					Action:  "getDatacenterNodes",
					Message: err.Error(),
				})
				log.Printf("Failed to obtain the datacenter nodes for %s - %v", obj.Parent, err)
				continue
			}
			for _, node := range datacenterNodes.Data {
				if node.NodeStatus != "online" {
					errors = append(errors, ApiError{
						Parent:  obj.Parent,
						Node:    node.Node,
						Action:  "onlineStatus",
						Message: fmt.Sprintf("The node %s is offline according to Proxmox", node.Node),
					})
					log.Printf("Skipping node %s - its offline", node.Node)
					continue
				}
				lxcNodeUrl := fmt.Sprintf("https://%s:%d/api2/json/nodes/%v/lxc", obj.Parent, obj.Port, node.Node)
				req, err := http.NewRequest(http.MethodGet, lxcNodeUrl, nil)
				if err != nil {
					errors = append(errors, ApiError{
						Parent:  obj.Parent,
						Node:    node.Node,
						Action:  "createRequest",
						Message: err.Error(),
					})
					log.Printf("Failed to create the LXC per node url for %s - %v", node.Node, err)
					continue
				}

				var temporary LxcEntryObject
				if err := sendRequest(req, &temporary, obj.Token); err != nil {
					errors = append(errors, ApiError{
						Parent:  obj.Parent,
						Node:    node.Node,
						Action:  "fetchLXC",
						Message: err.Error(),
					})
					log.Printf("Failed to process the request for %s - error %v", lxcNodeUrl, err)
					continue
				}

				for _, v := range temporary.Data {
					allLxc = append(allLxc, LxcInfo{
						Parent:      obj.Parent,
						Node:        node.Node,
						NodeStatus:  node.NodeStatus,
						DiskreadMb:  v.Diskread / (1024 * 1024),
						DiskwriteMb: v.Diskwrite / (1024 * 1024),
						MaxMemoryMb: v.MaxMemory / (1024 * 1024),
						MemoryMb:    v.Memory / (1024 * 1024),
						NetinMb:     v.Netin / (1024 * 1024),
						NetoutMb:    v.Netout / (1024 * 1024),
						Name:        v.Name,
						UptimeHours: v.Uptime / (60 * 60),
						Tags:        v.Tags,
						Status:      v.Status,
						Vmid:        v.Vmid,
					})
				}
			}
		}
	}

	if len(allLxc) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "The call executed with success, but no LXC containers were found"})
		return
	}
	c.JSON(http.StatusOK, LxcSummaryResponse{
		Data:   allLxc,
		Errors: errors,
	})
}
