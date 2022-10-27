package ssh

import "message"

type Listener struct {
	CPUInfoListener        message.CPUInfoListener
	CPUPerformanceListener message.CPUPerformanceListener
}
