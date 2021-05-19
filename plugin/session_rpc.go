package plugin

// rpcのサポートをする場合
// import (
// 	"net/rpc"

// 	"github.com/oidc-proxy-ecosystem/oidc-proxy/cache"
// )

// type RPCCacheClient struct {
// 	client *rpc.Client
// }

// var _ cache.Cache = &RPCCacheClient{}

// func (c *RPCCacheClient) Init(setting cache.CacheSetting) error {
// 	var resp interface{}
// 	return c.client.Call("Plugin.Init", setting, &resp)
// }

// func (c *RPCCacheClient) Put(key string, value string) error {
// 	var resp interface{}
// 	return c.client.Call("Plugin.Put", map[string]string{
// 		"key":   key,
// 		"value": value,
// 	}, &resp)
// }

// func (c *RPCCacheClient) Get(key string) (string, error) {
// 	var resp string
// 	err := c.client.Call("Plugin.Get", key, &resp)
// 	return resp, err
// }

// func (c *RPCCacheClient) Delete(key string) error {
// 	var resp interface{}
// 	return c.client.Call("Plugin.Delete", key, &resp)
// }

// func (c *RPCCacheClient) Close() error { return nil }

// type RPCCacheSever struct {
// 	Impl cache.Cache
// }

// func (c *RPCCacheSever) Init(setting cache.CacheSetting, resp *interface{}) {
// 	c.Impl.Init(setting)
// }
