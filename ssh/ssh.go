package ssh

import (
	"errors"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

type MonitorContext struct {
	client  *sftp.Client
	port    int
	host    string
	user    string
	oldS    string
	newS    string
	oldTime time.Time
	newTime time.Time
}

type Client[M any] struct {
	listener map[string]Listener[M]
	mutex    sync.Mutex
	where    string
}

func NewClient[M any](where string) Client[M] {
	return Client[M]{
		listener: make(map[string]Listener[M], 0),
		where:    where,
	}
}

func (c *Client[M]) Handler(m M) {
	c.mutex.Lock()
	for _, l := range c.listener {
		l(m)
	}
	c.mutex.Unlock()
}

func (c *Client[M]) RegisterHandler(key string, listener Listener[M]) {
	c.mutex.Lock()
	c.listener[key] = listener
	c.mutex.Unlock()
}

func (c *Client[M]) RemoveHandler(key string) {
	c.mutex.Lock()
	delete(c.listener, key)
	c.mutex.Unlock()
}

func (c *Client[M]) monitor(h *SSH, f func(context MonitorContext) *M, second int) {
	oldS := ""
	oldTime := time.Now()
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
			srcFile, err := h.sftpClient.OpenFile(c.where, os.O_RDONLY)
			if err != nil {
				log.Printf("Read %s file fail : %v", c.where, err)
				continue
			}
			newS, err := ioutil.ReadAll(srcFile)
			if err != nil {
				log.Printf("Read %s file fail : %v", c.where, err)
				continue
			}
			newTime := time.Now()
			err = srcFile.Close()
			if err != nil {
				log.Printf("Close %s file fail : %v", c.where, err)
			}
			m := f(MonitorContext{
				client:  h.sftpClient,
				port:    h.Port,
				host:    h.Host,
				user:    h.User,
				oldS:    oldS,
				newS:    string(newS),
				oldTime: oldTime,
				newTime: newTime,
			})
			if m != nil {
				c.Handler(*m)
			}
			oldS = string(newS)
			oldTime = newTime
		}
	}
}

func (c *Client[M]) LenListener() int {
	return len(c.listener)
}

type SSH struct {
	Key                     string
	Port                    int
	Host                    string
	User                    string
	sshClient               *ssh.Client
	sftpClient              *sftp.Client
	stop                    chan int
	wg                      sync.WaitGroup
	parser                  Parser
	cpuInfoClient           Client[CPUInfoMessage]
	cpuPerformanceClient    Client[CPUPerformanceMessage]
	memoryPerformanceClient Client[MemoryPerformanceMessage]
	uptimeClient            Client[UptimeMessage]
	loadavgClient           Client[LoadavgMessage]
	netDecClient            Client[NetDevMessage]
	netStatClient           Client[NetStatMessage]
	tempClient              Client[TempMessage]
	diskClient              Client[DiskMessage]
}

func newSimpleSSH(port int, host, user, passwd string) (*ssh.Client, error) {
	config := &ssh.ClientConfig{
		Timeout:         time.Second,
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
		cpuInfoClient:           NewClient[CPUInfoMessage]("/proc/cpuinfo"),
		cpuPerformanceClient:    NewClient[CPUPerformanceMessage]("/proc/stat"),
		memoryPerformanceClient: NewClient[MemoryPerformanceMessage]("/proc/meminfo"),
		uptimeClient:            NewClient[UptimeMessage]("/proc/uptime"),
		loadavgClient:           NewClient[LoadavgMessage]("/proc/loadavg"),
		netDecClient:            NewClient[NetDevMessage]("/proc/net/dev"),
		netStatClient:           NewClient[NetStatMessage]("/proc/net/snmp"),
		tempClient:              NewClient[TempMessage]("/sys/class/thermal/thermal_zone0/temp"),
		diskClient:              NewClient[DiskMessage]("/proc/diskstats"),
	}, nil
}

func (h *SSH) Close() {
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

func (h *SSH) startAllMonitor() {
	h.wg.Add(9)
	go h.cpuInfoClient.monitor(h, h.parser.parseCPUInfoMessage, 10)
	go h.cpuPerformanceClient.monitor(h, h.parser.parseCPUPerformanceMessage, 2)
	go h.memoryPerformanceClient.monitor(h, h.parser.parseMemoryPerformanceMessage, 2)
	go h.uptimeClient.monitor(h, h.parser.parseUptimeMessage, 2)
	go h.loadavgClient.monitor(h, h.parser.parseLoadavgMessage, 2)
	go h.netDecClient.monitor(h, h.parser.parseNetDevMessage, 2)
	go h.netStatClient.monitor(h, h.parser.parseNetStatMessage, 2)
	go h.tempClient.monitor(h, h.parser.parseTempMessage, 2)
	go h.diskClient.monitor(h, h.parser.parseDiskMessage, 2)
}

func (h *SSH) RegisterAllListener(key string, listeners AllListener) {
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
}

func (h *SSH) RemoveAllListener(key string) {
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

func (h *SSH) LenListener() int {
	return h.cpuInfoClient.LenListener()
}
