package main

import "time"

type PVEConnectionObject struct {
	Parent string `json:"Parent"`
	Token  string `json:"Token"`
	Port   int    `json:"Port"`
}

type ClusterNodesObject struct {
	Data []ClusterNodeEntry `json:"data"`
}

type ClusterNodeEntry struct {
	Quorum string `json:"quorum_votes"`
	Name   string `json:"name"`
	Ip     string `json:"ring0_addr"`
}

type NodeGuestOverview struct {
	Data []GuestInfo `json:"data"`
}

type GuestInfo struct {
	Parent string `json:"parent"`
	Node   string `json:"node"`
	Name   string `json:"name"`
	Vmid   int    `json:"vmid"`
}

type VmSummaryObject struct {
	Data []VmSummary `json:"data"`
}

type VmSummaryResponse struct {
	Data   []VmSummary `json:"data"`
	Errors []ApiError  `json:"errors"`
}

type VmSummary struct {
	Parent        string `json:"parent"`
	Node          string `json:"node"`
	NodeStatus    string `json:"nodeStatus"`
	Name          string `json:"name"`
	Vmid          int    `json:"vmid"`
	Status        string `json:"status"`
	Cpus          int    `json:"cpus"`
	GuestMemoryGb int    `json:"mem"`
	MaxMemoryGb   int    `json:"maxMem"`
	Uptime        int    `json:"uptime"`
	UptimeHours   int    `json:"uptimeHours"`
}

type NodeDetailsResponse struct {
	Data   []NodeDetails `json:"data"`
	Errors []ApiError    `json:"errors"`
}

type NodeDetails struct {
	NodeInfo struct {
		Parent     string `json:"parent"`
		Node       string `json:"node"`
		NodeStatus string `json:"nodeStatus"`
	} `json:"nodeinfo"`
	Cpuinfo struct {
		Cores   int    `json:"cores"`
		Model   string `json:"model"`
		Mhz     string `json:"mhz"`
		Cpus    int    `json:"cpus"`
		Sockets int    `json:"sockets"`
	} `json:"cpuinfo"`
	Pveversion string `json:"pveversion"`
	BootInfo   struct {
		Mode       string `json:"mode"`
		Secureboot int    `json:"secureboot"`
	} `json:"boot-info"`
	CurrentKernel struct {
		Release string `json:"release"`
		Machine string `json:"machine"`
	} `json:"current-kernel"`
	Dns struct {
		Dns1   string `json:"dns1"`
		Dns2   string `json:"dns2"`
		Dns3   string `json:"dns3"`
		Search string `json:"search"`
	} `json:"dns"`
	Time struct {
		Time     time.Time `json:"time"`
		Timezone string    `json:"timezone"`
	} `json:"time"`
}

type PVENodesObject struct {
	Data []PVENodeInfo `json:"data"`
}

type PVENodeInfo struct {
	Parent      string  `json:"parent"`
	Node        string  `json:"node"`
	NodeStatus  string  `json:"status"`
	MaxCPU      int     `json:"maxCpu"`
	UptimeHours int     `json:"uptime"`
	Mem         int     `json:"mem"`
	MaxMem      int     `json:"maxMem"`
	Cpu         float64 `json:"cpu"`
	Maxdisk     int     `json:"maxDisk"`
	Disk        int     `json:"disk"`
}

type NodeDnsObject struct {
	Data struct {
		Search string `json:"search"`
		Dns1   string `json:"dns1"`
		Dns2   string `json:"dns2"`
		Dns3   string `json:"dns3"`
	} `json:"data"`
}

type NodeTimeObject struct {
	Data struct {
		Time     int    `json:"time"`
		Timezone string `json:"timezone"`
	} `json:"data"`
}

type NodeStatusObject struct {
	Data struct {
		CurrentKernel struct {
			Version string `json:"version"`
			Sysname string `json:"sysname"`
			Release string `json:"release"`
			Machine string `json:"machine"`
		} `json:"current-kernel"`
		Memory struct {
			Free  int `json:"free"`
			Used  int `json:"used"`
			Total int `json:"total"`
		} `json:"memory"`
		BootInfo struct {
			Secureboot int    `json:"secureboot"`
			Mode       string `json:"mode"`
		} `json:"boot-info"`
		Cpuinfo struct {
			Cpus    int    `json:"cpus"`
			Flags   string `json:"flags"`
			Hvm     string `json:"hvm"`
			Mhz     string `json:"mhz"`
			Model   string `json:"model"`
			UserHz  int    `json:"user_hz"`
			Sockets int    `json:"sockets"`
			Cores   int    `json:"cores"`
		} `json:"cpuinfo"`
		Pveversion string `json:"pveversion"`
		Swap       struct {
			Used  int `json:"used"`
			Free  int `json:"free"`
			Total int `json:"total"`
		} `json:"swap"`
		Loadavg []string `json:"loadavg"`
		Ksm     struct {
			Shared int `json:"shared"`
		} `json:"ksm"`
		Wait   float64 `json:"wait"`
		Uptime int     `json:"uptime"`
		Rootfs struct {
			Free  int `json:"free"`
			Used  int `json:"used"`
			Avail int `json:"avail"`
			Total int `json:"total"`
		} `json:"rootfs"`
		Cpu  float64 `json:"cpu"`
		Idle int     `json:"idle"`
	} `json:"data"`
}

type NodeSummaryResponse struct {
	Data   []nodeSummaryWrapper `json:"data"`
	Errors []ApiError           `json:"errors"`
}

type nodeSummaryWrapper struct {
	Parent        string  `json:"parent"`
	Node          string  `json:"node"`
	NodeStatus    string  `json:"nodestatus"`
	MaxCPU        int     `json:"maxCpu"`
	UptimeHours   int     `json:"uptimeHours"`
	MemGb         int     `json:"memGb"`
	MaxMemGb      int     `json:"maxMemGb"`
	Cpu           float64 `json:"cpuLoad"`
	MaxRootDiskGb int     `json:"maxRootDiskGb"`
	RootDiskGb    int     `json:"rootDiskGb"`
}

type QemuGuestWrapper struct {
	Data   QemuGuestInfo `json:"data"`
	Errors []ApiError    `json:"errors"`
}

type QemuGuestInfo struct {
	Status      QemuGuestStatus                    `json:"status"`
	Hostname    QemuHostNameInfo                   `json:"hostname"`
	OSInfo      QemuOSInfo                         `json:"osinfo"`
	NetworkInfo []QemuGuestNetworkInfoObjectResult `json:"network"`
}

type QemuGuestStatus struct {
	Parent         string  `json:"parent"`
	Node           string  `json:"node"`
	Name           string  `json:"name"`
	Status         string  `json:"status"`
	Agent          int     `json:"agent"`
	DiskreadMB     int     `json:"diskreadMB"`
	DiskwriteMB    int     `json:"diskwriteMB"`
	NetoutMB       int     `json:"netoutMB"`
	NetinMB        int     `json:"netinMB"`
	Cpus           int     `json:"cpus"`
	CpuLoad        float64 `json:"cpu"`
	MemoryMB       int     `json:"currentMemMB"`
	MaxMemoryMB    int     `json:"maxMemMB"`
	MachineVersion string  `json:"running-machine"`
}

type QemuCurrentStatusObject struct {
	Data struct {
		Name           string  `json:"name"`
		Status         string  `json:"status"`
		Agent          int     `json:"agent"`
		Diskread       int     `json:"diskread"`
		Diskwrite      int     `json:"diskwrite"`
		Netout         int     `json:"netout"`
		Netin          int     `json:"netin"`
		Cpus           int     `json:"cpus"`
		CpuLoad        float64 `json:"cpu"`
		Memory         int     `json:"mem"`
		Maxmemory      int     `json:"maxmem"`
		MachineVersion string  `json:"running-machine"`
	} `json:"data"`
}

type QemuHostNameObject struct {
	Data struct {
		Result struct {
			Hostname string `json:"host-name"`
		} `json:"result"`
	} `json:"data"`
}

type QemuHostNameInfo struct {
	HostName string `json:"hostname"`
}

type QemuOSInfoObject struct {
	Data struct {
		Result struct {
			Version       string `json:"version-id"`
			Name          string `json:"pretty-name"`
			KernelVersion string `json:"kernel-version"`
			KernelRelease string `json:"kernel-release"`
		} `json:"result"`
	} `json:"data"`
}

type QemuOSInfo struct {
	MajorVersion     string `json:"Version"`
	Name             string `json:"Name"`
	MajorBuildNumber string `json:"KernelVersion"`
	MinorBuildNumber string `json:"KernelRelease"`
}

type QemuGuestNetworkInfoObject struct {
	Data struct {
		Result []QemuGuestNetworkInfoObjectResult `json:"result"`
	} `json:"data"`
}

type QemuGuestNetworkInfoObjectResult struct {
	Name          string `json:"name"`
	IPAddressList []struct {
		IPAddress string `json:"ip-address"`
		Prefix    int    `json:"prefix"`
		Type      string `json:"ip-address-type"`
	} `json:"ip-addresses"`
	HardwareAddress string `json:"hardware-address"`
}

type LxcSummaryResponse struct {
	Data   []LxcInfo  `json:"data"`
	Errors []ApiError `json:"errors"`
}

type LxcInfo struct {
	Parent      string `json:"parent"`
	Node        string `json:"node"`
	NodeStatus  string `json:"nodeStatus"`
	Name        string `json:"name"`
	Vmid        int    `json:"vmid"`
	Status      string `json:"status"`
	Tags        string `json:"tags"`
	UptimeHours int    `json:"uptimeHours"`
	NetoutMb    int    `json:"netOutMb"`
	NetinMb     int    `json:"netInMb"`
	DiskreadMb  int    `json:"diskReadMb"`
	DiskwriteMb int    `json:"diskWriteMb"`
	MemoryMb    int    `json:"memMb"`
	MaxMemoryMb int    `json:"maxMemMb"`
}

type LxcEntryObject struct {
	Data []LxcEntryInfo `json:"data"`
}

type LxcEntryInfo struct {
	Parent    string `json:"parent"`
	Node      string `json:"node"`
	Name      string `json:"name"`
	Vmid      int    `json:"vmid"`
	Status    string `json:"status"`
	Tags      string `json:"tags"`
	Uptime    int    `json:"uptime"`
	Netout    int    `json:"netout"`
	Netin     int    `json:"netin"`
	Diskread  int    `json:"diskread"`
	Diskwrite int    `json:"diskwrite"`
	Memory    int    `json:"mem"`
	MaxMemory int    `json:"maxmem"`
}

type hostStorageList struct {
	Data []struct {
		Storage   string `json:"storage"`
		Active    int    `json:"active"`
		Enabled   int    `json:"enabled"`
		Shared    int    `json:"shared"`
		Type      string `json:"type"`
		Content   string `json:"content"`
		Total     int    `json:"total"`
		Used      int    `json:"used"`
		Available int    `json:"avail"`
	} `json:"data"`
}

type NodeStorageResponse struct {
	Data   []NodeStorageInfo `json:"data"`
	Errors []ApiError        `json:"errors"`
}

type NodeStorageObject struct {
	Data []NodeStorageInfo `json:"data"`
}

type NodeStorageInfo struct {
	Parent      string `json:"parent"`
	Node        string `json:"node"`
	NodeStatus  string `json:"nodeStatus"`
	Storage     string `json:"storage"`
	Active      int    `json:"active"`
	Enabled     int    `json:"enabled"`
	Shared      int    `json:"shared"`
	Type        string `json:"type"`
	Content     string `json:"content"`
	TotalGb     int    `json:"totalGb"`
	UsedGb      int    `json:"usedGb"`
	AvailableGb int    `json:"availableGb"`
}

type NodeDiskObject struct {
	Data   []NodeDiskInfo `json:"data"`
	Errors []ApiError     `json:"errors"`
}

type NodeDiskInfo struct {
	Parent     string `json:"parent"`
	Node       string `json:"node"`
	NodeStatus string `json:"nodeStatus"`
	Vendor     string `json:"vendor"`
	Gpt        int    `json:"gpt"`
	Devpath    string `json:"devpath"`
	Health     string `json:"health"`
	Type       string `json:"type"`
	Wearout    int    `json:"wearout"`
	Serial     string `json:"serial"`
	SizeGb     int    `json:"sizeGb"`
	Model      string `json:"model"`
	Rpm        any    `json:"rpm"`
}

type hostDiskList struct {
	Data []struct {
		Vendor  string `json:"vendor"`
		Gpt     int    `json:"gpt"`
		Devpath string `json:"devpath"`
		Health  string `json:"health"`
		Type    string `json:"type"`
		Wearout any    `json:"wearout"`
		Serial  string `json:"serial"`
		Used    string `json:"used"`
		Model   string `json:"model"`
		Size    int    `json:"size"`
		Rpm     any    `json:"rpm"`
	}
}

type ApiError struct {
	Parent  string `json:"parent"`
	Node    string `json:"node"`
	Action  string `json:"action"`
	Message string `json:"message"`
}
