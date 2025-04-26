package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func vmSummary(c *gin.Context) {
	jsonObjects, err := convertJSON()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to convert the JSON data - %v", err)})
		return
	}

	//var glbError []error
	var allVms []VmSummaryStruct
	for _, obj := range jsonObjects {
		v, _ := testHostPort(obj.Name, obj.Port)
		if v {
			datacenterNodes, err := getDatacenterNodes(obj.Name, obj.Port, obj.Token)
			if err != nil {
				log.Printf("Failed to obtain the datacenter nodes for %s - %v", obj.Name, err)
				continue
			}

			// customUrl := fmt.Sprintf("https://%v:%d/api2/json/nodes", obj.Name, obj.Port)
			// req, err := http.NewRequest("GET", customUrl, nil)
			// if err != nil {
			// 	log.Printf("Failed to create the HTTP request in vmSummary - %s - %v", customUrl, err)

			// }
			// req.Header.Add("Authorization", obj.Token)
			// resp, err := httpClient.Do(req)
			// if err != nil {
			// 	c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to conver the JSON data - %v", err)})
			// 	return
			// }
			// body, err := io.ReadAll(resp.Body)
			// if err != nil {
			// 	//glbError = append(glbError, err)
			// 	//c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to decode the response %v", err)})
			// 	continue
			// }
			// defer resp.Body.Close()

			// var nodes nodesStruct
			// if err := json.Unmarshal(body, &nodes); err != nil {
			// 	c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to decode the response %v", err)})
			// 	return
			// }

			for _, singleNode := range datacenterNodes.Data {
				if singleNode.Status != "online" {
					log.Printf("Skipping node %s (offline)", singleNode.Node)
					continue
				}
				singleNodeUrl := fmt.Sprintf("https://%v:%d/api2/json/nodes/%v/qemu", obj.Name, obj.Port, singleNode.Node)
				req, err := http.NewRequest(http.MethodGet, singleNodeUrl, nil)
				if err != nil {
					log.Printf("Failed to create HTTP request for %v - %v", singleNode.Node, err)
					continue
				}
				req.Header.Add("Authorization", obj.Token)
				res, err := httpClient.Do(req)
				if err != nil {
					log.Printf("Failed to perform HTTP request for %v - %v", singleNode.Node, err)
					continue
				}
				defer res.Body.Close()

				body, err := io.ReadAll(res.Body)
				if err != nil {
					log.Printf("Failed to read HTTP response body for %v - %v", singleNode.Node, err)
					continue
				}

				var nodeVms VmSummaryRootStruct
				if err := json.Unmarshal(body, &nodeVms); err != nil {
					log.Printf("Failed to unmarshal JSON for %v - %v", singleNode.Node, err)
					continue
				}

				for i, _ := range nodeVms.Data {
					nodeVms.Data[i].Parent = obj.Name
					nodeVms.Data[i].Node = singleNode.Node
					nodeVms.Data[i].MaxMemoryGb = nodeVms.Data[i].MaxMemoryGb / 1024 / 1024
					nodeVms.Data[i].GuestMemoryGb = nodeVms.Data[i].GuestMemoryGb / 1024 / 1024

				}

				allVms = append(allVms, nodeVms.Data...)
			}
		}
	}
	c.JSON(http.StatusOK, allVms)
}

func vmDetailedOverview(c *gin.Context) {
	var result QemuGuestWrapper

	qemuId := c.Param("id")
	qemuParent := c.Param("parent")
	if qemuId == "" || qemuParent == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "VM ID not found"})
		return
	}
	jsonObjects, err := convertJSON()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to convert the JSON data - %v", err)})
		return
	}
	for _, jsonObject := range jsonObjects {
		if jsonObject.Name == qemuParent {
			v, err := testHostPort(jsonObject.Name, jsonObject.Port)
			if err != nil {
				log.Printf("Failed to check if %s is connectable - %v", jsonObject.Name, err)
				continue
			}
			if v {
				datacenterNodes, err := getDatacenterNodes(jsonObject.Name, jsonObject.Port, jsonObject.Token)
				if err != nil {
					log.Printf("Failed to get cluster nodes for %s - %v", jsonObject.Name, err)
					return
				}
				var vmObj GuestInfo
				for _, datacenterNode := range datacenterNodes.Data {
					if datacenterNode.Status != "online" {
						log.Printf("Skipping node %s (offline)", datacenterNode.Node)
						continue
					}
					guestsResult, err := nodeGuestsOverview("qemu", jsonObject.Name, jsonObject.Port, datacenterNode.Node, jsonObject.Token)
					if err != nil {
						log.Printf("Failed to get the guests for %v - %v", jsonObject.Name, err)
						continue
					} else {
						for _, guest := range guestsResult.Data {
							if strconv.Itoa(guest.Vmid) == qemuId {
								vmObj = guest
							}
						}
					}
				}

				if vmObj.Vmid != 0 {
					var qemuCombined QemuGuestCombined
					qemuStatus, err := qemuCurrentStatus(vmObj.Vmid, vmObj.Parent, jsonObject.Port, vmObj.Node, jsonObject.Token)
					if err != nil {
						log.Printf("Failed to obtain the status for the qemu %v - %v\n", vmObj.Vmid, err)
					} else {
						qemuCombined.Status = QemuStatusClient{
							Name:           qemuStatus.Data.Name,
							Status:         qemuStatus.Data.Status,
							Agent:          qemuStatus.Data.Agent,
							DiskreadMB:     qemuStatus.Data.Diskread / 1024 / 1024,
							DiskwriteMB:    qemuStatus.Data.Diskwrite / 1024 / 1024,
							NetoutMB:       qemuStatus.Data.Netout / 1024 / 1024,
							NetinMB:        qemuStatus.Data.Netin / 1024 / 1024,
							Cpus:           qemuStatus.Data.Cpus,
							CpuLoad:        qemuStatus.Data.CpuLoad,
							MemoryMB:       qemuStatus.Data.Memory / 1024 / 1024,
							MaxMemoryMB:    qemuStatus.Data.Maxmemory / 1024 / 1024,
							MachineVersion: qemuStatus.Data.MachineVersion,
						}
					}

					qemuHostName, err := qemuGuestHostName(vmObj.Vmid, vmObj.Parent, jsonObject.Port, vmObj.Node, jsonObject.Token)
					if err != nil {
						log.Printf("Failed to obtain the hostname for the qemu %v - %v\n", vmObj.Vmid, err)
					} else {
						qemuCombined.Hostname = qemuHostNameClient{
							HostName: qemuHostName.Data.Result.Hostname,
						}
					}

					qemuOsInfo, err := qemuGuestOsInfo(vmObj.Vmid, vmObj.Parent, jsonObject.Port, vmObj.Node, jsonObject.Token)
					if err != nil {
						log.Printf("Failed to obtain the OS info for %v - %v\n", vmObj.Vmid, err)
					} else {
						qemuCombined.OSInfo = QemuOsInfoClient{
							MajorVersion:     qemuOsInfo.Data.Result.Version,
							Name:             qemuOsInfo.Data.Result.Name,
							MajorBuildNumber: qemuOsInfo.Data.Result.KernelVersion,
							MinorBuildNumber: qemuOsInfo.Data.Result.KernelRelease,
						}
					}

					qemuIpInfo, err := qemuGuestIpInfo(vmObj.Vmid, vmObj.Parent, jsonObject.Port, vmObj.Node, jsonObject.Token)
					if err != nil {
						log.Printf("Failed to obtain the IP info for %v - %v\n", vmObj.Vmid, err)
					} else {
						qemuCombined.NetworkInfo = qemuIpInfo.Data.Result
					}

					result.Data = qemuCombined
				}
			}
		}
	}

	c.JSON(http.StatusOK, result)
}

func qemuCurrentStatus(qemuId int, parent string, port int, node string, apiToken string) (qemuCurrentStatusStruct, error) {
	url := fmt.Sprintf("https://%v:%v/api2/json/nodes/%v/qemu/%v/status/current", parent, port, node, qemuId)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("Failed to create HTTP request for qemuCurrentStatus - %v", err)
		return qemuCurrentStatusStruct{}, err
	}

	req.Header.Add("Authorization", apiToken)
	res, err := httpClient.Do(req)
	if err != nil {
		return qemuCurrentStatusStruct{}, err
	}
	defer res.Body.Close()

	response, err := io.ReadAll(res.Body)
	if err != nil {
		return qemuCurrentStatusStruct{}, err
	}

	var qemuStatus qemuCurrentStatusStruct
	if err := json.Unmarshal(response, &qemuStatus); err != nil {
		return qemuCurrentStatusStruct{}, err
	}

	return qemuStatus, nil

}

func qemuGuestHostName(qemuId int, parent string, port int, node string, apiToken string) (qemuHostNameStruct, error) {
	url := fmt.Sprintf("https://%v:%v/api2/json/nodes/%v/qemu/%v/agent/get-host-name", parent, port, node, qemuId)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("Failed to create HTTP request for qemuGuestHostName - %s - %v", url, err)
		return qemuHostNameStruct{}, err
	}

	req.Header.Add("Authorization", apiToken)
	res, err := httpClient.Do(req)
	if err != nil {
		log.Printf("Failed to perform the HTTP request for qemuGuestHostName - %s - %v", url, err)
		return qemuHostNameStruct{}, err
	}
	defer res.Body.Close()

	response, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Failed to read the HTTP response for qemuGuestHostName - %s - %v", url, err)
		return qemuHostNameStruct{}, err
	}

	var qemuHostName qemuHostNameStruct
	if err := json.Unmarshal(response, &qemuHostName); err != nil {
		log.Printf("Failed to unmarshal the JSON data for qemuGuestHostName - %s - %v", url, err)
		return qemuHostNameStruct{}, err
	}
	return qemuHostName, nil
}

func qemuGuestOsInfo(qemuId int, parent string, port int, node string, apiToken string) (qemuOSInfoStruct, error) {
	url := fmt.Sprintf("https://%v:%v/api2/json/nodes/%v/qemu/%v/agent/get-osinfo", parent, port, node, qemuId)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("Failed to create HTTP request for qemuGuestOsInfo - %s - %v", url, err)
		return qemuOSInfoStruct{}, err
	}

	req.Header.Add("Authorization", apiToken)
	res, err := httpClient.Do(req)
	if err != nil {
		log.Printf("Failed to perform the HTTP request for qemuGuestOsInfo - %s - %v", url, err)
		return qemuOSInfoStruct{}, err
	}
	defer res.Body.Close()

	response, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Failed to read the HTTP respeons for qemuGuestOsInfo - %s - %v", url, err)
		return qemuOSInfoStruct{}, err
	}

	var qemuOsInfo qemuOSInfoStruct
	if err := json.Unmarshal(response, &qemuOsInfo); err != nil {
		log.Printf("Failed to unmarshal the JSON data for qemuGuestOsInfo - %s - %v", url, err)
		return qemuOSInfoStruct{}, err
	}

	return qemuOsInfo, nil

}

func qemuGuestIpInfo(qemuId int, parent string, port int, node string, apiToken string) (qemuGuestNetworkInfoStruct, error) {
	url := fmt.Sprintf("https://%v:%v/api2/json/nodes/%v/qemu/%v/agent/network-get-interfaces", parent, port, node, qemuId)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Printf("Failed to create HTTP request for qemuGuestIpInfo - %s - %v", url, err)
		return qemuGuestNetworkInfoStruct{}, err
	}

	req.Header.Add("Authorization", apiToken)
	res, err := httpClient.Do(req)
	if err != nil {
		log.Printf("Failed to perform the HTTP request for qemuGuestIpInfo - %s - %v", url, err)
		return qemuGuestNetworkInfoStruct{}, err
	}
	defer res.Body.Close()

	response, err := io.ReadAll(res.Body)
	if err != nil {
		log.Printf("Failed to read the HTTP response for qemuGuestIpInfo - %s - %v", url, err)
		return qemuGuestNetworkInfoStruct{}, err
	}

	var qemuIpInfo qemuGuestNetworkInfoStruct
	if err := json.Unmarshal(response, &qemuIpInfo); err != nil {
		log.Printf("Failed to unmarshal the JSON data for qemuGuestIpInfo - %s - %v", url, err)
		return qemuGuestNetworkInfoStruct{}, err
	}

	return qemuIpInfo, nil

}
