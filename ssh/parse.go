package ssh

import (
	"log"
	"message"
	"regexp"
	"strings"
)

func parseCPUInfoMessage(sshKey, s string) (m message.CPUInfoMessage) {
	m = message.CPUInfoMessage{
		CPUInfoMap: make(map[string]message.CPUInfo),
		SSHKey:     sshKey,
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
		physicalIdRegResults := physicalIdReg.FindAllSubmatch([]byte(cpu), -1)
		physicalId := string(physicalIdRegResults[0][1])
		cpuInfo, ok := m.CPUInfoMap[physicalId]
		if !ok {
			cpuInfo = message.CPUInfo{
				PhysicalId: physicalId,
			}
			m.CPUInfoMap[physicalId] = cpuInfo
		}
	}
	return
}
