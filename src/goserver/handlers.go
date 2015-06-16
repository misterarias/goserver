package goserver

import (
	//	"fmt"
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"
	"syscall"
	"time"
)

/** Simple JSON message structure */
type jsonResponse struct {
	Ok     bool   `json:"ok"`
	Result string `json:"result"`
	Error  string `json:"error"`
}

func jsonMsg(w http.ResponseWriter, msg jsonResponse) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	enc := json.NewEncoder(w)
	enc.Encode(msg)
}

func jsonError(w http.ResponseWriter, s string) {
	errMsg := jsonResponse{Ok: false, Error: s}
	jsonMsg(w, errMsg)
}

func runHandler(w http.ResponseWriter, r *http.Request) {

	var status int = getStatus()
	if status == 0 {
		returnMsg := jsonResponse{Ok: true}
		jsonMsg(w, returnMsg)
		return
	}

	pid := -1
	done := make(chan int, 1)
	go func() {
		// Redirect process output to our logs
		watchedCmd.Stdout = os.Stdout
		watchedCmd.Stderr = os.Stderr

		/**
		* Spawn the process with a different Groip PID, so we can kill them all
		 */
		watchedCmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

		if err := watchedCmd.Start(); err != nil {
			Error.Printf("Unable to launch process %s: %v", strings.Join(watchedCmd.Args, " "), err)
			jsonError(w, err.Error())
			return
		}
		pid = watchedCmd.Process.Pid
		done <- 1
		Info.Printf("Launched process %s with pid %d", strings.Join(watchedCmd.Args, " "), pid)

		watchedCmd.Wait()
	}()

	select {
	case <-done:
		{
			if pid == -1 {
				err := "Process died after spawn, check logs"
				Error.Println(err)
				jsonError(w, err)
			} else {
				returnMsg := jsonResponse{Ok: true}
				jsonMsg(w, returnMsg)
			}
		}
	case <-time.After(time.Second):
		{
			// Errors have been sent already
			deleteProcess()
		}
	}

	return
}

func stopHandler(w http.ResponseWriter, r *http.Request) {
	if getStatus() == 0 {
		//p, _ := os.FindProcess(pid)
		deleteProcess()
		p := watchedCmd.Process
		if p != nil {
			//		Error.Printf("[stop] failed to kill process")
			jsonError(w, "Failed to kill process")
		} else {
			returnMsg := jsonResponse{Ok: true}
			jsonMsg(w, returnMsg)
		}
		return
	}
	//	Error.Printf("[stop] Process does not exist")
	jsonError(w, "Process does not exist")
	return
}

func statusHandler(w http.ResponseWriter, r *http.Request) {
	status := getStatus()
	returnMsg := jsonResponse{Ok: true, Result: strconv.Itoa(status)}
	jsonMsg(w, returnMsg)
}

/**
* Used inside the ServeHTTP override to decide what to do with requests
 */
var mux = map[string]func(http.ResponseWriter, *http.Request){
	"/run":    runHandler,
	"/stop":   stopHandler,
	"/status": statusHandler,
}

/* Interface-defined type + minimum method needed in order to be used below */
type myHandler struct{}

func (*myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h, ok := mux[r.URL.String()]; ok {
		// allow cross domain AJAX requests
		w.Header().Set("Access-Control-Allow-Origin", "*")
		h(w, r)
	} else {
		http.NotFound(w, r)
	}
	return
}

func deleteProcess() {
	if watchedCmd.Process != nil {
		// Kill the whole spawned group
		pgid, err := syscall.Getpgid(watchedCmd.Process.Pid)
		if err == nil {
			syscall.Kill(-pgid, 15)
		}

		watchedCmd.Process.Kill()
		//watchedCmd.Process.Signal(syscall.SIGKILL)
	}
	watchedCmd.Process = nil
}
