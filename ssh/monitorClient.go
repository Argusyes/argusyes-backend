package ssh

import (
	"github.com/pkg/sftp"
	"io/ioutil"
	"log"
	"mutexMap"
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
			if c.where != "" {
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
				context.newS = string(newS)
				context.newTime = time.Now()
				err = srcFile.Close()
				if err != nil {
					log.Printf("Close %s file fail : %v", c.where, err)
				}
			}
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
