package ssh

import (
	"errors"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"log"
	"mutexMap"
	"sync"
	"sync/atomic"
	"time"
)

type SSH struct {
	close                   int32
	empty                   int32
	Key                     string
	Port                    int
	Host                    string
	User                    string
	sshClient               *ssh.Client
	sftpClient              *sftp.Client
	stop                    chan int
	wg                      sync.WaitGroup
	parser                  Parser
	roughListener           mutexMap.MutexMap[Listener[RoughMessage]]
	cpuInfoClient           Client[CPUInfoMessage]
	cpuPerformanceClient    Client[CPUPerformanceMessage]
	memoryPerformanceClient Client[MemoryPerformanceMessage]
	uptimeClient            Client[UptimeMessage]
	loadavgClient           Client[LoadavgMessage]
	netDecClient            Client[NetDevMessage]
	netStatClient           Client[NetStatMessage]
	tempClient              Client[TempMessage]
	diskClient              Client[DiskMessage]
	processClient           Client[ProcessMessage]
}

func newSimpleSSH(port int, host, user, passwd string) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		Timeout:         time.Millisecond * 800,
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            []ssh.AuthMethod{ssh.Password(passwd)},
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	sshClient, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		errText := fmt.Sprintf("Create ssh client %s fail : %v", generalKey(port, host, user), err)
		log.Printf(errText)
		return nil, errors.New(errText)
	}
	return sshClient, nil
}

func newSSH(port int, host, user, passwd string) (*SSH, error) {

	sshClient, err := newSimpleSSH(port, host, user, passwd)
	if err != nil {
		return nil, err
	}
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		errText := fmt.Sprintf("Create sftp client %s fail : %v", generalKey(port, host, user), err)
		log.Printf(errText)
		_ = sshClient.Close()
		return nil, errors.New(errText)
	}

	return &SSH{
		Key:                     generalKey(port, host, user),
		Port:                    port,
		Host:                    host,
		User:                    user,
		sshClient:               sshClient,
		sftpClient:              sftpClient,
		stop:                    make(chan int),
		parser:                  Parser{},
		roughListener:           mutexMap.NewMutexMap[Listener[RoughMessage]](0),
		cpuInfoClient:           NewClient[CPUInfoMessage]("/proc/cpuinfo"),
		cpuPerformanceClient:    NewClient[CPUPerformanceMessage]("/proc/stat"),
		memoryPerformanceClient: NewClient[MemoryPerformanceMessage]("/proc/meminfo"),
		uptimeClient:            NewClient[UptimeMessage]("/proc/uptime"),
		loadavgClient:           NewClient[LoadavgMessage]("/proc/loadavg"),
		netDecClient:            NewClient[NetDevMessage]("/proc/net/dev"),
		netStatClient:           NewClient[NetStatMessage]("/proc/net/snmp"),
		tempClient:              NewClient[TempMessage]("/sys/class/thermal/thermal_zone0/temp"),
		diskClient:              NewClient[DiskMessage]("/proc/diskstats"),
		processClient:           NewClient[ProcessMessage](""),
	}, nil
}

func (h *SSH) Close() {
	closed := !atomic.CompareAndSwapInt32(&h.close, 0, 1)
	if !closed {
		close(h.stop)
		h.wg.Wait()
		err := h.sftpClient.Close()
		if err != nil {
			log.Printf("sftp client close fail : %v", err)
		}
		err = h.sshClient.Close()
		if err != nil {
			log.Printf("ssh client close fail : %v", err)
		}
	}
}

func (h *SSH) monitorRough(second int) {
	for ; ; time.Sleep(time.Duration(second) * time.Second) {
		select {
		case s, ok := <-h.stop:
			if !ok {
				h.wg.Done()
				return
			} else {
				log.Printf("Unexpect recv %d", s)
			}
		default:
			m := h.parser.parseRoughMessage(h.Port, h.Host, h.User)
			h.roughListener.Each(func(key string, val Listener[RoughMessage]) {
				val(*m)
			})
		}
	}
}

func (h *SSH) startAllMonitor() {
	h.wg.Add(11)
	go h.cpuInfoClient.monitor(h, h.parser.parseCPUInfoMessage, 10)
	go h.cpuPerformanceClient.monitor(h, h.parser.parseCPUPerformanceMessage, 2)
	go h.memoryPerformanceClient.monitor(h, h.parser.parseMemoryPerformanceMessage, 2)
	go h.uptimeClient.monitor(h, h.parser.parseUptimeMessage, 2)
	go h.loadavgClient.monitor(h, h.parser.parseLoadavgMessage, 2)
	go h.netDecClient.monitor(h, h.parser.parseNetDevMessage, 2)
	go h.netStatClient.monitor(h, h.parser.parseNetStatMessage, 2)
	go h.tempClient.monitor(h, h.parser.parseTempMessage, 2)
	go h.diskClient.monitor(h, h.parser.parseDiskMessage, 2)
	go h.processClient.monitor(h, h.parser.parseProcessMessage, 5)
	go h.monitorRough(2)
}

func (h *SSH) RegisterRoughListener(key string, listener Listener[RoughMessage]) {
	atomic.AddInt32(&h.empty, 1)
	h.roughListener.Set(key, listener)
}

func (h *SSH) RemoveRoughListener(key string) {
	atomic.AddInt32(&h.empty, -1)
	h.roughListener.Remove(key)
}

func (h *SSH) RegisterSSHListener(key string, listeners AllListener) {
	atomic.AddInt32(&h.empty, 1)
	if listeners.CPUInfoListener != nil {
		h.cpuInfoClient.RegisterHandler(key, listeners.CPUInfoListener)
	}
	if listeners.CPUPerformanceListener != nil {
		h.cpuPerformanceClient.RegisterHandler(key, listeners.CPUPerformanceListener)
	}
	if listeners.MemoryPerformanceListener != nil {
		h.memoryPerformanceClient.RegisterHandler(key, listeners.MemoryPerformanceListener)
	}
	if listeners.UptimeListener != nil {
		h.uptimeClient.RegisterHandler(key, listeners.UptimeListener)
	}
	if listeners.LoadavgListener != nil {
		h.loadavgClient.RegisterHandler(key, listeners.LoadavgListener)
	}
	if listeners.NetDevListener != nil {
		h.netDecClient.RegisterHandler(key, listeners.NetDevListener)
	}
	if listeners.NetStatListener != nil {
		h.netStatClient.RegisterHandler(key, listeners.NetStatListener)
	}
	if listeners.TempListener != nil {
		h.tempClient.RegisterHandler(key, listeners.TempListener)
	}
	if listeners.DiskListener != nil {
		h.diskClient.RegisterHandler(key, listeners.DiskListener)
	}
	if listeners.ProcessListener != nil {
		h.processClient.RegisterHandler(key, listeners.ProcessListener)
	}
}

func (h *SSH) RemoveSSHListener(key string) {
	atomic.AddInt32(&h.empty, -1)
	h.cpuInfoClient.RemoveHandler(key)
	h.cpuPerformanceClient.RemoveHandler(key)
	h.memoryPerformanceClient.RemoveHandler(key)
	h.uptimeClient.RemoveHandler(key)
	h.loadavgClient.RemoveHandler(key)
	h.netDecClient.RemoveHandler(key)
	h.netStatClient.RemoveHandler(key)
	h.tempClient.RemoveHandler(key)
	h.diskClient.RemoveHandler(key)
}

func (h *SSH) Empty() bool {
	return h.empty == 0
}
