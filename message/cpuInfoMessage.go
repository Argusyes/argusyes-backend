package message

type CPUInfoMessage struct {
	CPUInfo []CPUInfo
}

type CPUInfo struct {
	Processor int64
	ModelName string
}

type CPUInfoListener func(message CPUInfoMessage)
