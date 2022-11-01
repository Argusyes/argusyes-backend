package ssh

import (
	"fmt"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
	"io"
	"log"
	"sync"
)

func generalKey(port int, host, user string) string {
	return fmt.Sprintf("%s@%s:%d", user, host, port)
}

type SSHManager struct {
	sshMap map[string]*SSH
	mutex  sync.Mutex
}

var Manager = newManager()

func newManager() *SSHManager {
	return &SSHManager{
		sshMap: make(map[string]*SSH),
	}
}

func (m *SSHManager) getSSH(port int, host, user, passwd string) (*SSH, error) {
	key := generalKey(port, host, user)
	c, ok := m.sshMap[key]

	if ok {
		return c, nil
	}
	c, err := newSSH(port, host, user, passwd)
	if err != nil {
		return nil, err
	}
	m.sshMap[key] = c
	c.startAllMonitor()
	log.Printf("ssh client create %s", c.Key)
	return c, nil
}

func (m *SSHManager) deleteSSH(key string) {
	m.mutex.Lock()
	delete(m.sshMap, key)
	m.mutex.Unlock()
}

func (m *SSHManager) RegisterSSHListener(port int, host, user, passwd, wsKey string, listeners AllListener) error {
	m.mutex.Lock()
	defer m.mutex.Lock()
	s, err := m.getSSH(port, host, user, passwd)
	if err != nil {
		return err
	}
	s.RegisterSSHListener(wsKey, listeners)
	return nil
}

func (m *SSHManager) RemoveSSHListener(port int, host, user, wsKey string) {
	sshKey := generalKey(port, host, user)
	v, ok := m.sshMap[sshKey]
	if !ok {
		return
	}
	v.RemoveSSHListener(wsKey)
	if v.LenListener() == 0 {
		v.Close()
		m.deleteSSH(v.Key)
		log.Printf("ssh client delete %s", v.Key)
	}
}

func (m *SSHManager) RegisterRoughListener(port int, host string, user string, passwd string, wsKey string, listener func(m RoughMessage)) error {
	s, err := m.getSSH(port, host, user, passwd)
	m.mutex.Lock()
	defer m.mutex.Lock()
	if err != nil {
		return err
	}
	s.RegisterRoughListener(wsKey, listener)
	return nil
}

func (m *SSHManager) RemoveRoughListener(port int, host string, user string, wsKey string) {
	sshKey := generalKey(port, host, user)
	v, ok := m.sshMap[sshKey]
	if !ok {
		return
	}
	v.RemoveRoughListener(wsKey)
	if v.LenListener() == 0 {
		v.Close()
		m.deleteSSH(v.Key)
		log.Printf("ssh client delete %s", v.Key)
	}
}

func (m *SSHManager) ClearListener(wsKey string) {
	for _, v := range m.sshMap {
		v.RemoveSSHListener(wsKey)
		v.RemoveRoughListener(wsKey)
		if v.LenListener() == 0 {
			v.Close()
			m.deleteSSH(v.Key)
			log.Printf("ssh client delete %s", v.Key)
		}
	}
}

type myWriter struct {
	conn *websocket.Conn
	m    *sync.Mutex
}

func (w myWriter) Write(p []byte) (int, error) {
	w.m.Lock()
	err := w.conn.WriteMessage(websocket.TextMessage, p)
	w.m.Unlock()
	if err != nil {
		return 0, io.EOF
	}
	return len(p), nil
}

type myReader struct {
	conn  *websocket.Conn
	clean func()
}

func (w myReader) Read(p []byte) (int, error) {
	_, data, err := w.conn.ReadMessage()
	if err != nil {
		w.clean()
		return -1, io.EOF
	}
	for i, b := range data {
		p[i] = b
	}
	return len(data), nil
}

func (m *SSHManager) NewSSHClientWithConn(port int, host string, user string, passwd string, conn *websocket.Conn, mutex *sync.Mutex) (bool, error) {
	c, err := newSimpleSSH(port, host, user, passwd)
	if err != nil {
		log.Printf("new client fail : %v", err)
		return false, err
	}

	session, err := c.NewSession()
	if err != nil {
		log.Printf("new session fail : %v", err)
		return false, err
	}

	session.Stdout = myWriter{conn: conn, m: mutex}
	session.Stderr = myWriter{conn: conn, m: mutex}
	session.Stdin = myReader{conn: conn, clean: func() {
		err := session.Signal(ssh.SIGHUP)
		if err != nil {
			log.Printf("signal error: %s", err.Error())
		}
	}}

	//设置终端模式
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	// 请求伪终端
	if err = session.RequestPty("linux", 32, 160, modes); err != nil {
		log.Printf("request pty error: %s", err.Error())
		return false, err
	}

	//启动远程shell
	if err = session.Shell(); err != nil {
		log.Printf("start shell error: %s", err.Error())
		return false, err
	}
	go func() {
		//等待远程命令（终端）退出
		if err := session.Wait(); err != nil {
			log.Printf("return error: %s", err.Error())
		}
		if err := session.Close(); err != nil {
			log.Printf("close error: %s", err.Error())
		}
		if err := c.Close(); err != nil {
			log.Printf("close error: %s", err.Error())
		}
		if err := conn.Close(); err != nil {
			log.Printf("close error: %s", err.Error())
		}
	}()

	return true, nil
}
