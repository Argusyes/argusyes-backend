package ssh

import (
	"github.com/pkg/sftp"
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
	c.mutex.Lock()
	defer c.mutex.Unlock()
	return len(c.listener)
}

func (c *Client[M]) HasListener(key string) bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	_, ok := c.listener[key]
	return ok
}
