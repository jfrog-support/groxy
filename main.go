package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"regexp"
	"strings"

	"github.com/jfrog-support/groxy/utils"
	//"github.com/fatih/color"
	//"errors"
)

var (
	// Trace - trace logger
	Trace *log.Logger
	// Info - info logger
	Info *log.Logger
	// Warning - warning logger
	Warning *log.Logger
	// Error - error logger
	Error *log.Logger
)

const (
	dockerContextPath = "/artifactory/api/docker/"
)

// Conf Configuration
type Conf struct {
	ArtifactoryHost string
	DefaultUIPort   string
	DefaultV1Port   string
	DefaultV2Port   string
	V1RepoKey       string
	V2RepoKey       string
}

type prox struct {
	target        *url.URL
	proxy         *httputil.ReverseProxy
	routePatterns []*regexp.Regexp
	path          string
}

// used to capture response RoundTrip information
type myTransport struct {
}

type uiHandler struct {
	ArtifactoryHost string
}

type v1Handler struct {
	ArtifactoryHost string
	RepoKey         string
}

type v2Handler struct {
	ArtifactoryHost string
	RepoKey         string
}

// TODO: sub-type the handlers and make this function accept a handler
func newSingleHostReverseProxy(target *url.URL, path string, repoKey string) *httputil.ReverseProxy {
	director := func(req *http.Request) {
		req.URL.Scheme = target.Scheme
		req.URL.Host = target.Host
		var fullPath string
		if strings.HasPrefix(path, "/v1") || strings.HasPrefix(path, "/v2") {
			fullPath = utils.SingleJoiningSlash(dockerContextPath+repoKey, path)
		} else {
			fullPath = path
		}
		req.URL.Path = fullPath
		req.Header.Add("X-Artifactory-Override-Base-Url", "http://"+target.Host+"/artifactory")
		req.Header.Add("X-Forwarded-Proto", "http")
	}
	return &httputil.ReverseProxy{Director: director}
}

// initiallizes a new SingleHostReverseProxy with URL, path, and repo key
func newProxy(target string, path string, repoKey string) *prox {
	url, _ := url.Parse(target)
	return &prox{target: url, proxy: newSingleHostReverseProxy(url, path, repoKey), path: path}
}

// handler for UI requests
func (h *uiHandler) handleFunc(w http.ResponseWriter, r *http.Request) {
	Info.Println(r.UserAgent()+":"+r.Method, r.URL.Path, r.Body)
	w.Header().Set("X-Groxy-Vesrion", "0.1")
	path := r.URL.Path
	p := newProxy(h.ArtifactoryHost, path, "")
	p.proxy.Transport = &myTransport{}
	p.proxy.ServeHTTP(w, r)
}

// handler for v2 requests
func (h *v2Handler) handleFunc(w http.ResponseWriter, r *http.Request) {
	Info.Println(r.UserAgent()+":"+r.Method, r.URL.Path, r.Body)
	w.Header().Set("X-Groxy-Vesrion", "0.1")
	path := r.URL.Path
	p := newProxy(h.ArtifactoryHost, path, h.RepoKey)
	p.proxy.Transport = &myTransport{}
	p.proxy.ServeHTTP(w, r)
}

func (t *myTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	response, err := http.DefaultTransport.RoundTrip(request)
	if err != nil {
		return nil, err
	}
	// get body bytes
	body, err := httputil.DumpResponse(response, true)
	if err != nil {
		return nil, err
	}

	// Treating all Content-Type(s) as text
	Info.Println("\n\nUpstream response: " + string(body))

	return response, err
}

// handler for v1 requests
func (h *v1Handler) handleFunc(w http.ResponseWriter, r *http.Request) {
	// docker will try V2 before it tries V1. We need to make sure /v2 requests will get a 404
	isV1Request := utils.ValidateV1(r.URL.Path)
	if !isV1Request {
		http.NotFound(w, r)
		return
	}
	Info.Println(r.UserAgent()+":"+r.Method, r.URL.Path, r.Body)
	r.Header.Add("X-Groxy-Version", "0.1")
	p := newProxy(h.ArtifactoryHost, r.URL.Path, h.RepoKey)
	p.proxy.Transport = &myTransport{}
	p.proxy.ServeHTTP(w, r)
}

func printer(uiPort string, v1Port string, v2Port string, artifactoryTarget string) {
	fmt.Println("#### Groxy - v0.1 ####")
	fmt.Println("Listening for UI traffic on port", uiPort)
	fmt.Println("Listening for V1 traffic on port", v1Port)
	fmt.Println("Listening for V2 traffic on port", v2Port)
	fmt.Println("Proxying: " + artifactoryTarget)
}

func loadConf(location string) (configuration Conf) {
	// groxy can run in two conf modes:
	// 1.Host mode - the config.json location should be passed with the '-conf' flag for groxy.
	// If the param isn't being passed, groxy assumes the file to be under /
	// 2.Docker mode - the conf file is assumed to be placed under /. No need to pass the -conf param
	// Docker mode is indicated by the DOCKER_MODE env variable

	// We can't use usr.Current() since it's not supported with cross-compiled binaries :(
	// https://github.com/golang/go/issues/6376

	// Docker mode:
	if os.Getenv("DOCKER_MODE") == "true" {
		confFile, _ := os.Open("/config.json")
		decoder := json.NewDecoder(confFile)
		configuration = Conf{}
		err := decoder.Decode(&configuration)
		if err != nil {
			fmt.Println("WARNING: error decoding JSON! verify the validity of your JSON", err)
		} else {
			Info.Println("config.json was loaded")
		}
		return configuration
	} else {
		// Host mode:
		// validate ~/.groxy/config.json exists
		// TODO: add support for windows
		//if _, err = os.Stat(usr.HomeDir + "/.groxy/config.json"); os.IsNotExist(err) {
		//	fmt.Println("WARNING: ~/.groxy/config.json could not be found!")
		//	return
		//}
		if _, err := os.Stat(location); os.IsNotExist(err) {
			Error.Println("ERROR: config.json could not be found!")
			return
		}
		confFile, _ := os.Open(location)
		decoder := json.NewDecoder(confFile)
		configuration = Conf{}
		err := decoder.Decode(&configuration)
		if err != nil {
			Error.Println("error decoding JSON! verify the validity of your JSON", err)
		} else {
			Info.Println("config.json was loaded")
		}
		return configuration
	}
}

func initLoggers(
	traceHandle io.Writer,
	infoHandle io.Writer,
	warningHandle io.Writer,
	errorHandle io.Writer) {

	Trace = log.New(traceHandle,
		"TRACE: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Info = log.New(infoHandle,
		"INFO: ",
		log.Ldate|log.Ltime)

	Warning = log.New(warningHandle,
		"WARNING: ",
		log.Ldate|log.Ltime|log.Lshortfile)

	Error = log.New(errorHandle,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
}

func main() {

	initLoggers(ioutil.Discard, os.Stdout, os.Stdout, os.Stderr)

	// defaults
	const (
		defaultUIPort           = "9010"
		defaultV1Port           = "9011"
		defaultV2Port           = "9012"
		defaultArtifactory      = "http://127.0.0.1:8080"
		defaultUIPortUsage      = "default server port for UI requests"
		defaultV1PortUsage      = "default server port for V1 traffic"
		defaultV2PortUsage      = "default server port for V2 traffic"
		defaultArtifactoryUsage = "default redirect url, 'http://127.0.0.1:8080'"
		defaultConf             = "/"
		defaultConfUsage        = "Location of config.json"
	)

	// flags
	var confLocation, uiPortFlag, v1PortFlag, v2PortFlag, artifactoryFlag string
	flag.StringVar(&uiPortFlag, "uiPort", defaultUIPort, defaultUIPortUsage)
	flag.StringVar(&v1PortFlag, "v1Port", defaultV1Port, defaultV1PortUsage)
	flag.StringVar(&v2PortFlag, "v2Port", defaultV2Port, defaultV2PortUsage)
	flag.StringVar(&artifactoryFlag, "artifactory", defaultArtifactory, defaultArtifactoryUsage)
	flag.StringVar(&confLocation, "conf", defaultConf, defaultConfUsage)
	flag.Parse()

	configuration := loadConf(confLocation)

	// conf provided parameters always override flags
	var artifactoryTarget, uiPort, v1Port, v2Port string

	if configuration.ArtifactoryHost != "" {
		artifactoryTarget = configuration.ArtifactoryHost
	} else {
		artifactoryTarget = artifactoryFlag
	}

	if configuration.DefaultUIPort != "" {
		uiPort = configuration.DefaultUIPort
	} else {
		uiPort = uiPortFlag
	}

	if configuration.DefaultV1Port != "" {
		v1Port = configuration.DefaultV1Port
	} else {
		v1Port = v1PortFlag
	}

	if configuration.DefaultV2Port != "" {
		v2Port = configuration.DefaultV2Port
	} else {
		v2Port = v2PortFlag
	}

	printer(uiPort, v1Port, v2Port, artifactoryTarget)

	v1Hndlr := v1Handler{
		ArtifactoryHost: artifactoryTarget,
		RepoKey:         configuration.V1RepoKey,
	}

	v2Hndlr := v2Handler{
		ArtifactoryHost: artifactoryTarget,
		RepoKey:         configuration.V2RepoKey,
	}

	uiHndlr := uiHandler{
		ArtifactoryHost: artifactoryTarget,
	}

	go func() {
		http.HandleFunc("/v1/", v1Hndlr.handleFunc)
		http.ListenAndServe(":"+v1Port, nil)
	}()

	go func() {
		http.HandleFunc("/v2/", v2Hndlr.handleFunc)
		http.ListenAndServe(":"+v2Port, nil)
	}()

	http.HandleFunc("/", uiHndlr.handleFunc)
	http.ListenAndServe(":"+uiPort, nil)

}
