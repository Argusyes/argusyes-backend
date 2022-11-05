package ssh

import (
	"encoding/binary"
	"fmt"
	mapSet "github.com/deckarep/golang-set/v2"
	"github.com/pkg/sftp"
	"io/ioutil"
	"log"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Parser struct {
	CPU struct {
		CPUProcessorNum int64
		Utilization     float64
	}
	Temp struct {
		HighestTemp int64
	}
	Loadavg struct {
		OneOccupy     float64
		FiveOccupy    float64
		FifteenOccupy float64
	}
	Memory struct {
		FreeMemOccupy      float64
		AvailableMemOccupy float64
		CachedSwapOccupy   float64
	}
	Net struct {
		UpBytesH       float64
		UpBytesHUnit   string
		DownBytesH     float64
		DownBytesHUnit string
		UpSpeed        float64
		UpSpeedUnit    string
		DownSpeed      float64
		DownSpeedUnit  string
	}
	Disk struct {
		Write         float64
		WriteUnit     string
		Read          float64
		ReadUnit      string
		WriteRate     float64
		WriteRateUnit string
		ReadRate      float64
		ReadRateUnit  string
	}
}

func (p *Parser) parseRoughMessage(port int, host, user string) *RoughMessage {
	m := &RoughMessage{
		Message: Message{
			Port: port,
			Host: host,
			User: user,
		},
		CPU: RoughCPU{
			Utilization: p.CPU.Utilization,
		},
		Temp: RoughTemp{
			HighestTemp: p.Temp.HighestTemp,
		},
		Loadavg: RoughLoadavg{
			OneOccupy:     p.Loadavg.OneOccupy,
			FiveOccupy:    p.Loadavg.FiveOccupy,
			FifteenOccupy: p.Loadavg.FifteenOccupy,
		},
		Memory: RoughMemory{
			FreeMemOccupy:      p.Memory.FreeMemOccupy,
			AvailableMemOccupy: p.Memory.AvailableMemOccupy,
			CacheSwapOccupy:    p.Memory.CachedSwapOccupy,
		},
		Net: RoughNet{
			UpBytesH:       p.Net.UpBytesH,
			UpBytesHUnit:   p.Net.UpBytesHUnit,
			DownBytesH:     p.Net.DownBytesH,
			DownBytesHUnit: p.Net.DownBytesHUnit,
			UpSpeed:        p.Net.UpSpeed,
			UpSpeedUnit:    p.Net.UpSpeedUnit,
			DownSpeed:      p.Net.DownSpeed,
			DownSpeedUnit:  p.Net.DownSpeedUnit,
		},
		Disk: RoughDisk{
			Write:         p.Disk.Write,
			WriteUnit:     p.Disk.WriteUnit,
			Read:          p.Disk.Read,
			ReadUnit:      p.Disk.ReadUnit,
			WriteRate:     p.Disk.WriteRate,
			WriteRateUnit: p.Disk.WriteRateUnit,
			ReadRate:      p.Disk.ReadRate,
			ReadRateUnit:  p.Disk.ReadRateUnit,
		},
	}
	return m
}

func (p *Parser) parseCPUInfoMessage(c *MonitorContext) *CPUInfoMessage {
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
					p.CPU.CPUProcessorNum = parseInt
				} else if strings.HasPrefix(line, "cpu cores") {
					temp := strings.TrimSpace(strings.Split(line, ":")[1])
					parseInt, err := strconv.ParseInt(temp, 10, 64)
					if err != nil {
						log.Printf("parse int fail : %v", err)
						continue
					}
					cpuInfo.CPUCores = parseInt
				} else if strings.HasPrefix(line, "fpu_exception") {
					cpuInfo.FPUException = strings.TrimSpace(strings.Split(line, ":")[1]) == "yes"
				} else if strings.HasPrefix(line, "fpu") {
					cpuInfo.FPU = strings.TrimSpace(strings.Split(line, ":")[1]) == "yes"
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

func (p *Parser) parseCPUPerformanceMessage(c *MonitorContext) *CPUPerformanceMessage {
	defer func() {
		c.oldTime = c.newTime
		c.oldS = c.newS
		c.newS = ""
	}()
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
		m.Total.TotalTime = totalCPUTime / coreNum / 1000 / 60 / 60 / 24
		m.Total.TotalTimeUnit = "D"
	} else if totalCPUTime > coreNum*1000*60*60 {
		m.Total.TotalTime = totalCPUTime / coreNum / 1000 / 60 / 60
		m.Total.TotalTimeUnit = "H"
	} else if totalCPUTime > coreNum*1000*60 {
		m.Total.TotalTime = totalCPUTime / coreNum / 1000 / 60
		m.Total.TotalTimeUnit = "M"
	} else {
		m.Total.TotalTime = totalCPUTime / coreNum / 1000
		m.Total.TotalTimeUnit = "S"
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
	if diff == 0 {
		diff++
	}
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

	p.CPU.Utilization = m.Total.Utilization

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
		if pDiff == 0 {
			pDiff++
		}
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

func (p *Parser) parseMemoryPerformanceMessage(c *MonitorContext) *MemoryPerformanceMessage {
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
	m.Memory.TotalMem, m.Memory.TotalMemUnit = roundMem(TotalMem * 1024)
	if TotalMem == 0 {
		TotalMem++
	}

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
	m.Memory.TotalSwap, m.Memory.TotalSwapUnit = roundMem(SwapTotal * 1024)
	if SwapTotal == 0 {
		SwapTotal++
	}
	UsedMem := int64(0)
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
			m.Memory.FreeMem, m.Memory.FreeMemUnit = roundMem(t * 1024)
			m.Memory.FreeMemOccupy = roundFloat(float64(t)/float64(TotalMem), 2)
			p.Memory.FreeMemOccupy = m.Memory.FreeMemOccupy
		} else if strings.HasPrefix(line, "MemAvailable:") {
			t, ok := parseInt64(number)
			if !ok {
				return nil
			}
			UsedMem -= t
			m.Memory.AvailableMem, m.Memory.AvailableMemUnit = roundMem(t * 1024)
			m.Memory.AvailableMemOccupy = roundFloat(float64(t)/float64(TotalMem), 2)
			p.Memory.AvailableMemOccupy = m.Memory.AvailableMemOccupy
		} else if strings.HasPrefix(line, "Buffers:") {
			t, ok := parseInt64(number)
			if !ok {
				return nil
			}
			m.Memory.Buffer, m.Memory.BufferUnit = roundMem(t * 1024)
			m.Memory.BufferOccupy = roundFloat(float64(t)/float64(TotalMem), 2)
		} else if strings.HasPrefix(line, "Cached:") {
			t, ok := parseInt64(number)
			if !ok {
				return nil
			}
			m.Memory.Cached, m.Memory.CachedUnit = roundMem(t * 1024)
			m.Memory.CacheOccupy = roundFloat(float64(t)/float64(TotalMem), 2)
		} else if strings.HasPrefix(line, "Dirty:") {
			t, ok := parseInt64(number)
			if !ok {
				return nil
			}
			m.Memory.Dirty, m.Memory.DirtyUnit = roundMem(t * 1024)
			m.Memory.DirtyOccupy = roundFloat(float64(t)/float64(TotalMem), 2)
		} else if strings.HasPrefix(line, "SwapCached:") {
			t, ok := parseInt64(number)
			if !ok {
				return nil
			}
			m.Memory.CachedSwap, m.Memory.CachedSwapUnit = roundMem(t * 1024)
			m.Memory.CachedSwapOccupy = roundFloat(float64(t)/float64(SwapTotal), 2)
			p.Memory.CachedSwapOccupy = m.Memory.CachedSwapOccupy
		} else if strings.HasPrefix(line, "SwapFree:") {
			t, ok := parseInt64(number)
			if !ok {
				return nil
			}
			m.Memory.FreeSwap, m.Memory.FreeSwapUnit = roundMem(t * 1024)
			m.Memory.FreeSwapOccupy = roundFloat(float64(t)/float64(SwapTotal), 2)
		}
	}
	m.Memory.UsedMem, m.Memory.UsedMemUnit = roundMem(UsedMem * 1024)
	m.Memory.UsedMemOccupy = roundFloat(float64(UsedMem)/float64(TotalMem), 2)
	return m
}

func (p *Parser) parseUptimeMessage(c *MonitorContext) *UptimeMessage {
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

func (p *Parser) parseLoadavgMessage(c *MonitorContext) *LoadavgMessage {
	m := &LoadavgMessage{
		Message: Message{
			Port: c.port,
			Host: c.host,
			User: c.user,
		},
		Loadavg: Loadavg{},
	}
	if p.CPU.CPUProcessorNum == 0 {
		return nil
	}
	n := strings.ReplaceAll(c.newS, "\n", "")
	ss := strings.Split(n, " ")
	ok := true
	if m.Loadavg.One, ok = parseFloat64(ss[0]); !ok {
		return nil
	}
	m.Loadavg.OneOccupy = roundFloat(m.Loadavg.One/float64(p.CPU.CPUProcessorNum), 2)
	if m.Loadavg.Five, ok = parseFloat64(ss[1]); !ok {
		return nil
	}
	m.Loadavg.FiveOccupy = roundFloat(m.Loadavg.Five/float64(p.CPU.CPUProcessorNum), 2)
	if m.Loadavg.Fifteen, ok = parseFloat64(ss[2]); !ok {
		return nil
	}
	m.Loadavg.FifteenOccupy = roundFloat(m.Loadavg.Fifteen/float64(p.CPU.CPUProcessorNum), 2)
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
	p.Loadavg.OneOccupy = m.Loadavg.OneOccupy
	p.Loadavg.FiveOccupy = m.Loadavg.FiveOccupy
	p.Loadavg.FifteenOccupy = m.Loadavg.FifteenOccupy
	return m
}

func (p *Parser) parseNetDevMessage(c *MonitorContext) *NetDevMessage {
	defer func() {
		c.oldTime = c.newTime
		c.oldS = c.newS
		c.newS = ""
	}()
	if c.oldS == "" {
		return nil
	}
	m := &NetDevMessage{
		Message: Message{
			Port: c.port,
			Host: c.host,
			User: c.user,
		},
		NetDevTotal: NetDevTotal{},
		NetDevMap:   make(map[string]NetDev, 0),
	}

	reg := regexp.MustCompile(`([^:\n]+):\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\n`)
	if reg == nil {
		log.Fatalf("regexp parse fail : Net Dev")
	}
	oldResult := reg.FindAllStringSubmatch(c.oldS, -1)
	newResult := reg.FindAllStringSubmatch(c.newS, -1)
	if oldResult == nil || newResult == nil {
		log.Printf("parse Net Dev fail")
		return nil
	}
	oldMap := make(map[string][]string, 0)
	for _, ss := range oldResult {
		oldMap[strings.TrimSpace(ss[1])] = ss
	}
	// route info
	route, ok := readFile("/proc/net/route", c.client, true)
	if !ok {
		return nil
	}
	fib, ok := readFile("/proc/net/fib_trie", c.client, true)
	if !ok {
		return nil
	}

	routeMap := make(map[string]mapSet.Set[string], 0)

	for i, line := range strings.Split(route, "\n") {
		if i == 0 || line == "" {
			continue
		}
		lineSplit := strings.Split(line, "\t")
		gw, ok := parseInt64Base16(lineSplit[2])
		if !ok {
			continue
		}
		if gw != 0 {
			continue
		}
		// 取出所有网关为0的
		name := lineSplit[0]
		ipBigEnd, ok := parseInt64Base16(lineSplit[1])
		if !ok {
			continue
		}
		var ipPreBytes = make([]byte, 4)
		binary.LittleEndian.PutUint32(ipPreBytes, uint32(ipBigEnd))
		ipPre := fmt.Sprintf("%d.%d.%d.%d", ipPreBytes[0], ipPreBytes[1], ipPreBytes[2], ipPreBytes[3])

		if set, ok := routeMap[name]; ok {
			set.Add(ipPre)
		} else {
			set = mapSet.NewSet(ipPre)
			routeMap[name] = set
		}
	}

	EndIPReg := regexp.MustCompile(`.* (\d+\.\d+\.\d+\.\d+)$`)
	if EndIPReg == nil {
		log.Fatalf("regexp parse fail : EndIP")
		return nil
	}

	// 网卡IP
	fibMap := make(map[string]mapSet.Set[string], 0)
	for key, val := range routeMap {
		fibSet := mapSet.NewSet[string]()
		fibMap[key] = fibSet
		it := val.Iterator()
		for ipPre := range it.C {
			// 去除本地
			if strings.HasPrefix(ipPre, "169.254") {
				continue
			}
			FibReg := regexp.MustCompile(ipPre + `([^L]*)\n\s+/32 host LOCAL\n`)
			if FibReg == nil {
				continue
			}
			FibResult := FibReg.FindAllStringSubmatch(fib, -1)
			if FibResult == nil {
				continue
			}
			EndIPResult := EndIPReg.FindAllStringSubmatch(FibResult[0][1], -1)
			if EndIPResult == nil {
				continue
			}
			fibSet.Add(EndIPResult[0][1])
		}
	}

	oldTotalUpBytes := int64(0)
	oldTotalDownBytes := int64(0)
	difTime := c.newTime.Sub(c.oldTime).Milliseconds()
	if difTime == 0 {
		difTime++
	}
	for _, ss := range newResult {
		name := strings.TrimSpace(ss[1])
		oldSS, ok := oldMap[name]
		if !ok {
			continue
		}
		n := NetDev{
			Name: name,
			IP:   make([]string, 0),
		}
		ipSet, ok := fibMap[name]
		if ok {
			it := ipSet.Iterator()
			for s := range it.C {
				n.IP = append(n.IP, s)
			}
		}
		// 判断虚拟化
		n.Virtual = true
		_, err := c.client.Stat("/sys/devices/virtual/net/" + name)
		if err != nil {
			n.Virtual = false
		}
		if n.DownBytes, ok = parseInt64(ss[2]); !ok {
			continue
		}
		if n.DownPackets, ok = parseInt64(ss[3]); !ok {
			continue
		}
		if n.UpBytes, ok = parseInt64(ss[10]); !ok {
			continue
		}
		if n.UpPackets, ok = parseInt64(ss[11]); !ok {
			continue
		}
		n.UpBytesH, n.UpBytesHUnit = roundMem(n.UpBytes)
		n.DownBytesH, n.DownBytesHUnit = roundMem(n.DownBytes)
		oldUpBytes, ok := parseInt64(oldSS[10])
		if !ok {
			continue
		}
		oldDownBytes, ok := parseInt64(oldSS[2])
		if !ok {
			continue
		}
		n.UpSpeed, n.UpSpeedUnit = roundMem((n.UpBytes - oldUpBytes) * 1000 / difTime)
		n.DownSpeed, n.DownSpeedUnit = roundMem((n.DownBytes - oldDownBytes) * 1000 / difTime)

		if !n.Virtual {
			m.NetDevTotal.UpBytes += n.UpBytes
			m.NetDevTotal.DownBytes += n.DownBytes
			m.NetDevTotal.UpPackets += n.UpPackets
			m.NetDevTotal.DownPackets += n.DownPackets
			oldTotalDownBytes += oldDownBytes
			oldTotalUpBytes += oldUpBytes
		}
		m.NetDevMap[name] = n
	}
	m.NetDevTotal.UpBytesH, m.NetDevTotal.UpBytesHUnit = roundMem(m.NetDevTotal.UpBytes)
	m.NetDevTotal.DownBytesH, m.NetDevTotal.DownBytesHUnit = roundMem(m.NetDevTotal.DownBytes)
	m.NetDevTotal.UpSpeed, m.NetDevTotal.UpSpeedUnit = roundMem((m.NetDevTotal.UpBytes - oldTotalUpBytes) * 1000 / difTime)
	m.NetDevTotal.DownSpeed, m.NetDevTotal.DownSpeedUnit = roundMem((m.NetDevTotal.DownBytes - oldTotalDownBytes) * 1000 / difTime)
	p.Net.UpSpeed, p.Net.UpSpeedUnit = m.NetDevTotal.UpSpeed, m.NetDevTotal.UpSpeedUnit
	p.Net.DownSpeed, p.Net.DownSpeedUnit = m.NetDevTotal.DownSpeed, m.NetDevTotal.DownSpeedUnit
	p.Net.UpBytesH, p.Net.UpBytesHUnit = m.NetDevTotal.UpBytesH, m.NetDevTotal.UpBytesHUnit
	p.Net.DownBytesH, p.Net.DownBytesHUnit = m.NetDevTotal.DownBytesH, m.NetDevTotal.DownBytesHUnit
	return m
}

func (p *Parser) parseNetStatMessage(c *MonitorContext) *NetStatMessage {
	m := &NetStatMessage{
		Message{
			Port: c.port,
			Host: c.host,
			User: c.user,
		},
		NetTCP{},
		NetUDP{},
	}

	TCPReg := regexp.MustCompile(`Tcp:\s+(\d+)\s+(\d+)\s+(\d+)\s+([\d-]+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\n`)
	if TCPReg == nil {
		log.Fatalf("regexp parse fail : net stat tcp")
	}
	TCPRegResults := TCPReg.FindAllStringSubmatch(c.newS, -1)
	if TCPRegResults == nil {
		log.Printf("parse net stat tcp fail")
		return nil
	}

	UDPReg := regexp.MustCompile(`Udp:\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)[^\n]+\n`)
	if UDPReg == nil {
		log.Fatalf("regexp parse fail : net stat udp")
	}
	UDPRegResults := UDPReg.FindAllStringSubmatch(c.newS, -1)
	if UDPRegResults == nil {
		log.Printf("parse net stat udp fail")
		return nil
	}
	ok := true
	if m.NetTCP.ActiveOpens, ok = parseInt64(TCPRegResults[0][5]); !ok {
		return nil
	}
	if m.NetTCP.PassiveOpens, ok = parseInt64(TCPRegResults[0][6]); !ok {
		return nil
	}
	if m.NetTCP.FailOpens, ok = parseInt64(TCPRegResults[0][7]); !ok {
		return nil
	}
	if m.NetTCP.CurrConn, ok = parseInt64(TCPRegResults[0][8]); !ok {
		return nil
	}
	if m.NetTCP.InSegments, ok = parseInt64(TCPRegResults[0][9]); !ok {
		return nil
	}
	if m.NetTCP.OutSegments, ok = parseInt64(TCPRegResults[0][10]); !ok {
		return nil
	}
	if m.NetTCP.ReTransSegments, ok = parseInt64(TCPRegResults[0][11]); !ok {
		return nil
	}
	if m.NetTCP.OutSegments != 0 {
		m.NetTCP.ReTransRate = roundFloat(float64(m.NetTCP.ReTransSegments)/float64(m.NetTCP.OutSegments), 2)
	}
	if m.NetUDP.InDatagrams, ok = parseInt64(UDPRegResults[0][1]); !ok {
		return nil
	}
	if m.NetUDP.OutDatagrams, ok = parseInt64(UDPRegResults[0][4]); !ok {
		return nil
	}
	if m.NetUDP.ReceiveBufErrors, ok = parseInt64(UDPRegResults[0][5]); !ok {
		return nil
	}
	if m.NetUDP.SendBufErrors, ok = parseInt64(UDPRegResults[0][6]); !ok {
		return nil
	}

	return m
}

func (p *Parser) parseTempMessage(c *MonitorContext) *TempMessage {
	m := &TempMessage{
		Message: Message{
			Port: c.port,
			Host: c.host,
			User: c.user,
		},
		TempMap: make(map[string]int64, 0),
	}
	ok := true
	if m.TempMap["zone0"], ok = parseInt64(strings.TrimSpace(c.newS)); !ok {
		return nil
	}
	p.Temp.HighestTemp = m.TempMap["zone0"]
	for i := 1; i < 10; i++ {
		file := fmt.Sprintf("/sys/class/thermal/thermal_zone%d/temp", i)
		zone := fmt.Sprintf("zone%d", i)
		temp, ok := readFile(file, c.client, false)
		if !ok {
			break
		}
		if m.TempMap[zone], ok = parseInt64(strings.TrimSpace(temp)); ok {
			if m.TempMap[zone] > p.Temp.HighestTemp {
				p.Temp.HighestTemp = m.TempMap[zone]
			}
		} else {
			break
		}
	}
	return m
}

func (p *Parser) parseDiskMessage(c *MonitorContext) *DiskMessage {
	defer func() {
		c.oldTime = c.newTime
		c.oldS = c.newS
		c.newS = ""
	}()
	if c.oldS == "" {
		return nil
	}
	m := &DiskMessage{
		Message: Message{
			Port: c.port,
			Host: c.host,
			User: c.user,
		},
		DiskMap: make(map[string]Disk, 0),
	}
	mounts, ok := readFile("/proc/mounts", c.client, true)
	if !ok {
		return nil
	}
	MountReg := regexp.MustCompile(`(\S+) (\S+) (\S+) (\S+) (\d+) (\d+)\n`)
	if MountReg == nil {
		log.Fatalf("regexp parse fail : mounts")
	}
	MountStatsRegResults := MountReg.FindAllStringSubmatch(mounts, -1)
	if MountStatsRegResults == nil {
		log.Printf("parse mounts fail")
		return nil
	}
	mountSet := mapSet.NewSet[string]()
	mountMap := make(map[string][]string, 0)
	for _, ss := range MountStatsRegResults {
		if ss[3] == "ext4" || ss[3] == "vfat" {
			mountMap[ss[1]] = ss
			mountSet.Add(ss[1])
		}
	}

	DiskReg := regexp.MustCompile(`(\d+)\s+(\d+)\s+(\S+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)[^\n]+\n`)
	if DiskReg == nil {
		log.Fatalf("regexp parse fail : disk")
	}
	oldDiskRegResults := DiskReg.FindAllStringSubmatch(c.oldS, -1)
	newDiskRegResults := DiskReg.FindAllStringSubmatch(c.newS, -1)
	diff := c.newTime.Sub(c.oldTime).Milliseconds()
	if diff == 0 {
		diff++
	}
	if oldDiskRegResults == nil || newDiskRegResults == nil {
		log.Printf("parse disk fail")
		return nil
	}
	oldDiskMap := make(map[string][]string, 0)
	newDiskMap := make(map[string][]string, 0)
	for _, ss := range oldDiskRegResults {
		devName := fmt.Sprintf("/dev/%s", ss[3])
		if mountSet.Contains(devName) {
			oldDiskMap[devName] = ss
		}
	}
	for _, ss := range newDiskRegResults {
		devName := fmt.Sprintf("/dev/%s", ss[3])
		if mountSet.Contains(devName) {
			newDiskMap[devName] = ss
		}
	}

	OldTotalWrite := int64(0)
	OldTotalRead := int64(0)
	NewTotalWrite := int64(0)
	NewTotalRead := int64(0)
	SectorSize := int64(512)
	it := mountSet.Iterator()
	for devName := range it.C {
		oldSS, ok := oldDiskMap[devName]
		if !ok {
			continue
		}
		newSS, ok := newDiskMap[devName]
		if !ok {
			continue
		}
		mountSS, ok := mountMap[devName]
		if !ok {
			continue
		}
		d := Disk{
			DevName: devName,
		}

		d.Mount = mountSS[2]
		d.FileSystem = mountSS[3]
		stat, err := c.client.StatVFS(d.Mount)
		if err != nil {
			continue
		}

		total := int64(stat.Blocks * stat.Bsize)
		free := int64(stat.Bfree * stat.Bsize)
		d.Free, d.FreeUnit = roundMem(free)
		d.Total, d.TotalUnit = roundMem(total)
		if total != 0 {
			d.FreeRate = roundFloat(float64(free)/float64(total), 2)
		}

		oldWriteSector, ok := parseInt64(oldSS[10])
		if !ok {
			continue
		}
		oldWrite := SectorSize * oldWriteSector
		newWriteSector, ok := parseInt64(newSS[10])
		if !ok {
			continue
		}
		newWrite := SectorSize * newWriteSector
		d.Write, d.WriteUnit = roundMem(newWrite)
		d.WriteRate, d.WriteRateUnit = roundMem((newWrite - oldWrite) * 1000 / diff)
		oldReadSector, ok := parseInt64(oldSS[6])
		if !ok {
			continue
		}
		oldRead := SectorSize * oldReadSector
		newReadSector, ok := parseInt64(newSS[6])
		if !ok {
			continue
		}
		newRead := SectorSize * newReadSector
		d.Read, d.ReadUnit = roundMem(newRead)
		d.ReadRate, d.ReadRateUnit = roundMem((newRead - oldRead) * 1000 / diff)

		OldTotalWrite += oldWrite
		OldTotalRead += oldRead
		NewTotalWrite += newWrite
		NewTotalRead += newRead

		oldWriteIO, ok := parseInt64(oldSS[8])
		if !ok {
			continue
		}
		newWriteIO, ok := parseInt64(newSS[8])
		if !ok {
			continue
		}
		d.WriteIOPS = (newWriteIO - oldWriteIO) * 1000 / diff
		oldReadIO, ok := parseInt64(oldSS[4])
		if !ok {
			continue
		}
		newReadIO, ok := parseInt64(newSS[4])
		if !ok {
			continue
		}
		d.ReadIOPS = (newReadIO - oldReadIO) * 1000 / diff

		m.DiskMap[devName] = d
	}
	m.Write, m.WriteUnit = roundMem(NewTotalWrite)
	m.Read, m.ReadUnit = roundMem(NewTotalRead)
	m.WriteRate, m.WriteRateUnit = roundMem((NewTotalWrite - OldTotalWrite) * 1000 / diff)
	m.ReadRate, m.ReadRateUnit = roundMem((NewTotalRead - OldTotalRead) * 1000 / diff)
	p.Disk.Write, p.Disk.WriteUnit = m.Write, m.WriteUnit
	p.Disk.Read, p.Disk.ReadUnit = m.Read, m.ReadUnit
	p.Disk.WriteRate, p.Disk.WriteRateUnit = m.WriteRate, m.WriteRateUnit
	p.Disk.ReadRate, p.Disk.ReadRateUnit = m.ReadRate, m.ReadRateUnit
	return m
}

func (p *Parser) parseProcessMessage(c *MonitorContext) *ProcessMessage {

	cpuStat, ok := readFile("/proc/stat", c.client, true)
	if !ok {
		return nil
	}
	cpuStat = strings.Split(cpuStat, "\n")[0]

	proc, err := c.client.ReadDir("/proc")
	if err != nil {
		log.Printf("Read Proc fail : %v", err)
		return nil
	}
	numberReg := regexp.MustCompile(`\d+`)
	name := make([]string, 0)
	for _, n := range proc {
		s := n.Name()
		if numberReg.Match([]byte(s)) && n.IsDir() {
			name = append(name, s)
		}
	}

	stats := make([]string, 0)
	stats = append(stats, cpuStat)
	var mutex sync.Mutex
	var wg sync.WaitGroup
	count := 0
	var cMutex sync.Mutex
	total := len(name)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			for {
				if count >= total {
					break
				} else {
					cMutex.Lock()
					n := name[count]
					count++
					cMutex.Unlock()

					stat, ok := readFile("/proc/"+n+"/stat", c.client, false)
					if !ok {
						continue
					}
					stat = strings.ReplaceAll(stat, "\n", "")
					m, ok := readFile("/proc/"+n+"/statm", c.client, false)
					if !ok {
						continue
					}
					m = strings.ReplaceAll(m, "\n", "")
					mutex.Lock()
					stats = append(stats, stat+" "+m)
					mutex.Unlock()
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()

	c.newS = strings.Join(stats, "\n")
	c.newTime = time.Now()

	defer func() {
		c.oldTime = c.newTime
		c.oldS = c.newS
		c.newS = ""
	}()

	if c.oldS == "" || p.CPU.CPUProcessorNum == 0 {
		return nil
	}
	m := &ProcessMessage{
		Message: Message{
			Port: c.port,
			Host: c.host,
			User: c.user,
		},
		Process: make([]Process, 0),
	}
	oldStats := strings.Split(c.oldS, "\n")

	oldCPUStat := oldStats[0]
	oldCPUStatSS := strings.Split(oldCPUStat, " ")
	oldTotalCPUTime := int64(0)
	for i := 1; i < len(oldCPUStatSS); i++ {
		if oldCPUStatSS[i] == "" {
			continue
		}
		t, ok := parseInt64(oldCPUStatSS[i])
		if ok {
			oldTotalCPUTime += t
		}
	}

	cpuStatSS := strings.Split(cpuStat, " ")
	newTotalCPUTime := int64(0)
	for i := 1; i < len(cpuStatSS); i++ {
		if cpuStatSS[i] == "" {
			continue
		}
		t, ok := parseInt64(cpuStatSS[i])
		if ok {
			newTotalCPUTime += t
		}
	}
	cpuTimeDiff := newTotalCPUTime - oldTotalCPUTime
	if cpuTimeDiff == 0 {
		cpuTimeDiff++
	}

	statReg := regexp.MustCompile(`^\s*(\d+)\s+(\S+)\s+(\S+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+([\d-]+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+).*\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)\s+(\d+)$`)
	if statReg == nil {
		log.Printf("parse stat reg fail : %v", statReg)
		return nil
	}

	oldStatsMap := make(map[string][]string, 0)
	for i, s := range oldStats {
		if i == 0 {
			continue
		}
		ss := statReg.FindAllStringSubmatch(s, -1)
		if ss == nil {
			continue
		}
		oldStatsMap[ss[0][1]+ss[0][2]] = ss[0]
	}

	for i, s := range stats {
		if i == 0 {
			continue
		}
		statRegResult := statReg.FindAllStringSubmatch(s, -1)
		if statRegResult == nil {
			continue
		}
		ss := statRegResult[0]
		oldSS, ok := oldStatsMap[ss[1]+ss[2]]
		if !ok {
			continue
		}
		pid, ok := parseInt64(strings.TrimSpace(ss[1]))
		if !ok {
			continue
		}
		process := Process{
			PID:  pid,
			Name: strings.TrimSpace(ss[2]),
		}

		utime, ok := parseInt64(ss[14])
		if !ok {
			continue
		}
		stime, ok := parseInt64(ss[15])
		if !ok {
			continue
		}
		cutime, ok := parseInt64(ss[16])
		if !ok {
			continue
		}
		cstime, ok := parseInt64(ss[17])
		if !ok {
			continue
		}
		processTotalTime := utime + stime + cutime + cstime

		oldUtime, ok := parseInt64(oldSS[14])
		if !ok {
			continue
		}
		oldStime, ok := parseInt64(oldSS[15])
		if !ok {
			continue
		}
		oldCutime, ok := parseInt64(oldSS[16])
		if !ok {
			continue
		}
		oldCstime, ok := parseInt64(oldSS[17])
		if !ok {
			continue
		}
		oldProcessTotalTime := oldUtime + oldStime + oldCutime + oldCstime
		process.CPU = roundFloat(float64(p.CPU.CPUProcessorNum)*100*(float64(processTotalTime)-float64(oldProcessTotalTime))/float64(cpuTimeDiff), 2)
		mem, ok := parseInt64(ss[len(ss)-6])
		if !ok {
			continue
		}
		process.MemRaw = mem * 4096
		process.Mem, process.MemUnit = roundMem(mem * 4096)
		if process.CPU > 0 || process.Mem > 0 {
			m.Process = append(m.Process, process)
		}
	}

	sort.Sort(Processes(m.Process))
	m.Process = m.Process[:50]
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

func parseInt64Base16(s string) (int64, bool) {
	parseInt, err := strconv.ParseInt(s, 16, 64)
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

func roundMem(b int64) (float64, string) {
	if b > 1024*1024*1024*1024 {
		return roundFloat(float64(b)/1024/1024/1024/1024, 2), "TB"
	} else if b > 1024*1024*1024 {
		return roundFloat(float64(b)/1024/1024/1024, 2), "GB"
	} else if b > 1024*1024 {
		return roundFloat(float64(b)/1024/1024, 2), "MB"
	} else if b > 1024 {
		return roundFloat(float64(b)/1024, 2), "KB"
	} else {
		return float64(b), "B"
	}
}

func roundFloat(num float64, n int) float64 {
	s := "%." + fmt.Sprintf("%d", n) + "f"
	value, _ := strconv.ParseFloat(fmt.Sprintf(s, num), 64)
	return value
}

func readFile(where string, client *sftp.Client, doLog bool) (string, bool) {
	srcFile, err := client.OpenFile(where, os.O_RDONLY)
	if err != nil {
		if doLog {
			log.Printf("Read %s file fail : %v", where, err)
		}
		return "", false
	}
	f, err := ioutil.ReadAll(srcFile)
	if err != nil {
		if doLog {
			log.Printf("Read %s file fail : %v", where, err)
		}
		return "", false
	}
	err = srcFile.Close()
	if err != nil {
		if doLog {
			log.Printf("Close %s file fail : %v", where, err)
		}
	}
	return string(f), true
}
