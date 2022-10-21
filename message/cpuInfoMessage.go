package message

type CPUInfoMessage struct {
	ProcessorNum int64     `json:"processor_num"`
	CPUInfo      []CPUInfo `json:"cpu_info"`
}

type CPUInfo struct {
	Processor int64  `json:"processor"`
	ModelName string `json:"model_name"`
}

type CPUInfoListener func(message CPUInfoMessage)
