package ssh

import "sync"

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
	key := generalKey(port, host, user)
	m.sshMapMutex.Lock()
	defer m.sshMapMutex.Unlock()
	ssh, ok := m.sshMap[key]

	if ok {
		return ssh
	}
	ssh = NewSSH(port, host, user, passwd)
	m.sshMap[key] = ssh
	ssh.StartAllMonitor()
	return ssh
}

func (m *Manager) RemoveCPUInfoListener(key string) {
	for _, v := range m.sshMap {
		v.RemoveCPUInfoListener(key)
	}
}
