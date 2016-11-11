package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"runtime"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/sockets"
	"github.com/docker/go-connections/tlsconfig"
)

const (
	clientVersion = "1.24" // docker client version
)

func GetDockerClient() (client.APIClient, context.Context) {
	// get the docker client
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
	//given the request return the IP address of the requestor
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
	// This is the client UserAgent, we use when connecting to docker.
	return fmt.Sprintf("Docker-Client/%s (%s)", clientVersion, runtime.GOOS)
}

func GetRequestInfo(r *http.Request) {
	// Mostly used for debugging, we can cut this down,
	// since it isn't really used much anymore.
	fmt.Printf("Path:[%s] %s \n", r.Method, r.URL.Path)
	fmt.Printf("User: %s [%s]\n", r.UserAgent(), r.RemoteAddr)
	ip, port, _ := net.SplitHostPort(r.RemoteAddr)
	userIP := net.ParseIP(ip)
	fmt.Printf("userIP: %s on port %s\n", userIP, port)
	forward := r.Header.Get("X-Forwarded-For")
	if len(forward) > 0 {
		fmt.Printf("forward: %s\n", forward)
		ips := strings.Split(forward, ", ")
		for _, ip := range ips {
			fmt.Println(ip)
		}
	}
	fmt.Printf("\n")
}

func GetSwarmNodes() []swarm.Node {
	cli, ctx := GetDockerClient()

	// get the list of swarm nodes
	nodes, err := cli.NodeList(ctx, types.NodeListOptions{})
	if err != nil {
		panic(err)
	}
	return nodes
}
