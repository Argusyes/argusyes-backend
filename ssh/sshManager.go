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

func (m *Manager) getSSH(port int, host, user, passwd string) (*SSH, error) {
	key := generalKey(port, host, user)
	m.sshMapMutex.Lock()
	defer m.sshMapMutex.Unlock()
	ssh, ok := m.sshMap[key]

	if ok {
		return ssh, nil
	}
	ssh, err := newSSH(port, host, user, passwd)
	if err != nil {
		return nil, err
	}
	m.sshMap[key] = ssh
	ssh.startAllMonitor()
	log.Printf("ssh client create %s", ssh.Key)
	return ssh, nil
}

func (m *Manager) deleteSSH(key string) {
	m.sshMapMutex.Lock()
	delete(m.sshMap, key)
	m.sshMapMutex.Unlock()
}

func (m *Manager) RegisterAllMonitorListener(port int, host, user, passwd, wsKey string, listeners AllListener) error {
	s, err := m.getSSH(port, host, user, passwd)
	if err != nil {
		return err
	}
	s.RegisterAllListener(wsKey, listeners)
	return nil
}

func (m *Manager) ClearListener(wsKey string) {
	for _, v := range m.sshMap {
		v.RemoveAllListener(wsKey)
		if v.LenListener() == 0 {
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
	v.RemoveAllListener(wsKey)
	if v.LenListener() == 0 {
		v.Close()
		m.deleteSSH(v.Key)
		log.Printf("ssh client delete %s", v.Key)
	}
}
