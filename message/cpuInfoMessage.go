package message

type CPUInfoMessage struct {
	CPUInfo []CPUInfo
}

type CPUInfo struct {
	ModelName string
}
