package ssh

import (
	"fmt"
	"log"
	"sync"
)

func generalKey(port int, host, user string) string {
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

func (m *Manager) getSSH(port int, host, user, passwd string) *SSH {
	key := generalKey(port, host, user)
	m.sshMapMutex.Lock()
	defer m.sshMapMutex.Unlock()
	ssh, ok := m.sshMap[key]

	if ok {
		return ssh
	}
	ssh = newSSH(port, host, user, passwd)
	m.sshMap[key] = ssh
	ssh.startAllMonitor()
	log.Printf("ssh client create %s", ssh.Key)
	return ssh
}

func (m *Manager) deleteSSH(key string) {
	m.sshMapMutex.Lock()
	delete(m.sshMap, key)
	m.sshMapMutex.Unlock()
}

func (m *Manager) RegisterAllMonitorListener(port int, host, user, passwd, wsKey string, listener *Listener) {
	s := m.getSSH(port, host, user, passwd)
	s.RegisterCPUInfoListener(wsKey, listener.CPUInfoListener)
}

func (m *Manager) ClearListener(wsKey string) {
	for _, v := range m.sshMap {
		v.RemoveCPUInfoListener(wsKey)
		if len(v.cpuInfoClient.cpuInfoListener) == 0 {
			v.Close()
			m.deleteSSH(v.Key)
			log.Printf("ssh client delete %s", v.Key)
		}
	}
}

func (m *Manager) RemoveSSHListener(port int, host, user, wsKey string) {
	sshKey := generalKey(port, host, user)
	v, ok := m.sshMap[sshKey]
	if !ok {
		return
	}
	v.RemoveCPUInfoListener(wsKey)
	if len(v.cpuInfoClient.cpuInfoListener) == 0 {
		m.deleteSSH(sshKey)
		v.Close()
		log.Printf("ssh client delete %s", v.Key)
	}
}
