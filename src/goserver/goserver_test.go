package goserver

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

var httpsPort string = "10001"
var httpPort string = "10000"
var cmdName string = "sh"
var cmdArgs = []string{"tests/script.sh"}

var server *Goserver

func TestStart1(t *testing.T) {
	server = New("TEST", cmdName, cmdArgs, httpsPort)

	certFile := "tests/test.crt"
	keyFile := "tests/test.key"
	go server.NewServerTLS(certFile, keyFile)
	select {
	case <-time.After(time.Second):
		{
		}
	}
}

/**
* Encapsulate all the logic behind doing a simple request to an URL
* @string url	The endpoint to connect to
* @returns	jsonResponse the API Response in JSON format if everything it's ok, or an emtpy one in case of error.
 */
func doRequest(t *testing.T, url string) jsonResponse {
	var decoded jsonResponse

	client := getCertIgnoringClient()
	resp, err := client.Get(url)
	if err != nil {
		t.Error(err)
		return decoded
	}

	if resp == nil {
		t.Error("Empty response received")
		return decoded
	}

	defer resp.Body.Close()

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Error(err)
		return decoded
	}

	dec := json.NewDecoder(strings.NewReader(string(contents)))

	dec.Decode(&decoded)
	return decoded
}

/**
* Returns a valid http.Client object, tuned to ignore bad certificate responses from a HTTPS server
 */
func getCertIgnoringClient() http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	return http.Client{Transport: tr}
}

/**
* Veriy that we generate Valid JSON formatted messages
 */
func TestJsonMsg(t *testing.T) {

	var w *httptest.ResponseRecorder = httptest.NewRecorder()
	msg := jsonResponse{Ok: false, Result: "Testing", Error: "NO ERROR"}

	jsonMsg(w, msg)

	dec := json.NewDecoder(strings.NewReader(w.Body.String()))
	var decoded jsonResponse
	dec.Decode(&decoded)
	if decoded.Ok != false || decoded.Result != "Testing" || decoded.Error != "NO ERROR" {
		t.Error(decoded)
	}
}

/**
* Table of self-defined test requests
 */
var requestFlow = []struct {
	req    string
	ok     bool
	result interface{}
	err    interface{}
}{
	{req: "run", ok: true, result: "", err: ""},
	{req: "run", ok: true, result: "", err: ""},
	{req: "status", ok: true, result: "0", err: ""},
	{req: "stop", ok: true, result: "", err: ""},
	{req: "stop", ok: false, result: "", err: "Process does not exist"},
	{req: "status", ok: true, result: "-3", err: ""},
	{req: "run", ok: true, result: "", err: ""},
	{req: "status", ok: true, result: "0", err: ""},
}

/**
* Test the responses of our HTTPS server against a predefined flow of requests
 */
func TestTLSFakeServer(t *testing.T) {
	for _, test := range requestFlow {
		decoded := doRequest(t, "https://127.0.0.1:"+httpsPort+"/"+test.req)

		if decoded.Ok != test.ok || decoded.Result != test.result || decoded.Error != test.err {
			t.Fatal("Invalid response for command:" + test.req)
		}
	}
}

/**
* Test that we return a 404 on undetermined commands (HTTPS server)
 */
func TestTLSError404(t *testing.T) {
	client := getCertIgnoringClient()
	resp, err := client.Get("https://127.0.0.1:" + httpsPort + "/petofijo")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 404 {
		t.Fatal("Received non-200 response: %d\n", resp.StatusCode)
	}
}

func TestTearDown1(t *testing.T) {
	server.forceStop()
}

func TestStart(t *testing.T) {

	server = New("TEST", cmdName, cmdArgs, httpPort)

	go server.NewServer()
	select {
	case <-time.After(time.Second):
		{
		}
	}
}

/**
* Test the responses of our server against a predefined flow of requests
 */
func TestFakeServer(t *testing.T) {
	for _, test := range requestFlow {
		decoded := doRequest(t, "http://127.0.0.1:"+httpPort+"/"+test.req)

		if decoded.Ok != test.ok || decoded.Result != test.result || decoded.Error != test.err {
			t.Fatal("Invalid response '%v' for command '%v'", decoded, test)
		}
	}
}

/**
* Test that we return a 404 on undetermined commands
 */
func doTestError404(t *testing.T) {
	client := getCertIgnoringClient()
	resp, err := client.Get("http://127.0.0.1:" + httpPort + "/petofijo")
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != 404 {
		t.Fatal("Received non-200 response: %d\n", resp.StatusCode)
	}
}
func TestTearDown2(t *testing.T) {
	server.forceStop()
}
