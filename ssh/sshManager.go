package ssh

import (
	"fmt"
	"log"
	"sync"
)

func GeneralKey(port int, host, user string) string {
	return fmt.Sprintf("%s@%s:%d", user, host, port)
}

type Manager struct {
	sshMap      map[string]*SSH
	sshMapMutex sync.Mutex
}

var SSHManager = newManager()

func newManager() *Manager {
	return &Manager{
		sshMap: make(map[string]*SSH),
	}
}

func (m *Manager) GetSSH(port int, host, user, passwd string) *SSH {
	key := GeneralKey(port, host, user)
	m.sshMapMutex.Lock()
	defer m.sshMapMutex.Unlock()
	ssh, ok := m.sshMap[key]

	if ok {
		return ssh
	}
	ssh = NewSSH(port, host, user, passwd)
	m.sshMap[key] = ssh
	ssh.StartAllMonitor()
	log.Printf("ssh client create %s", ssh.Key)
	return ssh
}

func (m *Manager) RemoveCPUInfoListener(key string) {
	for _, v := range m.sshMap {
		v.RemoveCPUInfoListener(key)
		if len(v.cpuInfoClient.cpuInfoListener) == 0 {
			m.sshMapMutex.Lock()
			delete(m.sshMap, v.Key)
			m.sshMapMutex.Unlock()
			v.Close()
			log.Printf("ssh client delete %s", v.Key)
		}
	}
}

func (m *Manager) RemoveSSHCPUInfoListener(sshKeys []string, key string) {
	for _, sshKey := range sshKeys {
		v, ok := m.sshMap[sshKey]
		if !ok {
			return
		}
		v.RemoveCPUInfoListener(key)
		if len(v.cpuInfoClient.cpuInfoListener) == 0 {
			m.sshMapMutex.Lock()
			delete(m.sshMap, v.Key)
			m.sshMapMutex.Unlock()
			v.Close()
			log.Printf("ssh client delete %s", v.Key)
		}
	}
}
