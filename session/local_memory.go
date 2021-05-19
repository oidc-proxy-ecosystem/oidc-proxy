package session

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"sync"
	"time"

	"github.com/oidc-proxy-ecosystem/oidc-proxy/logger"
)

type item struct {
	Value   string `json:"value"`
	Expires int64  `json:"expires"`
}

func newItem(value string, expiredTime int64) *item {
	return &item{
		Value:   value,
		Expires: expiredTime,
	}
}

func (i *item) ToJson() string {
	buf, _ := json.Marshal(i)
	return string(buf)
}

func (i *item) Expired(time int64) bool {
	if i.Expires == 0 {
		return true
	}
	return time > i.Expires
}

type memorySession struct {
	items  map[string]*item
	mu     sync.Mutex
	prefix string
	ttl    int
	log    logger.ILogger
}

var _ Session = &memorySession{}

func (c *memorySession) Get(ctx context.Context, originalKey string) (string, error) {
	c.mu.Lock()
	key := path.Join(c.prefix, originalKey)
	var s string = ""
	if v, ok := c.items[key]; ok {
		s = v.Value
	}
	c.log.Debug(fmt.Sprintf("[GET] %s:%s", key, s))
	c.mu.Unlock()
	return s, nil
}
func (c *memorySession) Put(ctx context.Context, originalKey string, value string) error {
	c.mu.Lock()
	expiredTime := time.Now().Add(time.Duration(c.ttl) * time.Minute).UnixNano()
	key := path.Join(c.prefix, originalKey)
	c.items[key] = newItem(value, expiredTime)
	c.log.Debug(fmt.Sprintf("[PUT] %s:%s", key, value))
	c.mu.Unlock()
	return nil
}
func (c *memorySession) Delete(ctx context.Context, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	key = path.Join(c.prefix, key)
	c.log.Debug(fmt.Sprintf("[DEL] %s", key))
	delete(c.items, key)
	return nil
}
func (c *memorySession) Close(ctx context.Context) error {
	return nil
}
func (c *memorySession) Init(ctx context.Context, setting map[string]interface{}) error {
	c.log.Info(fmt.Sprintf("%#v", setting))
	return nil
}

func NewLocalMemory() *memorySession {
	c := &memorySession{
		items:  make(map[string]*item),
		mu:     sync.Mutex{},
		prefix: "memory",
		ttl:    90,
		log:    logger.New(os.Stderr, logger.Debug, logger.FormatLong, logger.FormatDatetime),
	}
	go func() {
		t := time.NewTicker(time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				c.mu.Lock()
				for k, v := range c.items {
					if v.Expired(time.Now().UnixNano()) {
						delete(c.items, k)
					}
				}
				c.mu.Unlock()
			}
		}
	}()
	return c
}
