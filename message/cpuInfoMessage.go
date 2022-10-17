package message

type CPUInfoMessage struct {
	CPUInfo []CPUInfo `json:"cpu_info"`
}

type CPUInfo struct {
	Processor int64  `json:"processor"`
	ModelName string `json:"modelName"`
}

type CPUInfoListener func(message CPUInfoMessage)
