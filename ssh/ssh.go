package ssh

import (
	"errors"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"message"
	"os"
	"sync"
	"time"
)

type SSH struct {
	Key                  string
	Port                 int
	Host                 string
	User                 string
	sshClient            *ssh.Client
	sftpClient           *sftp.Client
	stop                 chan int
	wg                   sync.WaitGroup
	cpuInfoClient        CPUInfoClient
	cpuPerformanceClient CPUPerformanceClient
}

type CPUInfoClient struct {
	cpuInfoListener map[string]message.CPUInfoListener
	mutex           sync.Mutex
}

type CPUPerformanceClient struct {
	cpuPerformanceListener map[string]message.CPUPerformanceListener
	mutex                  sync.Mutex
}

const XTERM = "xterm"

func newSSH(port int, host, user, passwd string) (*SSH, error) {

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
	sftpClient, err := sftp.NewClient(sshClient)
	if err != nil {
		errText := fmt.Sprintf("Create sftp client %s fail : %v", generalKey(port, host, user), err)
		log.Printf(errText)
		_ = sshClient.Close()
		return nil, errors.New(errText)
	}
	return &SSH{
		Key:        generalKey(port, host, user),
		Port:       port,
		Host:       host,
		User:       user,
		sshClient:  sshClient,
		sftpClient: sftpClient,
		stop:       make(chan int),
		cpuInfoClient: CPUInfoClient{
			cpuInfoListener: make(map[string]message.CPUInfoListener, 0),
		},
		cpuPerformanceClient: CPUPerformanceClient{
			cpuPerformanceListener: make(map[string]message.CPUPerformanceListener, 0),
		},
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
	h.wg.Add(1)
	go h.monitorCPUInfo()
	h.wg.Add(1)
	go h.monitorCPUPerformance()
}

func (h *SSH) RegisterCPUInfoListener(key string, listener message.CPUInfoListener) {
	h.cpuInfoClient.mutex.Lock()
	defer h.cpuInfoClient.mutex.Unlock()
	h.cpuInfoClient.cpuInfoListener[key] = listener
}

func (h *SSH) RemoveCPUInfoListener(key string) {
	h.cpuInfoClient.mutex.Lock()
	defer h.cpuInfoClient.mutex.Unlock()
	delete(h.cpuInfoClient.cpuInfoListener, key)
}

func (h *SSH) monitorCPUInfo() {

	where := "cpu info"
	for {
		select {
		case s, ok := <-h.stop:
			if !ok {
				h.wg.Done()
				return
			} else {
				log.Printf("Unexpect recv %d", s)
			}
		default:
			srcFile, err := h.sftpClient.OpenFile("/proc/cpuinfo", os.O_RDONLY)
			if err != nil {
				log.Printf("Read %s file fail : %v", where, err)
			}
			s, err := ioutil.ReadAll(srcFile)
			err = srcFile.Close()
			if err != nil {
				log.Printf("Close %s file fail : %v", where, err)
			}
			m := parseCPUInfoMessage(h.Port, h.Host, h.User, string(s))
			h.cpuInfoClient.mutex.Lock()
			for _, l := range h.cpuInfoClient.cpuInfoListener {
				l(m)
			}
			h.cpuInfoClient.mutex.Unlock()
		}
		time.Sleep(10 * time.Second)
	}
}

func (h *SSH) RegisterCPUPerformanceListener(key string, listener message.CPUPerformanceListener) {
	h.cpuPerformanceClient.mutex.Lock()
	defer h.cpuPerformanceClient.mutex.Unlock()
	h.cpuPerformanceClient.cpuPerformanceListener[key] = listener
}

func (h *SSH) RemoveCPUPerformanceListener(key string) {
	h.cpuPerformanceClient.mutex.Lock()
	defer h.cpuPerformanceClient.mutex.Unlock()
	delete(h.cpuPerformanceClient.cpuPerformanceListener, key)
}

func (h *SSH) monitorCPUPerformance() {

	where := "cpu performance"
	old := ""
	for {
		select {
		case s, ok := <-h.stop:
			if !ok {
				h.wg.Done()
				return
			} else {
				log.Printf("Unexpect recv %d", s)
			}
		default:
			srcFile, err := h.sftpClient.OpenFile("/proc/stat", os.O_RDONLY)
			if err != nil {
				log.Printf("Read %s file fail : %v", where, err)
			}
			s, err := ioutil.ReadAll(srcFile)
			err = srcFile.Close()
			if err != nil {
				log.Printf("Close %s file fail : %v", where, err)
			}
			if old != "" {
				m := parseCPUPerformanceMessage(h.Port, h.Host, h.User, old, string(s))
				h.cpuPerformanceClient.mutex.Lock()
				for _, l := range h.cpuPerformanceClient.cpuPerformanceListener {
					l(m)
				}
				h.cpuPerformanceClient.mutex.Unlock()
			}
			old = string(s)
		}
		time.Sleep(3 * time.Second)
	}
}
