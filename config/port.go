package config

import "net"

func findPort() (int, error) {
	l, err := net.Listen("tcp", ":0")
	if err != nil {
		return -1, err
	}
	addr := l.Addr().(*net.TCPAddr)
	if err := l.Close(); err != nil {
		return -1, err
	}
	return addr.Port, nil
}
