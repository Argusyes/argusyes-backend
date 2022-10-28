package message

// CPUInfoMessage is the message to describe CPU info
type CPUInfoMessage struct {
	Port       int               `json:"port"`
	Host       string            `json:"host"`
	User       string            `json:"user"`
	CPUInfoMap map[int64]CPUInfo `json:"cpuInfo"`
}

type CPUInfo struct {
	CPUCoreInfoMap map[int64]CPUCoreInfo `json:"cpuCoreInfo"`
	// CPU 制造商
	VendorId string `json:"vendorId"`
	// CPU 产品系列代号
	CPUFamily string `json:"cpuFamily"`
	// CPU 属于其产品系列那一代的编号
	Model string `json:"model"`
	// CPU 名字及其编号
	ModelName string `json:"modelName"`
	// CPU 制作更新版本
	Stepping string `json:"stepping"`
	// CPU 二级缓存大小
	CacheSize string `json:"cacheSize"`
	// CPU id
	PhysicalId int64 `json:"physicalId"`
	// CPU 逻辑核心数
	Siblings int64 `json:"siblings"`
	// CPU 物理核心数
	CPUCores int64 `json:"cpuCores"`
	// 是否有浮点运算单元
	FPU bool `json:"fpu"`
	// 是否支持浮点运算异常
	FPUException bool `json:"fpuException"`
	// 系统估算的CPU速度
	Bogomips float64 `json:"bogomips"`
	// 每次刷新缓存的大小单位
	ClFlushSize int64 `json:"clFlushSize"`
	// 缓存地址对齐单位
	CacheAlignment int64 `json:"cacheAlignment"`
	// 可访问地址空间位数
	AddressSizes string `json:"addressSizes"`
}

type CPUCoreInfo struct {
	CPUProcessorInfoMap map[int64]CPUProcessorInfo `json:"cpuProcessorInfo"`
	// CPU 物理核心ID
	CoreId int64 `json:"coreId"`
}

type CPUProcessorInfo struct {
	// 逻辑处理器编号
	Processor int64 `json:"processor"`
	// CPU 实际主频
	CPUMHz float64 `json:"CPUMHz"`
	// CPU 逻辑核编号
	Apicid int64 `json:"apicid"`
}

// CPUPerformanceMessage is a message to describe cpu Performance info
type CPUPerformanceMessage struct {
	Port              int                      `json:"port"`
	Host              string                   `json:"host"`
	User              string                   `json:"user"`
	Total             CPUPerformanceTotal      `json:"total"`
	CPUPerformanceMap map[int64]CPUPerformance `json:"cpuPerformance"`
}

type CPUPerformanceTotal struct {
	TotalTime   string  `json:"totalTime"`
	Utilization float64 `json:"utilization"`
	Free        float64 `json:"free"`
	System      float64 `json:"system"`
	User        float64 `json:"user"`
	IO          float64 `json:"IO"`
	Steal       float64 `json:"steal"`
}
type CPUPerformance struct {
	Processor   int64   `json:"processor"`
	Utilization float64 `json:"utilization"`
	Free        float64 `json:"free"`
	System      float64 `json:"system"`
	User        float64 `json:"user"`
	IO          float64 `json:"IO"`
	Steal       float64 `json:"steal"`
}

// MemoryPerformanceMessage is a message to describe memory Performance info
type MemoryPerformanceMessage struct {
	Port   int               `json:"port"`
	Host   string            `json:"host"`
	User   string            `json:"user"`
	Memory MemoryPerformance `json:"memory"`
}

type MemoryPerformance struct {
	TotalMem string `json:"totalMem"`
	// 未用占比
	FreeMemOccupy float64 `json:"freeMemOccupy"`
	// 未用
	FreeMem string `json:"freeMem"`
	// 可用占比
	AvailableMemOccupy float64 `json:"availableMemOccupy"`
	// 可用
	AvailableMem string `json:"availableMem"`
	// 块设备缓存占比
	BufferOccupy float64 `json:"bufferOccupy"`
	// 块设备缓存
	Buffer string `json:"buffer"`
	// 普通文件缓存占比
	CacheOccupy float64 `json:"cacheOccupy"`
	// 普通文件缓存
	Cached string `json:"cached"`
	// 脏页占比
	DirtyOccupy float64 `json:"dirtyOccupy"`
	// 脏页
	Dirty string `json:"dirty"`
	// 交换总内存
	SwapTotal string `json:"swapTotal"`
	// 交换可用占比
	SwapFreeOccupy float64 `json:"swapFreeOccupy"`
	// 交换可用内存
	SwapFree string `json:"swapFree"`
	// 交换文件缓存占比
	SwapCachedOccupy float64 `json:"swapCachedOccupy"`
	// 交换文件缓存
	SwapCached string `json:"swapCached"`
}
