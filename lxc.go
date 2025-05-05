package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func lxcSummary(c *gin.Context) {
	jsonObjects, err := convertJSON()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to convert the JSON data - %v", err)})
		return
	}

	var allLxc []LxcInfo
	for _, obj := range jsonObjects {
		portOpen, err := testHostPort(obj.Parent, obj.Port)
		if err != nil {
			log.Printf("Failed to check if the port %d for %s is open - %v", obj.Port, obj.Parent, err)
			return
		}
		if portOpen {
			datacenterNodes, err := getDatacenterNodes(obj.Parent, obj.Port, obj.Token)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to obtain the datacenter nodes for %s - %v", obj.Parent, err)})
				return
			}
			for _, node := range datacenterNodes.Data {
				if node.NodeStatus != "online" {
					var lxcReturn LxcInfo
					lxcReturn.Parent = obj.Parent
					lxcReturn.Node = node.Node
					lxcReturn.NodeStatus = node.NodeStatus
					allLxc = append(allLxc, lxcReturn)
					log.Printf("Skipping node %s - its offline", node.Node)
					continue
				}
				lxcNodeUrl := fmt.Sprintf("https://%s:%d/api2/json/nodes/%v/lxc", obj.Parent, obj.Port, node.Node)
				req, err := http.NewRequest(http.MethodGet, lxcNodeUrl, nil)
				if err != nil {
					log.Printf("Failed to create the LXC per node url for %s - %v", node.Node, err)
					continue
				}

				var temporary LxcEntryObject
				if err := sendRequest(req, &temporary, obj.Token); err != nil {
					log.Printf("Failed to process the request for %s - error %v", lxcNodeUrl, err)
					continue
				}

				for _, v := range temporary.Data {
					var lxcReturn LxcInfo
					lxcReturn.Parent = obj.Parent
					lxcReturn.Node = node.Node
					lxcReturn.NodeStatus = node.NodeStatus
					lxcReturn.DiskreadMb = v.Diskread / (1024 * 1024)
					lxcReturn.DiskwriteMb = v.Diskwrite / (1024 * 1024)
					lxcReturn.MaxMemoryMb = v.MaxMemory / (1024 * 1024)
					lxcReturn.MemoryMb = v.Memory / (1024 * 1024)
					lxcReturn.NetinMb = v.Netin / (1024 * 1024)
					lxcReturn.NetoutMb = v.Netout / (1024 * 1024)
					lxcReturn.Name = v.Name
					lxcReturn.UptimeHours = v.Uptime / (60 * 60)
					lxcReturn.Tags = v.Tags
					lxcReturn.Status = v.Status
					lxcReturn.Vmid = v.Vmid
					allLxc = append(allLxc, lxcReturn)
				}

			}
		}

	}

	if len(allLxc) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "The call executed with success, but  no LXC containers were found"})
		return
	}
	c.JSON(http.StatusOK, allLxc)
}
