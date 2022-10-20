package ssh

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"message"
	"sync"
	"time"
)

type SSH struct {
	Key           string
	sshClient     *ssh.Client
	stop          chan int
	wg            sync.WaitGroup
	cpuInfoClient CPUInfoClient
}

type CPUInfoClient struct {
	cpuInfoListener map[string]message.CPUInfoListener
	mutex           sync.Mutex
}

const XTERM = "xterm"

func generalKey(port int, host, user string) string {
	return fmt.Sprintf("%s@%s:%d", user, host, port)
}

func NewSSH(port int, host, user, passwd string) *SSH {

	config := &ssh.ClientConfig{
		Timeout:         time.Second,
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth:            []ssh.AuthMethod{ssh.Password(passwd)},
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	sshClient, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		log.Fatalf("Create ssh client fail : %v", err)
	}
	return &SSH{
		Key:       generalKey(port, host, user),
		sshClient: sshClient,
		stop:      make(chan int),
		cpuInfoClient: CPUInfoClient{
			cpuInfoListener: make(map[string]message.CPUInfoListener, 0),
		},
	}
}

func (h *SSH) Close() {
	close(h.stop)
	err := h.sshClient.Close()
	if err != nil {
		log.Fatalf("SSH cleint close fail : %v", err)
	}
}

func (h *SSH) newSession(where string) (session *ssh.Session) {
	session, err := h.sshClient.NewSession()
	if err != nil {
		log.Fatal(err)
	}

	modes := ssh.TerminalModes{
		ssh.ECHO: 0,
	}

	if err := session.RequestPty(XTERM, 50, 80, modes); err != nil {
		log.Fatalf("Request %s session pty fail : %v", where, err)
	}

	return
}

func closeSession(where string, session *ssh.Session) {
	err := session.Close()
	if err != nil && err != io.EOF {
		log.Fatalf("Close %s session fail : %v", where, err)
	}
}

func (h *SSH) StartAllMonitor() {
	h.wg.Add(1)
	go h.monitorCPUInfo()
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
		session := h.newSession(where)
		select {
		case s, ok := <-h.stop:
			if !ok {
				h.wg.Done()
				closeSession(where, session)
				return
			} else {
				log.Printf("Unexpect recv %d", s)
			}
		default:
			s, err := session.Output("cat /proc/cpuinfo")
			if err != nil {
				log.Printf("Run %s command fail : %v", where, err)
			}
			m := parseCPUInfoMessage(string(s))
			h.cpuInfoClient.mutex.Lock()
			for _, l := range h.cpuInfoClient.cpuInfoListener {
				l(m)
			}
			h.cpuInfoClient.mutex.Unlock()
		}
		closeSession(where, session)
		time.Sleep(time.Second)
	}
}
