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
	cpus := strings.Split(s, "\r\n\r\n")
	for _, cpu := range cpus {
		cpu = strings.TrimSpace(cpu)
		if len(cpu) == 0 {
			continue
		}
		// 正则匹配
		physicalIdReg := regexp.MustCompile(`physical id\t: (\d+)\r\n`)
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
			lines := strings.Split(cpu, "\r\n")
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
		coreIdReg := regexp.MustCompile(`core id\t\t: (\d+)\r\n`)
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
		processorIdReg := regexp.MustCompile(`processor\t: (\d+)\r\n`)
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
			lines := strings.Split(cpu, "\r\n")
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
	jiffies := int64(1)
	m := message.CPUPerformanceMessage{
		Port:              port,
		Host:              host,
		User:              user,
		Total:             message.CPUPerformanceTotal{},
		CPUPerformanceMap: make(map[int64]message.CPUPerformance),
	}
	reg := regexp.MustCompile(`cpu(\d*)\s+(\d+) (\d+) (\d+) (\d+) (\d+) (\d+) (\d+) (\d+) (\d+) (\d+)\r\n`)

	if reg == nil {
		log.Fatalf("regexp parse fail : cpu performance")
	}
	oldResult := reg.FindAllStringSubmatch(old, -1)
	newResult := reg.FindAllStringSubmatch(new, -1)

	// 运行时间
	totalCPUTime := int64(0)
	for i := 2; i < len(newResult[0]); i++ {
		totalCPUTime += parseInt64(newResult[0][i])
	}
	totalCPUTime *= jiffies
	if totalCPUTime > 60*60*24 {
		m.Total.TotalTime = fmt.Sprintf("%dD", totalCPUTime/1000/60/60/24)
	} else if totalCPUTime > 60*60 {
		m.Total.TotalTime = fmt.Sprintf("%dH", totalCPUTime/1000/60/60)
	} else if totalCPUTime > 60 {
		m.Total.TotalTime = fmt.Sprintf("%dMin", totalCPUTime/1000/60)
	} else {
		m.Total.TotalTime = fmt.Sprintf("%dS", totalCPUTime/1000)
	}

	// 利用率
	oldTotalCPUTime := int64(0)
	for i := 2; i < len(oldResult[0]); i++ {
		oldTotalCPUTime += parseInt64(oldResult[0][i])
	}
	oldTotalCPUTime *= jiffies
	utilization := float64(100) - (float64(100*jiffies*(parseInt64(newResult[0][5])-parseInt64(oldResult[0][5]))) / float64(totalCPUTime-oldTotalCPUTime))
	m.Total.Utilization = fmt.Sprintf("%.2f%", utilization)
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
