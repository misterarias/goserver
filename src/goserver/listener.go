package goserver

import (
	"crypto/tls"
	"errors"
	"net"
	"net/http"
)

type StoppableListener struct {
	net.Listener          //Wrapped listener
	stop         chan int //Channel used only to indicate listener should shutdown
}

func getStoppableListener(server http.Server) (*StoppableListener, error) {

	l, err := net.Listen("tcp", server.Addr)
	if err != nil {
		panic(err)
	}

	if server.TLSConfig != nil {
		l = tls.NewListener(l, server.TLSConfig)
	}

	retval := &StoppableListener{}
	retval.stop = make(chan int)
	retval.Listener = l
	return retval, nil
}

var StoppedError = errors.New("Listener stopped")

func (sl *StoppableListener) Accept() (net.Conn, error) {

	for {

		newConn, err := sl.Listener.Accept()
		select {
		case <-sl.stop:
			return nil, StoppedError
		default:
			//If the channel is still open, continue as normal
		}

		if err != nil {
			netErr, ok := err.(net.Error)
			if ok && netErr.Timeout() && netErr.Temporary() {
				continue
			}
		}

		return newConn, err
	}
}

func (sl *StoppableListener) Stop() {
	close(sl.stop)
}
