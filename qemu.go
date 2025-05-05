package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func vmSummary(c *gin.Context) {
	selectedObjs, err := convertJSON()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to convert the JSON data - %v", err)})
		return
	}

	var allVms []VmSummary
	for _, obj := range selectedObjs {
		v, err := testHostPort(obj.Parent, obj.Port)
		if err != nil {
			log.Printf("Failed to check if the port %d for %s is open - %v", obj.Port, obj.Parent, err)
			continue
		}
		if v {
			datacenterNodes, err := getDatacenterNodes(obj.Parent, obj.Port, obj.Token)
			if err != nil {
				log.Printf("Failed to obtain the datacenter nodes for %s - %v", obj.Parent, err)
				continue
			}

			for _, singleNode := range datacenterNodes.Data {
				if singleNode.NodeStatus != "online" {
					log.Printf("Skipping node %s (offline)", singleNode.Node)
					continue
				}
				customUrl := fmt.Sprintf("https://%s:%d/api2/json/nodes/%v/qemu", obj.Parent, obj.Port, singleNode.Node)
				req, err := http.NewRequest(http.MethodGet, customUrl, nil)
				if err != nil {
					log.Printf("Failed to create HTTP request for %v - %v", singleNode.Node, err)
					continue
				}

				var nodeVms VmSummaryObject
				if err := sendRequest(req, &nodeVms, obj.Token); err != nil {
					log.Printf("Failed to process the request for %s - error %v", customUrl, err)
					continue
				}

				for i := range nodeVms.Data {
					nodeVms.Data[i].Parent = obj.Parent
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
	parentName := c.Param("parent")
	if qemuId == "" || parentName == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "VM ID not found"})
		return
	}
	selectedObjs, err := convertJSON()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to convert the JSON data - %v", err)})
		return
	}
	for _, selectedObj := range selectedObjs {
		if selectedObj.Parent == parentName {
			v, err := testHostPort(selectedObj.Parent, selectedObj.Port)
			if err != nil {
				log.Printf("Failed to check if %s is connectable - %v", selectedObj.Parent, err)
				continue
			}
			if v {
				datacenterNodes, err := getDatacenterNodes(selectedObj.Parent, selectedObj.Port, selectedObj.Token)
				if err != nil {
					log.Printf("Failed to get cluster nodes for %s - %v", selectedObj.Parent, err)
					return
				}
				var vmObj GuestInfo
				for _, datacenterNode := range datacenterNodes.Data {
					if datacenterNode.NodeStatus != "online" {
						log.Printf("Skipping node %s (offline)", datacenterNode.Node)
						continue
					}
					guestsResult, err := nodeGuestsOverview("qemu", selectedObj.Parent, selectedObj.Port, datacenterNode.Node, selectedObj.Token)
					if err != nil {
						log.Printf("Failed to get the guests for %v - %v", selectedObj.Parent, err)
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
					var qemuCombined QemuGuestInfo
					qemuStatus, err := qemuCurrentStatus(vmObj.Vmid, vmObj.Parent, selectedObj.Port, vmObj.Node, selectedObj.Token)
					if err != nil {
						log.Printf("Failed to obtain the status for the qemu %v - %v\n", vmObj.Vmid, err)
					} else {
						qemuCombined.Status = QemuGuestStatus{
							Parent:         parentName,
							Name:           qemuStatus.Data.Name,
							NodeStatus:     qemuStatus.Data.Status,
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

					qemuHostName, err := qemuGuestHostName(vmObj.Vmid, vmObj.Parent, selectedObj.Port, vmObj.Node, selectedObj.Token)
					if err != nil {
						log.Printf("Failed to obtain the hostname for the qemu %v - %v\n", vmObj.Vmid, err)
					} else {
						qemuCombined.Hostname = QemuHostNameInfo{
							HostName: qemuHostName.Data.Result.Hostname,
						}
					}

					qemuOsInfo, err := qemuGuestOsInfo(vmObj.Vmid, vmObj.Parent, selectedObj.Port, vmObj.Node, selectedObj.Token)
					if err != nil {
						log.Printf("Failed to obtain the OS info for %v - %v\n", vmObj.Vmid, err)
					} else {
						qemuCombined.OSInfo = QemuOSInfo{
							MajorVersion:     qemuOsInfo.Data.Result.Version,
							Name:             qemuOsInfo.Data.Result.Name,
							MajorBuildNumber: qemuOsInfo.Data.Result.KernelVersion,
							MinorBuildNumber: qemuOsInfo.Data.Result.KernelRelease,
						}
					}

					qemuIpInfo, err := qemuGuestIpInfo(vmObj.Vmid, vmObj.Parent, selectedObj.Port, vmObj.Node, selectedObj.Token)
					if err != nil {
						log.Printf("Failed to obtain the IP info for %v - %v\n", vmObj.Vmid, err)
					} else {
						qemuCombined.NetworkInfo = qemuIpInfo
					}

					result.Data = qemuCombined
				}
			}
		}
	}

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
