package ssh

type Listener[M any] func(message M)

type Message struct {
	Port int    `json:"port"`
	Host string `json:"host"`
	User string `json:"user"`
}

type CPUInfoMessage struct {
	Message
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

type CPUPerformanceMessage struct {
	Message
	Total             CPUPerformanceTotal      `json:"total"`
	CPUPerformanceMap map[int64]CPUPerformance `json:"cpuPerformance"`
}

type CPUPerformanceTotal struct {
	TotalTime     int64   `json:"totalTime"`
	TotalTimeUnit string  `json:"totalTimeUnit"`
	Utilization   float64 `json:"utilization"`
	Free          float64 `json:"free"`
	System        float64 `json:"system"`
	User          float64 `json:"user"`
	IO            float64 `json:"IO"`
	Steal         float64 `json:"steal"`
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

type MemoryPerformanceMessage struct {
	Message
	Memory MemoryPerformance `json:"memory"`
}

type MemoryPerformance struct {
	TotalMem     float64 `json:"totalMem"`
	TotalMemUnit string  `json:"totalMemUnit"`
	// 未用占比
	FreeMemOccupy float64 `json:"freeMemOccupy"`
	// 未用
	FreeMem     float64 `json:"freeMem"`
	FreeMemUnit string  `json:"freeMemUnit"`
	// 可用占比
	AvailableMemOccupy float64 `json:"availableMemOccupy"`
	// 可用
	AvailableMem     float64 `json:"availableMem"`
	AvailableMemUnit string  `json:"availableMemUnit"`
	// 块设备缓存占比
	BufferOccupy float64 `json:"bufferOccupy"`
	// 块设备缓存
	Buffer     float64 `json:"buffer"`
	BufferUnit string  `json:"bufferUnit"`
	// 普通文件缓存占比
	CacheOccupy float64 `json:"cacheOccupy"`
	// 普通文件缓存
	Cached     float64 `json:"cached"`
	CachedUnit string  `json:"cachedUnit"`
	// 脏页占比
	DirtyOccupy float64 `json:"dirtyOccupy"`
	// 脏页
	Dirty     float64 `json:"dirty"`
	DirtyUnit string  `json:"dirtyUnit"`
	// 交换总内存
	SwapTotal     float64 `json:"swapTotal"`
	SwapTotalUnit string  `json:"swapTotalUnit"`
	// 交换可用占比
	SwapFreeOccupy float64 `json:"swapFreeOccupy"`
	// 交换可用内存
	SwapFree     float64 `json:"swapFree"`
	SwapFreeUnit string  `json:"swapFreeUnit"`
	// 交换文件缓存占比
	SwapCachedOccupy float64 `json:"swapCachedOccupy"`
	// 交换文件缓存
	SwapCached     float64 `json:"swapCached"`
	SwapCachedUnit string  `json:"swapCachedUnit"`
}

type UptimeMessage struct {
	Message
	Uptime Uptime `json:"uptime"`
}

type Uptime struct {
	UpDay  int64 `json:"upDay"`
	UpHour int64 `json:"UpHour"`
	UpMin  int64 `json:"upMin"`
	UpSec  int64 `json:"upSec"`
}

type LoadavgMessage struct {
	Message
	Loadavg Loadavg `json:"loadavg"`
}

type Loadavg struct {
	One           float64 `json:"one"`
	OneOccupy     float64 `json:"oneOccupy"`
	Five          float64 `json:"five"`
	FiveOccupy    float64 `json:"fiveOccupy"`
	Fifteen       float64 `json:"fifteen"`
	FifteenOccupy float64 `json:"fifteenOccupy"`
	Running       int64   `json:"running"`
	Active        int64   `json:"active"`
	LastPid       int64   `json:"lastPid"`
}

type NetDevMessage struct {
	Message
	NetDevTotal NetDevTotal       `json:"netDevTotal"`
	NetDevMap   map[string]NetDev `json:"netDev"`
}

type NetDevTotal struct {
	UpBytesH       float64 `json:"upBytesH"`
	UpBytesHUnit   string  `json:"upBytesHUnit"`
	UpBytes        int64   `json:"upBytes"`
	DownBytesH     float64 `json:"downBytesH"`
	DownBytesHUnit string  `json:"downBytesHUnit"`
	DownBytes      int64   `json:"downBytes"`
	UpPackets      int64   `json:"upPackets"`
	DownPackets    int64   `json:"downPackets"`
	UpSpeed        float64 `json:"upSpeed"`
	UpSpeedUnit    string  `json:"upSpeedUnit"`
	DownSpeed      float64 `json:"downSpeed"`
	DownSpeedUnit  string  `json:"downSpeedUnit"`
}

type NetDev struct {
	Name           string   `json:"name"`
	IP             []string `json:"ip"`
	Virtual        bool     `json:"virtual"`
	UpBytesH       float64  `json:"upBytesH"`
	UpBytesHUnit   string   `json:"upBytesHUnit"`
	UpBytes        int64    `json:"upBytes"`
	DownBytesH     float64  `json:"downBytesH"`
	DownBytesHUnit string   `json:"downBytesHUnit"`
	DownBytes      int64    `json:"downBytes"`
	UpPackets      int64    `json:"upPackets"`
	DownPackets    int64    `json:"downPackets"`
	UpSpeed        float64  `json:"upSpeed"`
	UpSpeedUnit    string   `json:"upSpeedUnit"`
	DownSpeed      float64  `json:"downSpeed"`
	DownSpeedUnit  string   `json:"downSpeedUnit"`
}

type NetStatMessage struct {
	Message
	NetTCP NetTCP `json:"netTCP"`
	NetUDP NetUDP `json:"netUDP"`
}

type NetUDP struct {
	InDatagrams      int64 `json:"inDatagrams"`
	OutDatagrams     int64 `json:"outDatagrams"`
	ReceiveBufErrors int64 `json:"receiveBufErrors"`
	SendBufErrors    int64 `json:"sendBufErrors"`
}
type NetTCP struct {
	ActiveOpens     int64   `json:"activeOpens"`
	PassiveOpens    int64   `json:"passiveOpens"`
	FailOpens       int64   `json:"failOpens"`
	CurrConn        int64   `json:"currConn"`
	InSegments      int64   `json:"inSegments"`
	OutSegments     int64   `json:"outSegments"`
	ReTransSegments int64   `json:"reTransSegments"`
	ReTransRate     float64 `json:"reTransRate"`
}

type TempMessage struct {
	Message
	TempMap map[string]int64 `json:"tempMap"`
}

type DiskMessage struct {
	Message
	DiskMap       map[string]Disk `json:"diskMap"`
	Write         float64         `json:"write"`
	WriteUnit     string          `json:"writeUnit"`
	Read          float64         `json:"read"`
	ReadUnit      string          `json:"readUnit"`
	WriteRate     float64         `json:"writeRate"`
	WriteRateUnit string          `json:"writeRateUnit"`
	ReadRate      float64         `json:"readRate"`
	ReadRateUnit  string          `json:"readRateUnit"`
}

type Disk struct {
	DevName       string  `json:"devName"`
	Mount         string  `json:"mount"`
	FileSystem    string  `json:"fileSystem"`
	Free          float64 `json:"free"`
	FreeUnit      string  `json:"freeUnit"`
	Total         float64 `json:"total"`
	TotalUnit     string  `json:"totalUnit"`
	FreeRate      float64 `json:"freeRate"`
	Write         float64 `json:"write"`
	WriteUnit     string  `json:"writeUnit"`
	Read          float64 `json:"read"`
	ReadUnit      string  `json:"readUnit"`
	WriteRate     float64 `json:"writeRate"`
	WriteRateUnit string  `json:"writeRateUnit"`
	ReadRate      float64 `json:"readRate"`
	ReadRateUnit  string  `json:"readRateUnit"`
	WriteIOPS     int64   `json:"writeIOPS"`
	ReadIOPS      int64   `json:"readIOPS"`
}

type ProcessMessage struct {
	Message
	Process []Process `json:"process"`
}

type Processes []Process

func (s Processes) Len() int { return len(s) }

func (s Processes) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

func (s Processes) Less(i, j int) bool {
	if s[i].CPU > s[j].CPU {
		return true
	} else if s[i].CPU < s[j].CPU {
		return false
	} else {
		return s[i].MemRaw > s[j].MemRaw
	}
}

type Process struct {
	PID     int64   `json:"PID"`
	Name    string  `json:"name"`
	CPU     float64 `json:"CPU"`
	Mem     float64 `json:"mem"`
	MemRaw  int64
	MemUnit string `json:"memUnit"`
}

type RoughMessage struct {
	Message
	CPU     RoughCPU     `json:"CPU"`
	Temp    RoughTemp    `json:"Temp"`
	Loadavg RoughLoadavg `json:"Loadavg"`
	Memory  RoughMemory  `json:"Memory"`
	Net     RoughNet     `json:"Net"`
	Disk    RoughDisk    `json:"Disk"`
}

type RoughCPU struct {
	Utilization float64 `json:"utilization"`
}

type RoughTemp struct {
	HighestTemp int64 `json:"highestTemp"`
}

type RoughLoadavg struct {
	OneOccupy     float64 `json:"oneOccupy"`
	FiveOccupy    float64 `json:"fiveOccupy"`
	FifteenOccupy float64 `json:"fifteenOccupy"`
}

type RoughMemory struct {
	FreeMemOccupy      float64 `json:"freeMemOccupy"`
	AvailableMemOccupy float64 `json:"availableMemOccupy"`
	SwapFreeOccupy     float64 `json:"swapFreeOccupy"`
}

type RoughNet struct {
	UpBytesH       float64 `json:"upBytesH"`
	UpBytesHUnit   string  `json:"upBytesHUnit"`
	DownBytesH     float64 `json:"downBytesH"`
	DownBytesHUnit string  `json:"downBytesHUnit"`
	UpSpeed        float64 `json:"upSpeed"`
	UpSpeedUnit    string  `json:"upSpeedUnit"`
	DownSpeed      float64 `json:"downSpeed"`
	DownSpeedUnit  string  `json:"downSpeedUnit"`
}

type RoughDisk struct {
	Write         float64 `json:"write"`
	WriteUnit     string  `json:"writeUnit"`
	Read          float64 `json:"read"`
	ReadUnit      string  `json:"readUnit"`
	WriteRate     float64 `json:"writeRate"`
	WriteRateUnit string  `json:"writeRateUnit"`
	ReadRate      float64 `json:"readRate"`
	ReadRateUnit  string  `json:"readRateUnit"`
}
