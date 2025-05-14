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
				log.Printf("Failed to obtain the datacenter nodes for %s - %v", host.Parent, err)
				continue
			}
			for _, node := range parentNodes.Data {
				if node.NodeStatus != "online" {
					errors = append(errors, ApiError{
						Parent:  host.Parent,
						Node:    node.Node,
						Action:  "onlineStatus",
						Message: fmt.Sprintf("The node %s is offline according to Proxmox", node.Node),
					})
					log.Printf("Skipping node %s - its offline", node.Node)
					continue
				}
				lxcNodeUrl := fmt.Sprintf("https://%s:%d/api2/json/nodes/%v/lxc", host.Parent, host.Port, node.Node)
				req, err := http.NewRequest(http.MethodGet, lxcNodeUrl, nil)
				if err != nil {
					errors = append(errors, ApiError{
						Parent:  host.Parent,
						Node:    node.Node,
						Action:  "createRequest",
						Message: err.Error(),
					})
					log.Printf("Failed to create the LXC per node url for %s - %v", node.Node, err)
					continue
				}

				var details LxcEntryObject
				if err := sendRequest(req, &details, host.Token); err != nil {
					errors = append(errors, ApiError{
						Parent:  host.Parent,
						Node:    node.Node,
						Action:  "fetchLXC",
						Message: err.Error(),
					})
					log.Printf("Failed to process the request for %s - error %v", lxcNodeUrl, err)
					continue
				}

				for _, entry := range details.Data {
					allLxc = append(allLxc, LxcInfo{
						Parent:      host.Parent,
						Node:        node.Node,
						NodeStatus:  node.NodeStatus,
						DiskreadMb:  entry.Diskread / (1024 * 1024),
						DiskwriteMb: entry.Diskwrite / (1024 * 1024),
						MaxMemoryMb: entry.MaxMemory / (1024 * 1024),
						MemoryMb:    entry.Memory / (1024 * 1024),
						NetinMb:     entry.Netin / (1024 * 1024),
						NetoutMb:    entry.Netout / (1024 * 1024),
						Name:        entry.Name,
						UptimeHours: entry.Uptime / (60 * 60),
						Tags:        entry.Tags,
						Status:      entry.Status,
						Vmid:        entry.Vmid,
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
