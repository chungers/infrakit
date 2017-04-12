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
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/sockets"
	"github.com/docker/go-connections/tlsconfig"
)

const (
	clientVersion    = "1.24"    // docker client version
	dtrContainerName = "dtr-api" // dtr-api container name lookup
)

// DockerClient get docker APIclient and context
func DockerClient() (client.APIClient, context.Context) {
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

// RequestIP get IP from request
func RequestIP(r *http.Request) string {
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

// RequestInfo get info about the request
func RequestInfo(r *http.Request) {
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
	fmt.Println("")
}

// SwarmNodes get a list of swarm nodes
func SwarmNodes() []swarm.Node {
	cli, ctx := DockerClient()

	// get the list of swarm nodes
	nodes, err := cli.NodeList(ctx, types.NodeListOptions{})
	if err != nil {
		panic(err)
	}
	return nodes
}

func nodeIP(id string) string {
	cli, ctx := DockerClient()
	// inspect the node ID provided
	node, _, err := cli.NodeInspectWithRaw(ctx, id)
	if err != nil {
		return ""
	}
	return node.Status.Addr
}

func alreadyInSwarm(ip string) bool {
	// Is the node making the request, already in the swarm.
	nodes := SwarmNodes()
	for _, node := range nodes {
		nodeIP := nodeIP(node.ID)
		if ip == nodeIP {
			return true
		}
	}
	return false
}

func isNodeInList(ip string, instances []WebInstance) bool {
	// given an IP, find out if it is in the instance list.
	for _, instance := range instances {
		if ip == instance.PrivateIPAddress {
			return true
		}
	}
	return false
}

func isManagerNode(a Web, ip string) bool {
	// Is the node making the request a manager node
	return isNodeInList(ip, a.Managers())
}

func isWorkerNode(a Web, ip string) bool {
	// Is the node making the request a worker node
	return isNodeInList(ip, a.Workers())
}

func getDTRImage() string {
	cli, ctx := DockerClient()
	// Find DTR image
	filters := filters.NewArgs()
	// filters.Add("reference", dtrReference)
	filters.Add("name", dtrContainerName)
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{Filters: filters})
	if err != nil {
		return ""
	}
	if len(containers) > 0 {
		return strings.Split(containers[0].Image, ":")[1]
	}
	return ""
}
