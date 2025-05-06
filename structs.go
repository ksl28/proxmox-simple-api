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
	Data []GuestInfo `json:"data,omitempty"`
}

type GuestInfo struct {
	Parent string `json:"parent,omitempty"`
	Node   string `json:"node,omitempty"`
	Name   string `json:"name,omitempty"`
	Vmid   int    `json:"vmid,omitempty"`
}

type VmSummaryObject struct {
	Data []VmSummary `json:"data"`
}

type VmSummary struct {
	Parent        string `json:"parent"`
	Node          string `json:"node"`
	Name          string `json:"name"`
	Vmid          int    `json:"vmid"`
	Status        string `json:"status"`
	Cpus          int    `json:"cpus"`
	GuestMemoryGb int    `json:"mem"`
	MaxMemoryGb   int    `json:"maxMem"`
	Uptime        int    `json:"uptime"`
}

type NodeDetailsObject struct {
	Data NodeDetails `json:"data"`
}

type NodeDetails struct {
	NodeInfo struct {
		Parent     string `json:"parent,omitempty"`
		Node       string `json:"node,omitempty"`
		NodeStatus string `json:"nodeStatus,omitempty"`
	} `json:"nodeinfo"`
	Cpuinfo struct {
		Cores   int    `json:"cores,omitempty"`
		Model   string `json:"model,omitempty"`
		Mhz     string `json:"mhz,omitempty"`
		Cpus    int    `json:"cpus,omitempty"`
		Sockets int    `json:"sockets,omitempty"`
	} `json:"cpuinfo"`
	Pveversion string `json:"pveversion,omitempty"`
	BootInfo   struct {
		Mode       string `json:"mode,omitempty"`
		Secureboot int    `json:"secureboot,omitempty"`
	} `json:"boot-info"`
	CurrentKernel struct {
		Release string `json:"release,omitempty"`
		Machine string `json:"machine,omitempty"`
	} `json:"current-kernel"`
	Dns struct {
		Dns1   string `json:"dns1,omitempty"`
		Dns2   string `json:"dns2,omitempty"`
		Dns3   string `json:"dns3,omitempty"`
		Search string `json:"search,omitempty"`
	} `json:"dns"`
	Time struct {
		Time     time.Time `json:"time,omitempty"`
		Timezone string    `json:"timezone,omitempty"`
	} `json:"time"`
}

type PVENodesObject struct {
	Data []PVENodeInfo `json:"data"`
}

type PVENodeInfo struct {
	Parent      string  `json:"parent,omitempty"`
	Node        string  `json:"node,omitempty"`
	NodeStatus  string  `json:"status,omitempty"`
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
		Search string `json:"search,omitempty"`
		Dns1   string `json:"dns1,omitempty"`
		Dns2   string `json:"dns2,omitempty"`
		Dns3   string `json:"dns3,omitempty"`
	} `json:"data"`
}

type NodeTimeObject struct {
	Data struct {
		Time     int    `json:"time,omitempty"`
		Timezone string `json:"timezone,omitempty"`
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

type nodeSummaryWrapper struct {
	Parent        string  `json:"parent,omitempty"`
	Node          string  `json:"node,omitempty"`
	NodeStatus    string  `json:"nodestatus,omitempty"`
	MaxCPU        int     `json:"maxCpu,omitempty"`
	UptimeHours   int     `json:"uptimeHours,omitempty"`
	MemGb         int     `json:"memGb,omitempty"`
	MaxMemGb      int     `json:"maxMemGb,omitempty"`
	Cpu           float64 `json:"cpuLoad,omitempty"`
	MaxRootDiskGb int     `json:"maxRootDiskGb,omitempty"`
	RootDiskGb    int     `json:"rootDiskGb,omitempty"`
}

type QemuGuestWrapper struct {
	Data QemuGuestInfo `json:"data"`
}

type QemuGuestInfo struct {
	Status      QemuGuestStatus                    `json:"status"`
	Hostname    QemuHostNameInfo                   `json:"hostname"`
	OSInfo      QemuOSInfo                         `json:"osinfo"`
	NetworkInfo []QemuGuestNetworkInfoObjectResult `json:"network"`
}

type QemuGuestStatus struct {
	Parent         string  `json:"parent,omitempty"`
	Name           string  `json:"name,omitempty"`
	NodeStatus     string  `json:"nodestatus,omitempty"`
	Agent          int     `json:"agent,omitempty"`
	DiskreadMB     int     `json:"diskreadMB,omitempty"`
	DiskwriteMB    int     `json:"diskwriteMB,omitempty"`
	NetoutMB       int     `json:"netoutMB,omitempty"`
	NetinMB        int     `json:"netinMB,omitempty"`
	Cpus           int     `json:"cpus,omitempty"`
	CpuLoad        float64 `json:"cpu,omitempty"`
	MemoryMB       int     `json:"currentMemMB,omitempty"`
	MaxMemoryMB    int     `json:"maxMemMB,omitempty"`
	MachineVersion string  `json:"running-machine,omitempty"`
}

type QemuCurrentStatusObject struct {
	Data struct {
		Name           string  `json:"name,omitempty"`
		Status         string  `json:"status,omitempty"`
		Agent          int     `json:"agent,omitempty"`
		Diskread       int     `json:"diskread,omitempty"`
		Diskwrite      int     `json:"diskwrite,omitempty"`
		Netout         int     `json:"netout,omitempty"`
		Netin          int     `json:"netin,omitempty"`
		Cpus           int     `json:"cpus,omitempty"`
		CpuLoad        float64 `json:"cpu,omitempty"`
		Memory         int     `json:"mem,omitempty"`
		Maxmemory      int     `json:"maxmem,omitempty"`
		MachineVersion string  `json:"running-machine,omitempty"`
	} `json:"data"`
}

type QemuHostNameObject struct {
	Data struct {
		Result struct {
			Hostname string `json:"host-name,omitempty"`
		} `json:"result"`
	} `json:"data"`
}

type QemuHostNameInfo struct {
	HostName string `json:"hostname"`
}

type QemuOSInfoObject struct {
	Data struct {
		Result struct {
			Version       string `json:"version-id,omitempty"`
			Name          string `json:"pretty-name,omitempty"`
			KernelVersion string `json:"kernel-version,omitempty"`
			KernelRelease string `json:"kernel-release,omitempty"`
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
		Result []QemuGuestNetworkInfoObjectResult `json:"result,omitempty"`
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
	Parent    string `json:"parent,omitempty"`
	Node      string `json:"node,omitempty"`
	Name      string `json:"name,omitempty"`
	Vmid      int    `json:"vmid,omitempty"`
	Status    string `json:"status,omitempty"`
	Tags      string `json:"tags,omitempty"`
	Uptime    int    `json:"uptime,omitempty"`
	Netout    int    `json:"netout,omitempty"`
	Netin     int    `json:"netin,omitempty"`
	Diskread  int    `json:"diskread,omitempty"`
	Diskwrite int    `json:"diskwrite,omitempty"`
	Memory    int    `json:"mem,omitempty"`
	MaxMemory int    `json:"maxmem,omitempty"`
}

type hostStorageList struct {
	Data []struct {
		Storage   string `json:"storage,omitempty"`
		Active    int    `json:"active,omitempty"`
		Enabled   int    `json:"enabled,omitempty"`
		Shared    int    `json:"shared,omitempty"`
		Type      string `json:"type,omitempty"`
		Content   string `json:"content,omitempty"`
		Total     int    `json:"total,omitempty"`
		Used      int    `json:"used,omitempty"`
		Available int    `json:"avail,omitempty"`
	} `json:"data,omitempty"`
}

type NodeStorageObject struct {
	Data []NodeStorageInfo `json:"data"`
}

type NodeStorageInfo struct {
	Parent      string `json:"parent,omitempty"`
	Node        string `json:"node,omitempty"`
	NodeStatus  string `json:"nodeStatus,omitempty"`
	Storage     string `json:"storage,omitempty"`
	Active      int    `json:"active,omitempty"`
	Enabled     int    `json:"enabled,omitempty"`
	Shared      int    `json:"shared,omitempty"`
	Type        string `json:"type,omitempty"`
	Content     string `json:"content,omitempty"`
	TotalGb     int    `json:"totalGb"`
	UsedGb      int    `json:"usedGb"`
	AvailableGb int    `json:"availableGb"`
}

type NodeDiskObject struct {
	Data []NodeDiskInfo `json:"data,omitempty"`
}

type NodeDiskInfo struct {
	Parent     string `json:"parent,omitempty"`
	Node       string `json:"node,omitempty"`
	NodeStatus string `json:"nodeStatus,omitempty"`
	Vendor     string `json:"vendor"`
	Gpt        int    `json:"gpt"`
	Devpath    string `json:"devpath"`
	Health     string `json:"health"`
	Type       string `json:"type"`
	Wearout    int    `json:"wearout"`
	Serial     string `json:"serial"`
	UsedGb     int    `json:"used"`
	Model      string `json:"model"`
	Size       int    `json:"size"`
	Rpm        any    `json:"rpm"`
}

type hostDiskList struct {
	Data []struct {
		Vendor  string `json:"vendor,omitempty"`
		Gpt     int    `json:"gpt,omitempty"`
		Devpath string `json:"devpath,omitempty"`
		Health  string `json:"health,omitempty"`
		Type    string `json:"type,omitempty"`
		Wearout any    `json:"wearout,omitempty"`
		Serial  string `json:"serial,omitempty"`
		Used    string `json:"used,omitempty"`
		Model   string `json:"model,omitempty"`
		Size    int    `json:"size,omitempty"`
		Rpm     any    `json:"rpm,omitempty"`
	}
}

type ApiError struct {
	Parent  string `json:"parent"`
	Node    string `json:"node,omitempty"`
	Action  string `json:"action"`
	Message string `json:"message"`
}
