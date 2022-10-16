package ssh

import (
	"fmt"
	"github.com/pelletier/go-toml"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
	"message"
	"sync"
	"time"
)

type SSH struct {
	sshClient     *ssh.Client
	stop          chan int
	wg            sync.WaitGroup
	cpuInfoClient CPUInfoClient
}

type CPUInfoClient struct {
	cpuInfoListener []message.CPUInfoListener
	mutex           sync.Mutex
}

var Client *SSH

const XTERM = "xterm"

func init() {
	conf, err := toml.LoadFile("./conf.toml")
	if err != nil {
		log.Fatalf("Read config file fail : %v", err)
	}
	host := conf.Get("ssh.Host").(string)
	port := conf.Get("ssh.Port").(int64)
	user := conf.Get("ssh.User").(string)

	config := &ssh.ClientConfig{
		Timeout:         time.Second,
		User:            user,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	sshType := conf.Get("ssh.Type").(string)

	if sshType == "Passwd" {
		passwd := conf.Get("ssh.Passwd").(string)
		config.Auth = []ssh.AuthMethod{ssh.Password(passwd)}
	} else if sshType == "Key" {
		keyPath := conf.Get("ssh.KeyPath").(string)
		config.Auth = []ssh.AuthMethod{publicKeyAuthFunc(keyPath)}
	}

	addr := fmt.Sprintf("%s:%d", host, port)
	sshClient, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		log.Fatalf("Create ssh client fail : %v", err)
	}
	Client = newSSH(sshClient)
}

func publicKeyAuthFunc(kPath string) ssh.AuthMethod {
	key, err := ioutil.ReadFile(kPath)
	if err != nil {
		log.Fatalf("SSH key file read fail : %v", err)
	}

	signer, err := ssh.ParsePrivateKey(key)
	if err != nil {
		log.Fatalf("SSH key signer fail : %v", err)
	}
	return ssh.PublicKeys(signer)
}

func newSSH(sshClient *ssh.Client) *SSH {
	return &SSH{
		sshClient: sshClient,
		stop:      make(chan int),
		cpuInfoClient: CPUInfoClient{
			cpuInfoListener: make([]message.CPUInfoListener, 0),
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

func (h *SSH) RegisterCPUInfoListener(listener message.CPUInfoListener) {
	h.cpuInfoClient.mutex.Lock()
	defer h.cpuInfoClient.mutex.Unlock()
	h.cpuInfoClient.cpuInfoListener = append(h.cpuInfoClient.cpuInfoListener, listener)
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
