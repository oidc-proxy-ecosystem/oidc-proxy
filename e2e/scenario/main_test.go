package scenario_test

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/oidc-proxy-ecosystem/oidc-proxy/config"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/e2e/framework"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/e2e/utils"
	"github.com/oidc-proxy-ecosystem/oidc-proxy/routes"
	"github.com/stretchr/testify/assert"
	"golang.org/x/net/websocket"
)

const applicationYAML = `
logging:
  level: info
  logformat: "standard"
  timeformat: "datetime"
port: 8888
ssl_certificate: ""
ssl_certificate_key: ""
servers:
  - server_name: 127.0.0.1
    port: 8080
    cookie_name: session
    login: "/oauth2/login"
    callback: "/oauth2/callback"
    logout: "/oauth2/logout"
    redirect: true
    oidc:
      provider: http://127.0.0.1
      client_id: "oidc-proxy-ecosystem-provider"
      client_secret: "test"
      logout: ""
      redirect_url: http://127.0.0.1:8888/oauth2/callback
      scopes:
        - email
        - openid
        - offline_access
        - profile
    locations:
      - proxy_pass: http://127.0.0.1, http://127.0.0.1
        urls:
          - path: /
            token: "id_token"
            type: Bearer
      - proxy_pass: http://127.0.0.1
        urls:
          - path: /ws/echo
            token: "id_token"
            type: Bearer
    logging:
      level: info
      logformat: "standard"
      timeformat: "datetime"
    session:
      name: "memory"
      plugin: false
      codecs: 
        - "something-very-secret"
      args:
        endpoints: 
          - "http://kube-oidc-proxy-etcd:2379"
        prefix: "/memory"
        filename: memory.log
        loglevel: debug
`

func buildServer(port int) *http.Server {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/hello", func(rw http.ResponseWriter, r *http.Request) {
		log.Println(port)
		rw.Write([]byte("hello, world"))
	})
	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
	return s
}

type mockTransport struct {
	rt     func(*http.Request) (*http.Response, error)
	cookie map[string][]*http.Cookie
}

func (m *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.cookie == nil {
		m.cookie = make(map[string][]*http.Cookie)
	}
	if len(req.Cookies()) == 0 {
		if cookies, ok := m.cookie[req.URL.Host]; ok {
			for _, cookie := range cookies {
				req.AddCookie(cookie)
			}
		}
	}
	resp, err := m.rt(req)
	if err == nil {
		m.cookie[req.URL.Host] = resp.Cookies()
	}
	return resp, err
}

func EchoHandler(ws *websocket.Conn) {
	io.Copy(ws, ws)
}

func buildWebSockerServer(port int) *http.Server {
	mux := http.NewServeMux()
	mux.Handle("/ws/echo", websocket.Handler(EchoHandler))
	s := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}
	return s
}

func TestProxy(t *testing.T) {
	filename := "/tmp/application.yaml"
	file, err := os.Create(filename)
	if !assert.NoError(t, err) {
		return
	}
	defer file.Close()
	file.WriteString(applicationYAML)
	file.Close()
	idpConnsClose := make(chan struct{})
	idleConnsClose := make(chan struct{})
	proxyConnsClose := make(chan struct{})
	conf, err := config.New(filename)
	if !assert.NoError(t, err) {
		return
	}

	resourcePort, _ := utils.FindPort()
	server := buildServer(resourcePort)
	l, err := net.Listen("tcp", server.Addr)
	if !assert.NoError(t, err) {
		return
	}
	go func() {
		if err := server.Serve(l); err != http.ErrServerClosed {
			assert.Error(t, err)
		}
		close(idleConnsClose)
	}()

	idleConnsClose2 := make(chan struct{})
	resourcePort2, _ := utils.FindPort()
	server2 := buildServer(resourcePort2)
	l2, err := net.Listen("tcp", server2.Addr)
	if !assert.NoError(t, err) {
		return
	}
	go func() {
		if err := server2.Serve(l2); err != http.ErrServerClosed {
			assert.Error(t, err)
		}
		close(idleConnsClose2)
	}()

	idleConnsClose3 := make(chan struct{})
	resourcePort3, _ := utils.FindPort()
	server3 := buildWebSockerServer(resourcePort3)
	l3, err := net.Listen("tcp", server3.Addr)
	if !assert.NoError(t, err) {
		return
	}
	go func() {
		if err := server3.Serve(l3); err != http.ErrServerClosed {
			assert.Error(t, err)
		}
		close(idleConnsClose3)
	}()

	idp, err := framework.NewIdpServer(idpConnsClose)
	if !assert.NoError(t, err) {
		return
	}
	proxyPort, _ := utils.FindPort()
	srv, err := routes.New(func() config.Servers {
		confSrv := *conf.Servers[0]
		confSrv.Oidc.Provider = idp.Issuer
		confSrv.Locations[0].ProxyPass = fmt.Sprintf("http://127.0.0.1:%d,http://127.0.0.1:%d", resourcePort, resourcePort2)
		confSrv.Locations[1].ProxyPass = fmt.Sprintf("http://127.0.0.1:%d", resourcePort3)
		confSrv.Oidc.RedirectUrl = fmt.Sprintf("http://127.0.0.1:%d/oauth2/callback", proxyPort)
		return confSrv
	})
	if !assert.NoError(t, err) {
		return
	}
	proxyServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", proxyPort),
		Handler: srv,
	}
	proxyListen, _ := net.Listen("tcp", proxyServer.Addr)
	proxyURL := func(url string) string {
		return fmt.Sprintf("http://127.0.0.1:%d/%s", proxyPort, url)
	}
	go func() {
		if err := proxyServer.Serve(proxyListen); err != http.ErrServerClosed {
			assert.NoError(t, err)
		}
		close(proxyConnsClose)
	}()
	mocktransport := &mockTransport{
		rt:     http.DefaultTransport.RoundTrip,
		cookie: make(map[string][]*http.Cookie),
	}
	defer func() {
		idp.Shutdown(context.Background())
		<-idpConnsClose
		server.Shutdown(context.Background())
		<-idleConnsClose
		server2.Shutdown(context.Background())
		<-idleConnsClose2
		server3.Shutdown(context.Background())
		<-idleConnsClose3
		proxyServer.Shutdown(context.Background())
		<-proxyConnsClose
	}()
	tests := []struct {
		name string
		fn   func(t *testing.T)
	}{
		{
			name: "authorize",
			fn: func(t *testing.T) {
				client := &http.Client{
					Transport: mocktransport,
				}
				res, err := client.Get(proxyURL("oauth2/login"))
				if !assert.NoError(t, err) {
					return
				}
				res.Body.Close()
			},
		},
		{
			name: "access /api/v1/hello",
			fn: func(t *testing.T) {
				for i := 0; i < 10; i++ {
					client := &http.Client{
						Transport: mocktransport,
					}
					res, err := client.Get(proxyURL("/api/v1/hello"))
					if !assert.NoError(t, err) {
						return
					}
					defer res.Body.Close()
					buf, _ := ioutil.ReadAll(res.Body)
					assert.Equal(t, "hello, world", string(buf))
				}
			},
		},
		{
			name: "websocket proxy",
			fn: func(t *testing.T) {
				type EchoMsg struct {
					Msg string
					ID  int32
				}
				expectedMsg := EchoMsg{
					Msg: "test",
					ID:  1,
				}
				sendMsg := func(ws *websocket.Conn, msg string, id int32) {
					var sndMsg = EchoMsg{msg, id}
					websocket.JSON.Send(ws, sndMsg)
				}

				receiveMsg := func(ws *websocket.Conn) {
					var rcvMsg EchoMsg
					for {
						websocket.JSON.Receive(ws, &rcvMsg)
						assert.Equal(t, expectedMsg, rcvMsg)
					}
				}

				url := fmt.Sprintf("ws://127.0.0.1:%d/ws/echo", resourcePort3)
				origin := fmt.Sprintf("http://127.0.0.1:%d/", resourcePort3)
				ws, err := websocket.Dial(url, "", origin)
				if !assert.NoError(t, err) {
					return
				}
				defer ws.Close()
				go receiveMsg(ws)
				sendMsg(ws, "test", 1)
				time.Sleep(1 * time.Second)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.fn)
	}
}
