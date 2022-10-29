package ssh

import (
	"fmt"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
)

func parseCPUInfoMessage(port int, host, user, old, new string) *CPUInfoMessage {
	m := &CPUInfoMessage{
		Message: Message{
			Port: port,
			Host: host,
			User: user,
		},
		CPUInfoMap: make(map[int64]CPUInfo),
	}
	cpus := strings.Split(new, "\n\n")
	for _, cpu := range cpus {
		cpu = strings.TrimSpace(cpu)
		if len(cpu) == 0 {
			continue
		}
		// 正则匹配
		physicalIdReg := regexp.MustCompile(`physical id\t: (\d+)\n`)
		if physicalIdReg == nil {
			log.Fatalf("regexp parse fail : physical id")
		}
		physicalIdRegResults := physicalIdReg.FindAllStringSubmatch(cpu, -1)
		if physicalIdRegResults == nil {
			log.Printf("parse physical id fail")
			continue
		}
		physicalId, err := strconv.ParseInt(physicalIdRegResults[0][1], 10, 64)
		if err != nil {
			log.Printf("parse int fail : %v", err)
			continue
		}
		cpuInfo, ok := m.CPUInfoMap[physicalId]
		if !ok {
			cpuInfo = CPUInfo{
				CPUCoreInfoMap: make(map[int64]CPUCoreInfo),
				PhysicalId:     physicalId,
			}
			lines := strings.Split(cpu, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "vendor_id") {
					cpuInfo.VendorId = strings.TrimSpace(strings.Split(line, ":")[1])
				} else if strings.HasPrefix(line, "cpu family") {
					cpuInfo.CPUFamily = strings.TrimSpace(strings.Split(line, ":")[1])
				} else if strings.HasPrefix(line, "model name") {
					cpuInfo.ModelName = strings.TrimSpace(strings.Split(line, ":")[1])
				} else if strings.HasPrefix(line, "model") {
					cpuInfo.Model = strings.TrimSpace(strings.Split(line, ":")[1])
				} else if strings.HasPrefix(line, "stepping") {
					cpuInfo.Stepping = strings.TrimSpace(strings.Split(line, ":")[1])
				} else if strings.HasPrefix(line, "cache size") {
					cpuInfo.CacheSize = strings.TrimSpace(strings.Split(line, ":")[1])
				} else if strings.HasPrefix(line, "siblings") {
					temp := strings.TrimSpace(strings.Split(line, ":")[1])
					parseInt, err := strconv.ParseInt(temp, 10, 64)
					if err != nil {
						log.Printf("parse int fail : %v", err)
						continue
					}
					cpuInfo.Siblings = parseInt
				} else if strings.HasPrefix(line, "cpu cores") {
					temp := strings.TrimSpace(strings.Split(line, ":")[1])
					parseInt, err := strconv.ParseInt(temp, 10, 64)
					if err != nil {
						log.Printf("parse int fail : %v", err)
						continue
					}
					cpuInfo.CPUCores = parseInt
				} else if strings.HasPrefix(line, "fpu") {
					cpuInfo.FPU = strings.TrimSpace(strings.Split(line, ":")[1]) == "yes"
				} else if strings.HasPrefix(line, "fpu_exception") {
					cpuInfo.FPUException = strings.TrimSpace(strings.Split(line, ":")[1]) == "yes"
				} else if strings.HasPrefix(line, "bogomips") {
					temp := strings.TrimSpace(strings.Split(line, ":")[1])
					parseFloat, err := strconv.ParseFloat(temp, 64)
					if err != nil {
						log.Printf("parse float fail : %v", err)
						continue
					}
					cpuInfo.Bogomips = parseFloat
				} else if strings.HasPrefix(line, "clflush size") {
					temp := strings.TrimSpace(strings.Split(line, ":")[1])
					parseInt, err := strconv.ParseInt(temp, 10, 64)
					if err != nil {
						log.Printf("parse int fail : %v", err)
						continue
					}
					cpuInfo.ClFlushSize = parseInt
				} else if strings.HasPrefix(line, "cache_alignment") {
					temp := strings.TrimSpace(strings.Split(line, ":")[1])
					parseInt, err := strconv.ParseInt(temp, 10, 64)
					if err != nil {
						log.Printf("parse int fail : %v", err)
						continue
					}
					cpuInfo.CacheAlignment = parseInt
				} else if strings.HasPrefix(line, "address sizes") {
					cpuInfo.AddressSizes = strings.TrimSpace(strings.Split(line, ":")[1])
				}
			}
			m.CPUInfoMap[physicalId] = cpuInfo
		}
		coreIdReg := regexp.MustCompile(`core id\t\t: (\d+)\n`)
		if coreIdReg == nil {
			log.Fatalf("regexp parse fail : core id")
		}
		coreIdRegResults := coreIdReg.FindAllStringSubmatch(cpu, -1)
		if coreIdRegResults == nil {
			log.Printf("parse core id fail")
			continue
		}
		coreId, err := strconv.ParseInt(coreIdRegResults[0][1], 10, 64)
		if err != nil {
			log.Printf("parse int fail : %v", err)
			continue
		}
		cpuCoreInfo, ok := cpuInfo.CPUCoreInfoMap[coreId]
		if !ok {
			cpuCoreInfo = CPUCoreInfo{
				CPUProcessorInfoMap: make(map[int64]CPUProcessorInfo),
				CoreId:              coreId,
			}
			cpuInfo.CPUCoreInfoMap[coreId] = cpuCoreInfo
		}
		processorIdReg := regexp.MustCompile(`processor\t: (\d+)\n`)
		if processorIdReg == nil {
			log.Fatalf("regexp parse fail : processor id")
		}
		processorIdRegResults := processorIdReg.FindAllStringSubmatch(cpu, -1)
		if processorIdRegResults == nil {
			log.Printf("parse processor id fail")
			continue
		}
		processorId, err := strconv.ParseInt(processorIdRegResults[0][1], 10, 64)
		if err != nil {
			log.Printf("parse int fail : %v", err)
			continue
		}
		cpuProcessorInfo, ok := cpuCoreInfo.CPUProcessorInfoMap[processorId]
		if !ok {
			cpuProcessorInfo = CPUProcessorInfo{
				Processor: processorId,
			}
			lines := strings.Split(cpu, "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "cpu MHz") {
					temp := strings.TrimSpace(strings.Split(line, ":")[1])
					parseFloat, err := strconv.ParseFloat(temp, 64)
					if err != nil {
						log.Printf("parse float fail : %v", err)
						continue
					}
					cpuProcessorInfo.CPUMHz = parseFloat
				} else if strings.HasPrefix(line, "apicid") {
					temp := strings.TrimSpace(strings.Split(line, ":")[1])
					parseInt, err := strconv.ParseInt(temp, 10, 64)
					if err != nil {
						log.Printf("parse int fail : %v", err)
						continue
					}
					cpuProcessorInfo.Apicid = parseInt
				}
			}
			cpuCoreInfo.CPUProcessorInfoMap[processorId] = cpuProcessorInfo
		}
	}
	return m
}

func parseCPUPerformanceMessage(port int, host, user, old, new string) *CPUPerformanceMessage {
	if old == "" {
		return nil
	}
	// Linux 时间片默认为 10ms
	jiffies := int64(10)
	m := &CPUPerformanceMessage{
		Message: Message{
			Port: port,
			Host: host,
			User: user,
		},
		Total:             CPUPerformanceTotal{},
		CPUPerformanceMap: make(map[int64]CPUPerformance),
	}
	reg := regexp.MustCompile(`cpu(\d*)\s+(\d+) (\d+) (\d+) (\d+) (\d+) (\d+) (\d+) (\d+) (\d+) (\d+)\n`)

	if reg == nil {
		log.Fatalf("regexp parse fail : cpu performance")
	}
	oldResult := reg.FindAllStringSubmatch(old, -1)
	newResult := reg.FindAllStringSubmatch(new, -1)

	if oldResult == nil || newResult == nil {
		log.Printf("parse cpu performance fail")
		return nil
	}
	// 运行时间
	newTotalCPUTime := int64(0)
	for i := 2; i < len(newResult[0]); i++ {
		newTotalCPUTime += parseInt64(newResult[0][i])
	}
	totalCPUTime := newTotalCPUTime * jiffies
	coreNum := int64(len(newResult) - 1)
	if totalCPUTime > coreNum*1000*60*60*24 {
		m.Total.TotalTime = fmt.Sprintf("%dD", totalCPUTime/coreNum/1000/60/60/24)
	} else if totalCPUTime > coreNum*1000*60*60 {
		m.Total.TotalTime = fmt.Sprintf("%dH", totalCPUTime/coreNum/1000/60/60)
	} else if totalCPUTime > coreNum*1000*60 {
		m.Total.TotalTime = fmt.Sprintf("%dMin", totalCPUTime/coreNum/1000/60)
	} else {
		m.Total.TotalTime = fmt.Sprintf("%dS", totalCPUTime/coreNum/1000)
	}

	// 利用率
	oldTotalCPUTime := int64(0)
	for i := 2; i < len(oldResult[0]); i++ {
		oldTotalCPUTime += parseInt64(oldResult[0][i])
	}
	diff := float64(newTotalCPUTime - oldTotalCPUTime)
	m.Total.Utilization = roundFloat(float64(100)-float64(100*(parseInt64(newResult[0][5])-parseInt64(oldResult[0][5])))/diff, 2)
	m.Total.Free = roundFloat(float64(100*(parseInt64(newResult[0][5])-parseInt64(oldResult[0][5])))/diff, 2)
	m.Total.System = roundFloat(float64(100*(parseInt64(newResult[0][4])-parseInt64(oldResult[0][4])))/diff, 2)
	m.Total.User = roundFloat(float64(100*(parseInt64(newResult[0][2])-parseInt64(oldResult[0][2])))/diff, 2)
	m.Total.IO = roundFloat(float64(100*(parseInt64(newResult[0][6])-parseInt64(oldResult[0][6])))/diff, 2)
	m.Total.Steal = roundFloat(float64(100*(parseInt64(newResult[0][9])-parseInt64(oldResult[0][9])))/diff, 2)

	for i := 1; i < len(oldResult); i++ {
		pOldTotalCPUTime := int64(0)
		for j := 2; j < len(oldResult[i]); j++ {
			pOldTotalCPUTime += parseInt64(oldResult[i][j])
		}
		pNewTotalCPUTime := int64(0)
		for j := 2; j < len(newResult[i]); j++ {
			pNewTotalCPUTime += parseInt64(newResult[i][j])
		}
		pDiff := float64(pNewTotalCPUTime - pOldTotalCPUTime)
		processor := parseInt64(oldResult[i][1])
		c := CPUPerformance{}
		c.Processor = processor
		c.Utilization = roundFloat(float64(100)-float64(100*(parseInt64(newResult[i][5])-parseInt64(oldResult[i][5])))/pDiff, 2)
		c.Free = roundFloat(float64(100*(parseInt64(newResult[i][5])-parseInt64(oldResult[i][5])))/pDiff, 2)
		c.System = roundFloat(float64(100*(parseInt64(newResult[i][4])-parseInt64(oldResult[i][4])))/pDiff, 2)
		c.User = roundFloat(float64(100*(parseInt64(newResult[i][2])-parseInt64(oldResult[i][2])))/pDiff, 2)
		c.IO = roundFloat(float64(100*(parseInt64(newResult[i][6])-parseInt64(oldResult[i][6])))/pDiff, 2)
		c.Steal = roundFloat(float64(100*(parseInt64(newResult[i][9])-parseInt64(oldResult[i][9])))/pDiff, 2)
		m.CPUPerformanceMap[processor] = c
	}
	return m
}

func parseMemoryPerformanceMessage(port int, host, user, old, new string) *MemoryPerformanceMessage {
	m := &MemoryPerformanceMessage{
		Message: Message{
			Port: port,
			Host: host,
			User: user,
		},
		Memory: MemoryPerformance{},
	}

	MemTotalReg := regexp.MustCompile(`MemTotal:\D+(\d+) kB\n`)
	if MemTotalReg == nil {
		log.Fatalf("regexp parse fail : memory total")
	}
	MemTotalRegResults := MemTotalReg.FindAllStringSubmatch(new, -1)
	if MemTotalRegResults == nil {
		log.Printf("parse memory total fail")
		return nil
	}
	TotalMem := parseInt64(MemTotalRegResults[0][1])
	m.Memory.TotalMem = roundMem(TotalMem)

	SwapTotalReg := regexp.MustCompile(`SwapTotal:\D+(\d+) kB\n`)
	if SwapTotalReg == nil {
		log.Fatalf("regexp parse fail : swap total")
	}
	SwapTotalRegResult := SwapTotalReg.FindAllStringSubmatch(new, -1)
	if SwapTotalRegResult == nil {
		log.Printf("parse swap total fail")
		return nil
	}
	SwapTotal := parseInt64(SwapTotalRegResult[0][1])
	m.Memory.SwapTotal = roundMem(SwapTotal)

	lines := strings.Split(new, "\n")
	for _, line := range lines {
		if !strings.Contains(line, ":") {
			continue
		}
		line := strings.Replace(line, "kB", "", len(lines)-len("kB"))
		number := strings.TrimSpace(strings.Split(line, ":")[1])
		if strings.HasPrefix(line, "MemFree:") {
			t := parseInt64(number)
			m.Memory.FreeMem = roundMem(t)
			m.Memory.FreeMemOccupy = roundFloat(float64(t)/float64(TotalMem), 2)
		} else if strings.HasPrefix(line, "MemAvailable:") {
			t := parseInt64(number)
			m.Memory.AvailableMem = roundMem(t)
			m.Memory.AvailableMemOccupy = roundFloat(float64(t)/float64(TotalMem), 2)
		} else if strings.HasPrefix(line, "Buffers:") {
			t := parseInt64(number)
			m.Memory.Buffer = roundMem(t)
			m.Memory.BufferOccupy = roundFloat(float64(t)/float64(TotalMem), 2)
		} else if strings.HasPrefix(line, "Cached:") {
			t := parseInt64(number)
			m.Memory.Cached = roundMem(t)
			m.Memory.CacheOccupy = roundFloat(float64(t)/float64(TotalMem), 2)
		} else if strings.HasPrefix(line, "Dirty:") {
			t := parseInt64(number)
			m.Memory.Dirty = roundMem(t)
			m.Memory.DirtyOccupy = roundFloat(float64(t)/float64(TotalMem), 2)
		} else if strings.HasPrefix(line, "SwapCached:") {
			t := parseInt64(number)
			m.Memory.SwapCached = roundMem(t)
			m.Memory.SwapCachedOccupy = roundFloat(float64(t)/float64(TotalMem), 2)
		} else if strings.HasPrefix(line, "SwapFree:") {
			t := parseInt64(number)
			m.Memory.SwapFree = roundMem(t)
			m.Memory.SwapFreeOccupy = roundFloat(float64(t)/float64(TotalMem), 2)
		}
	}

	return m
}

func parseUptimeMessage(port int, host, user, old, new string) *UptimeMessage {
	m := &UptimeMessage{
		Message: Message{
			Port: port,
			Host: host,
			User: user,
		},
		Uptime: Uptime{},
	}
	uptime, err := strconv.ParseFloat(strings.Split(new, " ")[0], 64)
	if err != nil {
		log.Printf("parse float fail : %v", err)
		return nil
	}
	m.Uptime.UpDay = int64(uptime) / int64(60*60*24)
	uptime = math.Mod(uptime, float64(60*60*24))
	m.Uptime.UpHour = int64(uptime) / int64(60*60)
	uptime = math.Mod(uptime, float64(60*60))
	m.Uptime.UpMin = int64(uptime) / int64(60)
	uptime = math.Mod(uptime, float64(60))
	m.Uptime.UpSec = int64(uptime)
	return m
}

func parseInt64(s string) int64 {
	parseInt, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Printf("parse int fail %s : %v", s, err)
		return 0
	}
	return parseInt
}

func roundMem(kb int64) string {
	if kb > 1024*1024 {
		return fmt.Sprintf("%.2fGB", float64(kb)/1024/1024)
	} else if kb > 1024 {
		return fmt.Sprintf("%.2fMB", float64(kb)/1024)
	} else {
		return fmt.Sprintf("%dKB", kb)
	}
}

func roundFloat(num float64, n int) float64 {
	s := "%." + fmt.Sprintf("%d", n) + "f"
	value, _ := strconv.ParseFloat(fmt.Sprintf(s, num), 64)
	return value
}
