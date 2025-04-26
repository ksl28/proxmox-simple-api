package main

import (
	"encoding/json"
	"fmt"
	"io"
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

	var allLxc []lxcReturn
	for _, obj := range jsonObjects {
		portOpen, err := testHostPort(obj.Name, obj.Port)
		if err != nil {
			log.Printf("Failed to check if the port %d for %s is open - %v", obj.Port, obj.Name, err)
			return
		}
		if portOpen {
			datacenterNodes, err := getDatacenterNodes(obj.Name, obj.Port, obj.Token)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to obtain the datacenter nodes for %s - %v", obj.Name, err)})
				return
			}
			for _, node := range datacenterNodes.Data {
				if node.Status != "online" {
					log.Printf("Skipping node %s - its either offline or not reachable", node.Node)
					continue
				}
				lxcNodeUrl := fmt.Sprintf("https://%v:%v/api2/json/nodes/%v/lxc", obj.Name, obj.Port, node.Node)
				req, err := http.NewRequest(http.MethodGet, lxcNodeUrl, nil)
				if err != nil {
					log.Printf("Failed to create the LXC per node url for %s - %v", node.Node, err)
					continue
				}
				req.Header.Add("Authorization", obj.Token)
				res, err := httpClient.Do(req)
				if err != nil {
					log.Printf("Failed to perform the HTTP call for %s - %v", node.Node, err)
					continue
				}
				defer res.Body.Close()

				resp, err := io.ReadAll(res.Body)
				if err != nil {
					log.Printf("Failed to read the response body for %s - %v", node.Node, err)
					continue
				}

				var temporary lxcEntryWrapper
				if err = json.Unmarshal(resp, &temporary); err != nil {
					log.Printf("Failed to unmarshal the response body for %s - %v", node.Node, err)
					continue
				}

				for _, v := range temporary.Data {
					var lxcReturn lxcReturn
					lxcReturn.Parent = obj.Name
					lxcReturn.Node = node.Node
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
