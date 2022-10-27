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
	m.Total.Utilization = float64(100) - float64(100*(parseInt64(newResult[0][5])-parseInt64(oldResult[0][5])))/diff
	m.Total.Free = float64(100*(parseInt64(newResult[0][5])-parseInt64(oldResult[0][5]))) / diff
	m.Total.System = float64(100*(parseInt64(newResult[0][4])-parseInt64(oldResult[0][4]))) / diff
	m.Total.User = float64(100*(parseInt64(newResult[0][2])-parseInt64(oldResult[0][2]))) / diff
	m.Total.IO = float64(100*(parseInt64(newResult[0][6])-parseInt64(oldResult[0][6]))) / diff
	m.Total.Steal = float64(100*(parseInt64(newResult[0][9])-parseInt64(oldResult[0][9]))) / diff

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
		c.Utilization = float64(100) - float64(100*(parseInt64(newResult[i][5])-parseInt64(oldResult[i][5])))/pDiff
		c.Free = float64(100*(parseInt64(newResult[i][5])-parseInt64(oldResult[i][5]))) / pDiff
		c.System = float64(100*(parseInt64(newResult[i][4])-parseInt64(oldResult[i][4]))) / pDiff
		c.User = float64(100*(parseInt64(newResult[i][2])-parseInt64(oldResult[i][2]))) / pDiff
		c.IO = float64(100*(parseInt64(newResult[i][6])-parseInt64(oldResult[i][6]))) / pDiff
		c.Steal = float64(100*(parseInt64(newResult[i][9])-parseInt64(oldResult[i][9]))) / pDiff
		m.CPUPerformanceMap[processor] = c
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
