package message

// CPUInfoListener is a listener to accept cpu info
type CPUInfoListener func(message CPUInfoMessage)

// CPUPerformanceListener is a listener to accept cpu performance
type CPUPerformanceListener func(message CPUPerformanceMessage)

// MemoryPerformanceListener is a listener to accept cpu performance
type MemoryPerformanceListener func(message MemoryPerformanceMessage)
