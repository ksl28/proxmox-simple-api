package main

import "time"

type PVEObject struct {
	Type  string `json:"Type"`
	Name  string `json:"Name"`
	Token string `json:"Token"`
	Port  int    `json:"Port"`
}

type ClusterNodesObject struct {
	Data []ClusterNodesStruct `json:"data"`
}

type ClusterNodesStruct struct {
	Quorum string `json:"quorum_votes"`
	Name   string `json:"name"`
	Ip     string `json:"ring0_addr"`
}

type ClusterNodes struct {
	Data []struct {
		Quorum string `json:"quorum_votes,omitempty"`
		Name   string `json:"name,omitempty"`
		Ip     string `json:"ring0_addr,omitempty"`
	} `json:"data,omitempty"`
}

type NodeQemuId struct {
	Data []struct {
		Name string `json:"name,omitempty"`
		Vmid string `json:"vmid,omitempty"`
	} `json:"data,omitempty"`
}

type NodeGuestId struct {
	Data []struct {
		Parent string `json:"parent,omitempty"`
		Node   string `json:"node,omitempty"`
		Name   string `json:"name,omitempty"`
		Vmid   int    `json:"vmid,omitempty"`
	} `json:"data,omitempty"`
}

type GuestInfo struct {
	Parent string `json:"parent,omitempty"`
	Node   string `json:"node,omitempty"`
	Name   string `json:"name,omitempty"`
	Vmid   int    `json:"vmid,omitempty"`
}

type VmSummaryRootStruct struct {
	Data []VmSummaryStruct `json:"data"`
}

type VmSummaryStruct struct {
	Parent        string `json:"parent"`
	Node          string `json:"node"`
	Name          string `json:"name"`
	Vmid          int    `json:"vmid"`
	Status        string `json:"status"`
	Cpus          int    `json:"cpus"`
	GuestMemoryGb int    `json:"mem"`
	MaxMemoryGb   int    `json:"maxmem"`
	Uptime        int    `json:"uptime"`
}

type DataObject struct {
	Data DetailedHostStruct `json:"data"`
}

type DetailedHostStruct struct {
	NodeInfo struct {
		Parent string `json:"parent,omitempty"`
		Node   string `json:"node,omitempty"`
	} `json:"nodeinfo,omitempty"`
	Cpuinfo struct {
		Cores   int    `json:"cores,omitempty"`
		Model   string `json:"model,omitempty"`
		Mhz     string `json:"mhz,omitempty"`
		Cpus    int    `json:"cpus,omitempty"`
		Sockets int    `json:"sockets,omitempty"`
	} `json:"cpuinfo,omitempty"`
	Pveversion string `json:"pveversion,omitempty"`
	BootInfo   struct {
		Mode       string `json:"mode,omitempty"`
		Secureboot int    `json:"secureboot,omitempty"`
	} `json:"boot-info,omitempty"`
	CurrentKernel struct {
		Release string `json:"release,omitempty"`
		Machine string `json:"machine,omitempty"`
	} `json:"current-kernel,omitempty"`
	Dns struct {
		Dns1   string `json:"dns1,omitempty"`
		Dns2   string `json:"dns2,omitempty"`
		Dns3   string `json:"dns3,omitempty"`
		Search string `json:"search,omitempty"`
	} `json:"dns,omitempty"`
	Time struct {
		Time     time.Time `json:"time,omitempty"`
		Timezone string    `json:"timezone,omitempty"`
	} `json:"time,omitempty"`
}

type DataArray struct {
	Data []PVENodesResponse `json:"data"`
}

type PVENodesResponse struct {
	Parent      string  `json:"Parent,omitempty"`
	Node        string  `json:"node,omitempty"`
	Status      string  `json:"status,omitempty"`
	MaxCPU      int     `json:"maxcpu,omitempty"`
	UptimeHours int     `json:"uptime,omitempty"`
	Mem         int     `json:"mem,omitempty"`
	MaxMem      int     `json:"maxmem,omitempty"`
	Cpu         float64 `json:"cpu,omitempty"`
	Maxdisk     int     `json:"maxdisk,omitempty"`
	Disk        int     `json:"disk,omitempty"`
}

type nodesStruct struct {
	Data []struct {
		Parent      string  `json:"Parent,omitempty"`
		Node        string  `json:"node,omitempty"`
		Status      string  `json:"status,omitempty"`
		MaxCPU      int     `json:"maxcpu,omitempty"`
		UptimeHours int     `json:"uptime,omitempty"`
		Mem         int     `json:"mem,omitempty"`
		MaxMem      int     `json:"maxmem,omitempty"`
		Cpu         float64 `json:"cpu,omitempty"`
		Maxdisk     int     `json:"maxdisk,omitempty"`
		Disk        int     `json:"disk,omitempty"`
	} `json:"data,omitempty"`
}

type nodeDns struct {
	Data struct {
		Search string `json:"search,omitempty"`
		Dns1   string `json:"dns1,omitempty"`
		Dns2   string `json:"dns2,omitempty"`
		Dns3   string `json:"dns3,omitempty"`
	} `json:"data,omitempty"`
}

type nodeTime struct {
	Data struct {
		Time     int    `json:"time,omitempty"`
		Timezone string `json:"timezone,omitempty"`
	} `json:"data,omitempty"`
}

type nodeStatus struct {
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
	Parent        string  `json:"Parent,omitempty"`
	Node          string  `json:"node,omitempty"`
	Status        string  `json:"status,omitempty"`
	MaxCPU        int     `json:"maxcpu,omitempty"`
	UptimeHours   int     `json:"uptimeHours,omitempty"`
	MemGb         int     `json:"memGb,omitempty"`
	MaxMemGb      int     `json:"maxmemGb,omitempty"`
	Cpu           float64 `json:"cpuLoad,omitempty"`
	MaxRootDiskGb int     `json:"maxdiskGb,omitempty"`
	RootDiskGb    int     `json:"diskGb,omitempty"`
}

type QemuGuestWrapper struct {
	Data QemuGuestCombined `json:"data"`
}

type QemuGuestCombined struct {
	Status      any `json:"status"`
	Hostname    any `json:"hostname"`
	OSInfo      any `json:"osinfo"`
	NetworkInfo any `json:"network"`
}

type QemuStatusClient struct {
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

type qemuCurrentStatusStruct struct {
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

type qemuHostNameStruct struct {
	Data struct {
		Result struct {
			Hostname string `json:"host-name"`
		} `json:"result,omitempty"`
	} `json:"data"`
}

type qemuHostNameClient struct {
	HostName string `json:"hostname"`
}

type qemuOSInfoStruct struct {
	Data struct {
		Result struct {
			Version       int     `json:"version-id,omitempty,string"`
			Name          string  `json:"pretty-name,omitempty"`
			KernelVersion float64 `json:"kernel-version,omitempty,string"`
			KernelRelease int     `json:"kernel-release,omitempty,string"`
		} `json:"result,omitempty"`
	} `json:"data"`
}

type QemuOsInfoClient struct {
	MajorVersion     int     `json:"Version"`
	Name             string  `json:"Name"`
	MajorBuildNumber float64 `json:"KernelVersion"`
	MinorBuildNumber int     `json:"KernelRelease"`
}

type qemuGuestNetworkInfoStruct struct {
	Data struct {
		Result []struct {
			Name          string `json:"name"`
			IPAddressList []struct {
				IPAddress string `json:"ip-address"`
				Prefix    int    `json:"prefix"`
				Type      string `json:"ip-address-type"`
			} `json:"ip-addresses"`
			HardwareAddress string `json:"hardware-address"`
		} `json:"result"`
	} `json:"data"`
}

type lxcReturn struct {
	Parent      string `json:"parent"`
	Node        string `json:"node"`
	Name        string `json:"name"`
	Vmid        int    `json:"vmid"`
	Status      string `json:"status"`
	Tags        string `json:"tags"`
	UptimeHours int    `json:"uptimehours"`
	NetoutMb    int    `json:"netoutmb"`
	NetinMb     int    `json:"netinmb"`
	DiskreadMb  int    `json:"diskreadmb"`
	DiskwriteMb int    `json:"diskwritemb"`
	MemoryMb    int    `json:"memmb"`
	MaxMemoryMb int    `json:"maxmemmb"`
}

type lxcEntryWrapper struct {
	Data []lxcEntry `json:"data"`
}
type lxcEntry struct {
	Parent    string `json:"parent"`
	Node      string `json:"node"`
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

type hostStorageListWrapper struct {
	Data []hostStorageEntry `json:"data"`
}

type hostStorageEntry struct {
	Parent      string `json:"parent,omitempty"`
	Node        string `json:"node,omitempty"`
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

type hostDiskEntryWrapper struct {
	Data []hostDiskEntry `json:"data,omitempty"`
}

type hostDiskEntry struct {
	Parent  string `json:"parent,omitempty"`
	Node    string `json:"node,omitempty"`
	Vendor  string `json:"vendor,omitempty"`
	Gpt     int    `json:"gpt,omitempty"`
	Devpath string `json:"devpath,omitempty"`
	Health  string `json:"health,omitempty"`
	Type    string `json:"type,omitempty"`
	Wearout int    `json:"wearout,omitempty"`
	Serial  string `json:"serial,omitempty"`
	UsedGb  int    `json:"used,omitempty"`
	Model   string `json:"model,omitempty"`
	Size    int    `json:"size,omitempty"`
	Rpm     any    `json:"rpm,omitempty"`
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
