package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/tlsconfig"
	"github.com/docker/go-connections/sockets"
)

const (
	clientVersion = "1.24"
)

func GetDockerClient() (client.APIClient, context.Context) {
	tlsOptions := tlsconfig.Options{}
	host := "unix:///var/run/docker.sock"
	ctx := context.Background()
	dockerClient, err := NewDockerClient(host, &tlsOptions)
	if err != nil {
		panic(err)
	}
	return dockerClient, ctx
}

func GetRequestIP(r *http.Request) string {
	ip, _, _ := net.SplitHostPort(r.RemoteAddr)
	return ip
}

// NewDockerClient creates a new API client.
func NewDockerClient(host string, tls *tlsconfig.Options) (client.APIClient, error) {
	tlsOptions := tls
	if tls.KeyFile == "" || tls.CAFile == "" || tls.CertFile == "" {
		// The api doesn't like it when you pass in not nil but with zero field values...
		tlsOptions = nil
	}
	customHeaders := map[string]string{
		"User-Agent": clientUserAgent(),
	}
	verStr := clientVersion
	if tmpStr := os.Getenv("DOCKER_API_VERSION"); tmpStr != "" {
		verStr = tmpStr
	}
	httpClient, err := newHTTPClient(host, tlsOptions)
	if err != nil {
		return &client.Client{}, err
	}
	return client.NewClient(host, verStr, httpClient, customHeaders)
}

func newHTTPClient(host string, tlsOptions *tlsconfig.Options) (*http.Client, error) {

	var config *tls.Config
	var err error

	if tlsOptions != nil {
		config, err = tlsconfig.Client(*tlsOptions)
		if err != nil {
			return nil, err
		}
		//log.Infoln("TLS config=", config)
	}
	tr := &http.Transport{
		TLSClientConfig: config,
	}
	proto, addr, _, err := client.ParseHost(host)
	if err != nil {
		return nil, err
	}
	sockets.ConfigureTransport(tr, proto, addr)
	return &http.Client{
		Transport: tr,
	}, nil
}

func clientUserAgent() string {
	return fmt.Sprintf("Docker-Client/%s (%s)", clientVersion, runtime.GOOS)
}

func GetRequestInfo(r *http.Request) {
	fmt.Printf("Referer: %s\n", r.Referer())
	fmt.Printf("UserAgent: %s\n", r.UserAgent())
	fmt.Printf("RemoteAddr: %s\n", r.RemoteAddr)
	fmt.Printf("Method: %s\n", r.Method)
	fmt.Printf("Host: %s\n", r.Host)
	ip, port, _ := net.SplitHostPort(r.RemoteAddr)
	fmt.Printf("IP: %s\n", ip)
	fmt.Printf("port: %s\n", port)
	userIP := net.ParseIP(ip)
	fmt.Printf("userIP: %s\n", userIP)
	forward := r.Header.Get("X-Forwarded-For")
	fmt.Printf("forward: %s\n", forward)
	ips := strings.Split(forward, ", ")
	for _, ip := range ips {
		fmt.Println(ip)
	}
	for k, v := range r.Header {
		fmt.Println("key:", k, "value:", v)
	}

	dump, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Printf("dump error")
	}

	fmt.Printf("%q", dump)

}

func GetSwarmNodes() []swarm.Node {
	cli, ctx := GetDockerClient()

	// this is just for testing, remove.
	nodes, err := cli.NodeList(ctx, types.NodeListOptions{})
	if err != nil {
		panic(err)
	}
	return nodes
}
