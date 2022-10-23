package message

type CPUInfoMessage struct {
	Port       int               `json:"port"`
	Host       string            `json:"host"`
	User       string            `json:"user"`
	CPUInfoMap map[int64]CPUInfo `json:"cpu_info"`
}

type CPUInfo struct {
	CPUCoreInfoMap map[int64]CPUCoreInfo `json:"cpu_core_info"`
	// CPU 制造商
	VendorId string `json:"vendor_id"`
	// CPU 产品系列代号
	CPUFamily string `json:"cpu_family"`
	// CPU 属于其产品系列那一代的编号
	Model string `json:"model"`
	// CPU 名字及其编号
	ModelName string `json:"model_name"`
	// CPU 制作更新版本
	Stepping string `json:"stepping"`
	// CPU 二级缓存大小
	CacheSize string `json:"cache_size"`
	// CPU id
	PhysicalId int64 `json:"physical_id"`
	// CPU 逻辑核心数
	Siblings int64 `json:"siblings"`
	// CPU 物理核心数
	CPUCores int64 `json:"cpu_cores"`
	// 是否有浮点运算单元
	FPU bool `json:"fpu"`
	// 是否支持浮点运算异常
	FPUException bool `json:"fpu_exception"`
	// 系统估算的CPU速度
	Bogomips float64 `json:"bogomips"`
	// 每次刷新缓存的大小单位
	ClFlushSize int64 `json:"cl_flush_size"`
	// 缓存地址对齐单位
	CacheAlignment int64 `json:"cache_alignment"`
	// 可访问地址空间位数
	AddressSizes string `json:"address_sizes"`
}

type CPUCoreInfo struct {
	CPUProcessorInfoMap map[int64]CPUProcessorInfo `json:"cpu_processor_info"`
	// CPU 物理核心ID
	CoreId int64 `json:"core_id"`
}

type CPUProcessorInfo struct {
	// 逻辑处理器编号
	Processor int64 `json:"processor"`
	// CPU 实际主频
	CPUMHz float64 `json:"CPUMHz"`
	// CPU 逻辑核编号
	Apicid int64 `json:"apicid"`
}

type CPUInfoListener func(message CPUInfoMessage)
