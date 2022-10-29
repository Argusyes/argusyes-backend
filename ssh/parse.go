package ssh

import (
	"fmt"
	"log"
	"math"
	"regexp"
	"strconv"
	"strings"
)

type Parser struct {
	CPUProcessorNum int64
}

func (p *Parser) parseCPUInfoMessage(c MonitorContext) *CPUInfoMessage {
	m := &CPUInfoMessage{
		Message: Message{
			Port: c.port,
			Host: c.host,
			User: c.user,
		},
		CPUInfoMap: make(map[int64]CPUInfo),
	}
	cpus := strings.Split(c.newS, "\n\n")
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
					p.CPUProcessorNum = parseInt
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

func (p *Parser) parseCPUPerformanceMessage(c MonitorContext) *CPUPerformanceMessage {
	if c.oldS == "" {
		return nil
	}
	// Linux 时间片默认为 10ms
	jiffies := int64(10)
	m := &CPUPerformanceMessage{
		Message: Message{
			Port: c.port,
			Host: c.host,
			User: c.user,
		},
		Total:             CPUPerformanceTotal{},
		CPUPerformanceMap: make(map[int64]CPUPerformance),
	}
	reg := regexp.MustCompile(`cpu(\d*)\s+(\d+) (\d+) (\d+) (\d+) (\d+) (\d+) (\d+) (\d+) (\d+) (\d+)\n`)

	if reg == nil {
		log.Fatalf("regexp parse fail : cpu performance")
	}
	oldResult := reg.FindAllStringSubmatch(c.oldS, -1)
	newResult := reg.FindAllStringSubmatch(c.newS, -1)

	if oldResult == nil || newResult == nil {
		log.Printf("parse cpu performance fail")
		return nil
	}
	// 运行时间
	newTotalCPUTime := int64(0)
	for i := 2; i < len(newResult[0]); i++ {
		if x, ok := parseInt64(newResult[0][i]); ok {
			newTotalCPUTime += x
		} else {
			return nil
		}
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
		if x, ok := parseInt64(oldResult[0][i]); ok {
			oldTotalCPUTime += x
		} else {
			return nil
		}
	}
	diff := float64(newTotalCPUTime - oldTotalCPUTime)
	newUtilization, ok := parseInt64(newResult[0][5])
	if !ok {
		return nil
	}
	newSystem, ok := parseInt64(newResult[0][4])
	if !ok {
		return nil
	}
	newUser, ok := parseInt64(newResult[0][2])
	if !ok {
		return nil
	}
	newIO, ok := parseInt64(newResult[0][6])
	if !ok {
		return nil
	}
	newSteal, ok := parseInt64(newResult[0][9])
	if !ok {
		return nil
	}
	oldUtilization, ok := parseInt64(oldResult[0][5])
	if !ok {
		return nil
	}
	oldSystem, ok := parseInt64(oldResult[0][4])
	if !ok {
		return nil
	}
	oldUser, ok := parseInt64(oldResult[0][2])
	if !ok {
		return nil
	}
	oldIO, ok := parseInt64(oldResult[0][6])
	if !ok {
		return nil
	}
	oldSteal, ok := parseInt64(oldResult[0][9])
	if !ok {
		return nil
	}
	m.Total.Utilization = roundFloat(float64(100)-float64(100*(newUtilization-oldUtilization))/diff, 2)
	m.Total.Free = roundFloat(float64(100*(newUtilization-oldUtilization))/diff, 2)
	m.Total.System = roundFloat(float64(100*(newSystem-oldSystem))/diff, 2)
	m.Total.User = roundFloat(float64(100*(newUser-oldUser))/diff, 2)
	m.Total.IO = roundFloat(float64(100*(newIO-oldIO))/diff, 2)
	m.Total.Steal = roundFloat(float64(100*(newSteal-oldSteal))/diff, 2)

	for i := 1; i < len(oldResult); i++ {
		pOldTotalCPUTime := int64(0)
		for j := 2; j < len(oldResult[i]); j++ {
			if x, ok := parseInt64(oldResult[i][j]); ok {
				pOldTotalCPUTime += x
			} else {
				return nil
			}
		}
		pNewTotalCPUTime := int64(0)
		for j := 2; j < len(newResult[i]); j++ {
			if x, ok := parseInt64(newResult[i][j]); ok {
				pNewTotalCPUTime += x
			} else {
				return nil
			}
		}
		pDiff := float64(pNewTotalCPUTime - pOldTotalCPUTime)
		newPUtilization, ok := parseInt64(newResult[i][5])
		if !ok {
			return nil
		}
		newPSystem, ok := parseInt64(newResult[i][4])
		if !ok {
			return nil
		}
		newPUser, ok := parseInt64(newResult[i][2])
		if !ok {
			return nil
		}
		newPIO, ok := parseInt64(newResult[i][6])
		if !ok {
			return nil
		}
		newPSteal, ok := parseInt64(newResult[i][9])
		if !ok {
			return nil
		}
		oldPUtilization, ok := parseInt64(oldResult[i][5])
		if !ok {
			return nil
		}
		oldPSystem, ok := parseInt64(oldResult[i][4])
		if !ok {
			return nil
		}
		oldPUser, ok := parseInt64(oldResult[i][2])
		if !ok {
			return nil
		}
		oldPIO, ok := parseInt64(oldResult[i][6])
		if !ok {
			return nil
		}
		oldPSteal, ok := parseInt64(oldResult[i][9])
		if !ok {
			return nil
		}

		processor, ok := parseInt64(oldResult[i][1])
		if !ok {
			return nil
		}
		c := CPUPerformance{}
		c.Processor = processor
		c.Utilization = roundFloat(float64(100)-float64(100*(newPUtilization-oldPUtilization))/pDiff, 2)
		c.Free = roundFloat(float64(100*(newPUtilization-oldPUtilization))/pDiff, 2)
		c.System = roundFloat(float64(100*(newPSystem-oldPSystem))/pDiff, 2)
		c.User = roundFloat(float64(100*(newPUser-oldPUser))/pDiff, 2)
		c.IO = roundFloat(float64(100*(newPIO-oldPIO))/pDiff, 2)
		c.Steal = roundFloat(float64(100*(newPSteal-oldPSteal))/pDiff, 2)
		m.CPUPerformanceMap[processor] = c
	}
	return m
}

func (p *Parser) parseMemoryPerformanceMessage(c MonitorContext) *MemoryPerformanceMessage {
	m := &MemoryPerformanceMessage{
		Message: Message{
			Port: c.port,
			Host: c.host,
			User: c.user,
		},
		Memory: MemoryPerformance{},
	}

	MemTotalReg := regexp.MustCompile(`MemTotal:\D+(\d+) kB\n`)
	if MemTotalReg == nil {
		log.Fatalf("regexp parse fail : memory total")
	}
	MemTotalRegResults := MemTotalReg.FindAllStringSubmatch(c.newS, -1)
	if MemTotalRegResults == nil {
		log.Printf("parse memory total fail")
		return nil
	}
	TotalMem, ok := parseInt64(MemTotalRegResults[0][1])
	if !ok {
		return nil
	}
	m.Memory.TotalMem = roundMem(TotalMem)

	SwapTotalReg := regexp.MustCompile(`SwapTotal:\D+(\d+) kB\n`)
	if SwapTotalReg == nil {
		log.Fatalf("regexp parse fail : swap total")
	}
	SwapTotalRegResult := SwapTotalReg.FindAllStringSubmatch(c.newS, -1)
	if SwapTotalRegResult == nil {
		log.Printf("parse swap total fail")
		return nil
	}
	SwapTotal, ok := parseInt64(SwapTotalRegResult[0][1])
	if !ok {
		return nil
	}
	m.Memory.SwapTotal = roundMem(SwapTotal)

	lines := strings.Split(c.newS, "\n")
	for _, line := range lines {
		if !strings.Contains(line, ":") {
			continue
		}
		line := strings.Replace(line, "kB", "", len(lines)-len("kB"))
		number := strings.TrimSpace(strings.Split(line, ":")[1])
		if strings.HasPrefix(line, "MemFree:") {
			t, ok := parseInt64(number)
			if !ok {
				return nil
			}
			m.Memory.FreeMem = roundMem(t)
			m.Memory.FreeMemOccupy = roundFloat(float64(t)/float64(TotalMem), 2)
		} else if strings.HasPrefix(line, "MemAvailable:") {
			t, ok := parseInt64(number)
			if !ok {
				return nil
			}
			m.Memory.AvailableMem = roundMem(t)
			m.Memory.AvailableMemOccupy = roundFloat(float64(t)/float64(TotalMem), 2)
		} else if strings.HasPrefix(line, "Buffers:") {
			t, ok := parseInt64(number)
			if !ok {
				return nil
			}
			m.Memory.Buffer = roundMem(t)
			m.Memory.BufferOccupy = roundFloat(float64(t)/float64(TotalMem), 2)
		} else if strings.HasPrefix(line, "Cached:") {
			t, ok := parseInt64(number)
			if !ok {
				return nil
			}
			m.Memory.Cached = roundMem(t)
			m.Memory.CacheOccupy = roundFloat(float64(t)/float64(TotalMem), 2)
		} else if strings.HasPrefix(line, "Dirty:") {
			t, ok := parseInt64(number)
			if !ok {
				return nil
			}
			m.Memory.Dirty = roundMem(t)
			m.Memory.DirtyOccupy = roundFloat(float64(t)/float64(TotalMem), 2)
		} else if strings.HasPrefix(line, "SwapCached:") {
			t, ok := parseInt64(number)
			if !ok {
				return nil
			}
			m.Memory.SwapCached = roundMem(t)
			m.Memory.SwapCachedOccupy = roundFloat(float64(t)/float64(TotalMem), 2)
		} else if strings.HasPrefix(line, "SwapFree:") {
			t, ok := parseInt64(number)
			if !ok {
				return nil
			}
			m.Memory.SwapFree = roundMem(t)
			m.Memory.SwapFreeOccupy = roundFloat(float64(t)/float64(TotalMem), 2)
		}
	}

	return m
}

func (p *Parser) parseUptimeMessage(c MonitorContext) *UptimeMessage {
	m := &UptimeMessage{
		Message: Message{
			Port: c.port,
			Host: c.host,
			User: c.user,
		},
		Uptime: Uptime{},
	}
	uptime, err := strconv.ParseFloat(strings.Split(c.newS, " ")[0], 64)
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

func (p *Parser) parseLoadavgMessage(c MonitorContext) *LoadavgMessage {
	m := &LoadavgMessage{
		Message: Message{
			Port: c.port,
			Host: c.host,
			User: c.user,
		},
		Loadavg: Loadavg{},
	}
	if p.CPUProcessorNum == 0 {
		return nil
	}
	n := strings.ReplaceAll(c.newS, "\n", "")
	ss := strings.Split(n, " ")
	ok := true
	if m.Loadavg.One, ok = parseFloat64(ss[0]); !ok {
		return nil
	}
	m.Loadavg.OneOccupy = roundFloat(m.Loadavg.One/float64(p.CPUProcessorNum), 2)
	if m.Loadavg.Five, ok = parseFloat64(ss[1]); !ok {
		return nil
	}
	m.Loadavg.FiveOccupy = roundFloat(m.Loadavg.Five/float64(p.CPUProcessorNum), 2)
	if m.Loadavg.Fifteen, ok = parseFloat64(ss[2]); !ok {
		return nil
	}
	m.Loadavg.FifteenOccupy = roundFloat(m.Loadavg.Fifteen/float64(p.CPUProcessorNum), 2)
	thread := strings.Split(ss[3], "/")
	if m.Loadavg.Running, ok = parseInt64(thread[0]); !ok {
		return nil
	}
	if m.Loadavg.Active, ok = parseInt64(thread[1]); !ok {
		return nil
	}
	if m.Loadavg.LastPid, ok = parseInt64(ss[4]); !ok {
		return nil
	}
	return m
}

func (p *Parser) parseNetDevMessage(c MonitorContext) *NetDevMessage {
	m := &NetDevMessage{
		Message: Message{
			Port: c.port,
			Host: c.host,
			User: c.user,
		},
	}

	return m
}

func parseInt64(s string) (int64, bool) {
	parseInt, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Printf("parse int fail %s : %v", s, err)
		return 0, false
	}
	return parseInt, true
}

func parseFloat64(s string) (float64, bool) {
	parseFloat, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Printf("parse float fail %s : %v", s, err)
		return 0, false
	}
	return parseFloat, true
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
