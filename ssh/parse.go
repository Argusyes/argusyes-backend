package ssh

import (
	"fmt"
	"log"
	"message"
	"regexp"
	"strconv"
	"strings"
)

func parseCPUInfoMessage(port int, host, user, s string) (m message.CPUInfoMessage) {
	m = message.CPUInfoMessage{
		Port:       port,
		Host:       host,
		User:       user,
		CPUInfoMap: make(map[int64]message.CPUInfo),
	}
	cpus := strings.Split(s, "\n\n")
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
			cpuInfo = message.CPUInfo{
				CPUCoreInfoMap: make(map[int64]message.CPUCoreInfo),
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
			cpuCoreInfo = message.CPUCoreInfo{
				CPUProcessorInfoMap: make(map[int64]message.CPUProcessorInfo),
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
			cpuProcessorInfo = message.CPUProcessorInfo{
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
	return
}

func parseCPUPerformanceMessage(port int, host, user, old, new string) message.CPUPerformanceMessage {
	// Linux 时间片默认为 10ms
	jiffies := int64(10)
	m := message.CPUPerformanceMessage{
		Port:              port,
		Host:              host,
		User:              user,
		Total:             message.CPUPerformanceTotal{},
		CPUPerformanceMap: make(map[int64]message.CPUPerformance),
	}
	reg := regexp.MustCompile(`cpu(\d*)\s+(\d+) (\d+) (\d+) (\d+) (\d+) (\d+) (\d+) (\d+) (\d+) (\d+)\n`)

	if reg == nil {
		log.Fatalf("regexp parse fail : cpu performance")
	}
	oldResult := reg.FindAllStringSubmatch(old, -1)
	newResult := reg.FindAllStringSubmatch(new, -1)

	if oldResult == nil || newResult == nil {
		log.Printf("parse cpu performance fail")
		return m
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
		c := message.CPUPerformance{}
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

func parseMemoryPerformanceMessage(port int, host, user, s string) message.MemoryPerformanceMessage {
	m := message.MemoryPerformanceMessage{
		Port:   port,
		Host:   host,
		User:   user,
		Memory: message.MemoryPerformance{},
	}

	MemTotalReg := regexp.MustCompile(`MemTotal:\D+(\d+) kB\n`)
	if MemTotalReg == nil {
		log.Fatalf("regexp parse fail : memory total")
	}
	MemTotalRegResults := MemTotalReg.FindAllStringSubmatch(s, -1)
	if MemTotalRegResults == nil {
		log.Printf("parse memory total fail")
		return m
	}
	TotalMem := parseInt64(MemTotalRegResults[0][1])
	m.Memory.TotalMem = roundMem(TotalMem)

	SwapTotalReg := regexp.MustCompile(`SwapTotal:\D+(\d+) kB\n`)
	if SwapTotalReg == nil {
		log.Fatalf("regexp parse fail : swap total")
	}
	SwapTotalRegResult := SwapTotalReg.FindAllStringSubmatch(s, -1)
	if SwapTotalRegResult == nil {
		log.Printf("parse swap total fail")
		return m
	}
	SwapTotal := parseInt64(SwapTotalRegResult[0][1])
	m.Memory.SwapTotal = roundMem(SwapTotal)

	lines := strings.Split(s, "\n")
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
