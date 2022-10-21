package ssh

import (
	"bufio"
	"io"
	"log"
	"message"
	"strconv"
	"strings"
)

func parseCPUInfoMessage(s string) (m message.CPUInfoMessage) {
	reader := bufio.NewReader(strings.NewReader(s))
	m = message.CPUInfoMessage{
		CPUInfo: make([]message.CPUInfo, 0),
	}
	info := message.CPUInfo{}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				log.Printf("ParseCPUInfo error : %v", err)
			}
			return m
		}
		if strings.HasPrefix(line, "processor") {
			line = strings.ReplaceAll(line, "\r\n", "")
			line = strings.ReplaceAll(line, " ", "")
			splits := strings.Split(line, ":")
			res, err := strconv.ParseInt(splits[1], 10, 64)
			if err != nil {
				log.Printf("ParseCPUInfo processor error : %v", err)
			}
			info.Processor = res
		} else if strings.HasPrefix(line, "model name") {
			splits := strings.Split(line, ":")
			info.ModelName = splits[1]
		} else if strings.HasPrefix(line, "power management") {
			m.CPUInfo = append(m.CPUInfo, info)
			info = message.CPUInfo{}
			m.ProcessorNum = int64(len(m.CPUInfo))
		}
	}
}
