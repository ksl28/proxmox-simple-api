package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

func vmSummary(c *gin.Context) {
	parentObjects, err := convertJSON()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to convert the JSON data - %v", err)})
		return
	}

	var (
		allVms []VmSummary
		errors []ApiError
	)

	ch := make(chan []VmSummary, len(parentObjects))
	errCh := make(chan ApiError, len(parentObjects)*2)
	var wg sync.WaitGroup
	wg.Add(len(parentObjects))
	for _, host := range parentObjects {
		go func(host PVEConnectionObject) {
			defer wg.Done()
			var vmSlice []VmSummary
			portOpen, err := testHostPort(host.Parent, host.Port)
			if err != nil {
				log.Printf("Failed to check if the port %d for %s is open - %v", host.Port, host.Parent, err)
			}
			if portOpen {
				parentNodes, err := getParentNodes(host.Parent, host.Port, host.Token)
				if err != nil {
					errCh <- ApiError{
						Parent:  host.Parent,
						Action:  "getParentNodes",
						Message: err.Error(),
					}
					log.Printf("Failed to obtain the datacenter nodes for %s - %v", host.Parent, err)
				}

				for _, node := range parentNodes.Data {
					if node.NodeStatus != "online" {
						errCh <- ApiError{
							Parent:  host.Parent,
							Node:    node.Node,
							Action:  "onlineStatus",
							Message: fmt.Sprintf("The node %s appears to be offline according to Proxmox", node.Node),
						}
						log.Printf("Skipping node %s (offline)", node.Node)
						continue
					}

					vmNodeUrl := fmt.Sprintf("https://%s:%d/api2/json/nodes/%v/qemu", host.Parent, host.Port, node.Node)
					req, err := http.NewRequest(http.MethodGet, vmNodeUrl, nil)
					if err != nil {
						errCh <- ApiError{
							Parent:  host.Parent,
							Node:    node.Node,
							Action:  "createRequest",
							Message: err.Error(),
						}
						log.Printf("Failed to create HTTP request for %v - %v", node.Node, err)
						continue
					}

					var vmWrapper VmSummaryObject
					if err := sendRequest(req, &vmWrapper, host.Token); err != nil {
						errCh <- ApiError{
							Parent:  host.Parent,
							Node:    node.Node,
							Action:  "sendRequest",
							Message: err.Error(),
						}
						log.Printf("Failed to process the request for %s - error %v", vmNodeUrl, err)
						continue
					}

					for i := range vmWrapper.Data {
						vmWrapper.Data[i].Parent = host.Parent
						vmWrapper.Data[i].Node = node.Node
						vmWrapper.Data[i].NodeStatus = node.NodeStatus
						vmWrapper.Data[i].MaxMemoryGb = vmWrapper.Data[i].MaxMemoryGb / 1024 / 1024
						vmWrapper.Data[i].GuestMemoryGb = vmWrapper.Data[i].GuestMemoryGb / 1024 / 1024
						vmWrapper.Data[i].UptimeHours = vmWrapper.Data[i].Uptime / 3600

					}

					vmSlice = append(vmSlice, vmWrapper.Data...)
					ch <- vmSlice
				}

			}
		}(host)
	}
	for range parentObjects {
		batch := <-ch
		allVms = append(allVms, batch...)
	}
	wg.Wait()
	close(errCh)
	for e := range errCh {
		errors = append(errors, e)
	}

	c.JSON(http.StatusOK, VmSummaryResponse{
		Data:   allVms,
		Errors: errors,
	})

}

func vmDetailedOverview(c *gin.Context) {
	var (
		result QemuGuestWrapper
		errors []ApiError
	)

	qemuId := c.Param("id")
	parentName := c.Param("parent")
	if qemuId == "" || parentName == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "The VM or Parent id is not added to the query."})
		return
	}
	parentObjects, err := convertJSON()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to convert the JSON data - %v", err)})
		return
	}
	for _, host := range parentObjects {
		if host.Parent == parentName {
			portOpen, err := testHostPort(host.Parent, host.Port)
			if err != nil {
				errors = append(errors, ApiError{
					Parent:  host.Parent,
					Action:  "testHostPort",
					Message: err.Error(),
				})
				log.Printf("Failed to check if the port was open on %s:%d - %v", host.Parent, host.Port, err)
				c.JSON(http.StatusBadGateway, QemuGuestWrapper{
					Data:   result.Data,
					Errors: errors,
				})
				return
			}
			if portOpen {
				parentNodes, err := getParentNodes(host.Parent, host.Port, host.Token)
				if err != nil {
					errors = append(errors, ApiError{
						Parent:  host.Parent,
						Action:  "testHostPort",
						Message: err.Error(),
					})
					log.Printf("Failed to obtain the cluster nodes for %s - %v", host.Parent, err)
					c.JSON(http.StatusInternalServerError, QemuGuestWrapper{
						Data:   result.Data,
						Errors: errors,
					})
					return
				}
				var vmObj GuestInfo
				for _, node := range parentNodes.Data {
					if node.NodeStatus != "online" {
						log.Printf("Skipping node %s (offline)", node.Node)
						continue
					}
					guestsResult, err := nodeGuestsOverview("qemu", host.Parent, host.Port, node.Node, host.Token)
					if err != nil {
						errors = append(errors, ApiError{
							Parent:  host.Parent,
							Node:    node.Node,
							Action:  "nodeGuestsOverview",
							Message: fmt.Sprintf("Failed to get the guests for %v - %v", host.Parent, err),
						})
						log.Printf("Failed to get the guests for %v - %v", host.Parent, err)
						continue
					}
					for _, guest := range guestsResult.Data {
						if strconv.Itoa(guest.Vmid) == qemuId {
							vmObj = guest
						}
					}
				}

				if vmObj.Vmid != 0 {
					var qemuCombined QemuGuestInfo
					qemuStatus, err := qemuCurrentStatus(vmObj.Vmid, host.Parent, host.Port, vmObj.Node, host.Token)
					if err == nil {
						qemuCombined.Status = QemuGuestStatus{
							Parent:         parentName,
							Node:           vmObj.Node,
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
					} else {
						log.Printf("Failed to obtain the status for the qemu %v - %v\n", vmObj.Vmid, err)
					}

					if qemuStatus.Data.Agent == 1 && qemuStatus.Data.Status == "running" {
						qemuHostName, err := qemuGuestHostName(vmObj.Vmid, host.Parent, host.Port, vmObj.Node, host.Token)
						if err != nil {
							errors = append(errors, ApiError{
								Parent:  host.Parent,
								Node:    vmObj.Node,
								Action:  "qemuGuestHostName",
								Message: err.Error(),
							})
							log.Printf("Failed to obtain the hostname for qemu %v - %v", vmObj.Vmid, err)
						} else {
							qemuCombined.Hostname = QemuHostNameInfo{
								HostName: qemuHostName.Data.Result.Hostname,
							}
						}

						qemuOsInfo, err := qemuGuestOsInfo(vmObj.Vmid, host.Parent, host.Port, vmObj.Node, host.Token)
						if err != nil {
							errors = append(errors, ApiError{
								Parent:  host.Parent,
								Node:    vmObj.Node,
								Action:  "qemuGuestOsInfo",
								Message: err.Error(),
							})
							log.Printf("Failed to obtain the OS info for qemu %v - %v\n", vmObj.Vmid, err)
						} else {
							qemuCombined.OSInfo = QemuOSInfo{
								MajorVersion:     qemuOsInfo.Data.Result.Version,
								Name:             qemuOsInfo.Data.Result.Name,
								MajorBuildNumber: qemuOsInfo.Data.Result.KernelVersion,
								MinorBuildNumber: qemuOsInfo.Data.Result.KernelRelease,
							}
						}

						qemuIpInfo, err := qemuGuestIpInfo(vmObj.Vmid, host.Parent, host.Port, vmObj.Node, host.Token)
						if err != nil {
							errors = append(errors, ApiError{
								Parent:  host.Parent,
								Node:    vmObj.Node,
								Action:  "qemuGuestIpInfo",
								Message: err.Error(),
							})
							log.Printf("Failed to obtain the IP info for qemu %v - %v", vmObj.Vmid, err)
						} else {
							qemuCombined.NetworkInfo = qemuIpInfo
						}
					} else {
						errors = append(errors, ApiError{
							Parent:  host.Parent,
							Node:    vmObj.Node,
							Action:  "checkAgent",
							Message: fmt.Sprintf("The VM %s is either not powered on or have an agent installed in the Guest OS - skipping guest inventory.", vmObj.Name),
						})
					}

					result.Data = qemuCombined
				} else {
					c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("The Qemu ID %s was not found of %s - no errors were encountered.", qemuId, parentName)})
					return
				}
			} else {
				c.JSON(http.StatusBadGateway, QemuGuestWrapper{
					Data: result.Data,
					Errors: append(errors, ApiError{
						Parent:  host.Parent,
						Action:  "testHostPort",
						Message: fmt.Sprintf("port %d closed", host.Port),
					}),
				})
				return
			}
		}
	}

	result.Errors = errors
	c.JSON(http.StatusOK, result)
}

func qemuCurrentStatus(qemuId int, parent string, port int, node string, apiToken string) (QemuCurrentStatusObject, error) {
	customUrl := fmt.Sprintf("https://%s:%d/api2/json/nodes/%v/qemu/%v/status/current", parent, port, node, qemuId)
	req, err := http.NewRequest(http.MethodGet, customUrl, nil)
	if err != nil {
		log.Printf("Failed to create HTTP request for qemuCurrentStatus - %v", err)
		return QemuCurrentStatusObject{}, err
	}

	var qemuStatus QemuCurrentStatusObject
	if err := sendRequest(req, &qemuStatus, apiToken); err != nil {
		log.Printf("Failed to process the request for %s - error %v", customUrl, err)
		return QemuCurrentStatusObject{}, err
	}

	return qemuStatus, nil

}

func qemuGuestHostName(qemuId int, parent string, port int, node string, apiToken string) (QemuHostNameObject, error) {
	customUrl := fmt.Sprintf("https://%s:%d/api2/json/nodes/%v/qemu/%v/agent/get-host-name", parent, port, node, qemuId)
	req, err := http.NewRequest(http.MethodGet, customUrl, nil)
	if err != nil {
		log.Printf("Failed to create HTTP request for qemuGuestHostName - %s - %v", customUrl, err)
		return QemuHostNameObject{}, err
	}

	var qemuHostName QemuHostNameObject
	if err := sendRequest(req, &qemuHostName, apiToken); err != nil {
		log.Printf("Failed to process the request for %s - error %v", customUrl, err)
		return QemuHostNameObject{}, err
	}
	return qemuHostName, nil
}

func qemuGuestOsInfo(qemuId int, parent string, port int, node string, apiToken string) (QemuOSInfoObject, error) {
	customUrl := fmt.Sprintf("https://%s:%d/api2/json/nodes/%v/qemu/%v/agent/get-osinfo", parent, port, node, qemuId)
	req, err := http.NewRequest(http.MethodGet, customUrl, nil)
	if err != nil {
		log.Printf("Failed to create HTTP request for qemuGuestOsInfo - %s - %v", customUrl, err)
		return QemuOSInfoObject{}, err
	}

	var qemuOsInfo QemuOSInfoObject
	if err := sendRequest(req, &qemuOsInfo, apiToken); err != nil {
		log.Printf("Failed to process the request for %s - error %v", customUrl, err)
		return QemuOSInfoObject{}, err
	}

	return qemuOsInfo, nil

}

func qemuGuestIpInfo(qemuId int, parent string, port int, node string, apiToken string) ([]QemuGuestNetworkInfoObjectResult, error) {
	customUrl := fmt.Sprintf("https://%s:%d/api2/json/nodes/%v/qemu/%v/agent/network-get-interfaces", parent, port, node, qemuId)
	req, err := http.NewRequest(http.MethodGet, customUrl, nil)
	if err != nil {
		log.Printf("Failed to create HTTP request for qemuGuestIpInfo - %s - %v", customUrl, err)
		return []QemuGuestNetworkInfoObjectResult{}, err
	}

	var qemuIpInfo QemuGuestNetworkInfoObject
	if err := sendRequest(req, &qemuIpInfo, apiToken); err != nil {
		log.Printf("Failed to process the request for %s - error %v", customUrl, err)
		return []QemuGuestNetworkInfoObjectResult{}, err
	}

	return qemuIpInfo.Data.Result, nil

}
