package goserver

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
)

/** The command currently being watched */
var watchedCmd = &exec.Cmd{}

/** These are the defined logger functions that we have available */
var (
	Trace   *log.Logger
	Info    *log.Logger
	Warning *log.Logger
	Error   *log.Logger
)

/**
 * This type serves as this class's member structure
 */
type Goserver struct {
	/** Server class that holds this wrapper's functionality */
	server http.Server
	/** Global reference to a signal channel to be able to stop the server */
	stop chan os.Signal
}

func New(name string, watchedProcessName string, watchedProcessArgs []string, _port string) *Goserver {
	setupLogging(name)

	// I make this a global since a) its immutable and b) it needs to be accessible from within the handlers
	watchedCmd = exec.Command(watchedProcessName, watchedProcessArgs...)
	server := GetServer(_port)
	stop := make(chan os.Signal, 1)

	return &Goserver{server, stop}
}

func GetServer(_port string) http.Server {
	return http.Server{
		Addr:     ":" + _port,
		Handler:  &myHandler{},
		ErrorLog: Error,
	}
}

func getStatus() int {
	if watchedCmd.Process == nil {
		return -3
	}
	return 0
}

/**
* There are three possible places to log that are useful, depending on what we want to achieve.
* They are sent as the first parameter of the log.New(...) call: os.Stderr, os.Stdout, and a file
 */
func setupLogging(module string) {
	Trace = log.New(os.Stdout, "[TRACE] "+module+": ", log.Ldate|log.Ltime|log.Lshortfile)
	Info = log.New(os.Stdout, "[INFO] "+module+": ", log.Ldate|log.Ltime|log.Lshortfile)
	Warning = log.New(os.Stdout, "[WARNING] "+module+": ", log.Ldate|log.Ltime|log.Lshortfile)
	Error = log.New(os.Stderr, "[ERROR] "+module+": ", log.Ldate|log.Ltime|log.Lshortfile)

	Info.Printf("Logging ready")
}

func setupTLS(server *http.Server, certFile string, keyFile string) {
	config := &tls.Config{}
	if config.NextProtos == nil {
		config.NextProtos = []string{"http/1.1"}
	}

	var err error
	config.Certificates = make([]tls.Certificate, 1)
	config.Certificates[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		Error.Printf("Error loading certificates: %v", err)
		os.Exit(2)
	}
	server.TLSConfig = config
}

func (s *Goserver) Serve() {
	sl, err := getStoppableListener(s.server)
	if err != nil {
		Error.Printf("%v", err)
		os.Exit(2)
	}

	signal.Notify(s.stop, syscall.SIGINT)
	go func() {
		var wg sync.WaitGroup
		wg.Add(1)
		defer wg.Done()
		s.server.Serve(sl)
	}()

	Info.Printf("Serving HTTPS\n")
	select {
	case signal := <-s.stop:
		Info.Printf("Got signal:%v\n", signal)
		p := watchedCmd.Process
		if p != nil {
			p.Signal(syscall.SIGTERM)
		}

	}
	Info.Printf("HTTPS Served stopped gracefully\n")
	sl.Stop()
}

func (s *Goserver) NewServer() {
	if getStatus() == 0 {
		Error.Printf("There is a process being watched, you must stop it first")
		os.Exit(1)
	}
	s.Serve()
}

func (s *Goserver) NewServerTLS(_certFile string, _keyFile string) {
	if getStatus() == 0 {
		Error.Printf("There is a process being watched, you must stop it first")
		os.Exit(1)
	}

	setupTLS(&s.server, _certFile, _keyFile)
	s.Serve()
}

func (s *Goserver) forceStop() {
	Trace.Printf("Stopping abruptly\n")
	s.stop <- syscall.SIGINT
}
