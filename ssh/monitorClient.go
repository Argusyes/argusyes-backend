package ssh

import (
	"github.com/pkg/sftp"
	"log"
	"mutexMap"
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
	where   string
}

type Client[M any] struct {
	listener mutexMap.MutexMap[Listener[M]]
	mutex    sync.Mutex
	where    string
}

func NewClient[M any](where string) Client[M] {
	return Client[M]{
		listener: mutexMap.NewMutexMap[Listener[M]](0),
		where:    where,
	}
}

func (c *Client[M]) Handler(m M) {
	c.listener.Each(func(key string, val Listener[M]) {
		val(m)
	})
}

func (c *Client[M]) RegisterHandler(key string, listener Listener[M]) {
	c.listener.Set(key, listener)
}

func (c *Client[M]) RemoveHandler(key string) {
	c.listener.Remove(key)
}

func (c *Client[M]) hasHandler(key string) bool {
	return c.listener.Has(key)
}

func (c *Client[M]) monitor(h *SSH, f func(context *MonitorContext) *M, second int) {
	context := &MonitorContext{
		client:  h.sftpClient,
		port:    h.Port,
		host:    h.Host,
		user:    h.User,
		oldS:    "",
		newS:    "",
		oldTime: time.Now(),
		newTime: time.Now(),
		where:   c.where,
	}
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
			m := f(context)
			if m != nil {
				c.Handler(*m)
			}
		}
	}
}

func (c *Client[M]) LenListener() int {
	return c.listener.Len()
}

func (c *Client[M]) HasListener(key string) bool {
	return c.listener.Has(key)
}
