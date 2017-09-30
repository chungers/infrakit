package oracle

import (
	"fmt"
	"time"

	"github.com/FrenchBen/oracle-sdk-go/core"
	olb "github.com/FrenchBen/oracle-sdk-go/loadbalancer"
	"github.com/Sirupsen/logrus"
	"github.com/docker/editions/pkg/loadbalancer"
	"github.com/prometheus/common/log"
	"github.com/ryanuber/go-glob"
)

type lbDriver struct {
	client      *olb.Client
	name        string
	componentID string
	stack       string
}

// Options - options for Oracle
type Options struct {
	PollingDelaySeconds    int // The number of seconds to delay between polls
	PollingDurationSeconds int // The number of seconds to poll for async status before cancel
	UserID                 string
	ComponentID            string
	TenancyID              string
	Fingerprint            string
	KeyFile                string
	Region                 string
	StackName              string
}

// NewLoadBalancerDriver creates a load balancer driver
func NewLoadBalancerDriver(client *olb.Client, name string, options *Options) (loadbalancer.Driver, error) {
	return &lbDriver{
		client:      client,
		name:        name,
		componentID: options.ComponentID,
		stack:       options.StackName,
	}, nil
}

// Name is the name of the load balancer
func (d *lbDriver) Name() string {
	return d.name
}

// Routes lists all known routes.
func (d *lbDriver) Routes() ([]loadbalancer.Route, error) {
	// List all loadbalancers to filter out the main one
	lbOC, bmcErr := d.client.GetLoadBalancer(d.name)
	if bmcErr != nil {
		log.Fatal("Fatal: ", bmcErr.Code, " Msg: ", bmcErr.Message)
	}
	routes := []loadbalancer.Route{}
	for _, backendSet := range lbOC.BackendSets {
		logrus.Debugln("backendSet: ", backendSet)
		routes = append(routes, loadbalancer.Route{
			Port:             uint32(backendSet.HealthChecker.Port),
			Protocol:         loadbalancer.ProtocolFromString(backendSet.HealthChecker.Protocol),
			LoadBalancerPort: uint32(backendSet.HealthChecker.Port),
		})
	}

	return []loadbalancer.Route{}, nil
}

// Publish publishes a route in the LB by adding a load balancing rule
func (d *lbDriver) Publish(route loadbalancer.Route) (loadbalancer.Result, error) {
	log.Debugln("Publish", route.LoadBalancerPort, "on", d.name)
	name := fmt.Sprintf("%s-%d-%d", route.Protocol, route.LoadBalancerPort, route.Port)
	frontendPort := int(route.LoadBalancerPort)
	backendPort := int(route.Port)
	timeoutMilliseconds := int(5 * 60 * 100)
	backendSet := &olb.BackendSet{
		Name:   name,
		Policy: "ROUND_ROBIN",
		HealthChecker: &olb.HealthChecker{
			Protocol: string(route.Protocol),
			Port:     backendPort,
			URLPath:  "/",
			Timeout:  timeoutMilliseconds,
		},
		Backends: d.backendServers(backendPort),
	}
	// Create backendSet
	if ok, err := d.client.CreateBackendSet(d.name, backendSet); !ok {
		logrus.Warning("Failed to create BackendSet config: ", err.Message)
	}
	listener := &olb.Listener{
		BackendSetName: name,
		Name:           fmt.Sprintf("%s-listener", name),
		Port:           frontendPort,
		Protocol:       string(route.Protocol),
	}
	if ok, err := d.client.CreateListener(d.name, listener); !ok {
		logrus.Warning("Failed to create Listener config: ", err.Message)
	}

	return Response(fmt.Sprintf("Frontend Port %d mapped to Backend Port %d published", frontendPort, backendPort)), nil
}

// Unpublish dissociates the load balancer from the backend service at the given port.
func (d *lbDriver) Unpublish(extPort uint32) (loadbalancer.Result, error) {
	// remove listener & remove backendSet
	port := int(extPort)
	name := fmt.Sprintf("%s-%d-%d", string(loadbalancer.HTTP), port, port)
	if ok, err := d.client.DeleteListener(d.name, fmt.Sprintf("%s-listener", name)); !ok {
		logrus.Warning("Failed to delete Listener: ", err.Message)
	} else {
		logrus.Info("Deleted listener for instance")
	}
	if ok, err := d.client.DeleteBackendSet(d.name, name); !ok {
		logrus.Warning("Failed to delete BackendSet: ", err.Message)
	}
	return Response(fmt.Sprintf("Frontend and Backend Port %d have been unpublished", port)), nil
}

// ConfigureHealthCheck configures the health checks for instance removal and reconfiguration
// The parameters healthy and unhealthy indicate the number of consecutive success or fail pings required to
// mark a backend instance as healthy or unhealthy.   The ping occurs on the backendPort parameter and
// at the interval specified.
func (d *lbDriver) ConfigureHealthCheck(backendPort uint32, healthy, unhealthy int, interval, timeout time.Duration) (loadbalancer.Result, error) {
	return nil, fmt.Errorf("not-implemented")
}

// RegisterBackend registers instances identified by the IDs to the LB's backend pool
func (d *lbDriver) RegisterBackend(id string, more ...string) (loadbalancer.Result, error) {
	return nil, fmt.Errorf("not-implemented")
}

// DeregisterBackend removes the specified instances from the backend pool
func (d *lbDriver) DeregisterBackend(id string, more ...string) (loadbalancer.Result, error) {
	return nil, fmt.Errorf("not-implemented")
}

// filterLoadBalancers returns a filtered list of instances based on the filter provided
func filterLoadBalancers(loadBalancers []olb.LoadBalancer, filter string) []olb.LoadBalancer {
	finalLoadBalancers := loadBalancers[:0]
	for _, loadBalancer := range loadBalancers {
		conditional := true
		if filter != "" {
			conditional = conditional && glob.Glob(filter, loadBalancer.DisplayName)
		}
		if conditional {
			finalLoadBalancers = append(finalLoadBalancers, loadBalancer)
		}
	}
	return finalLoadBalancers
}

func (d *lbDriver) backendServers(port int) []olb.Backend {
	// Create instances client
	coreClient := core.NewClient(d.client.Client, d.componentID)

	options := &core.InstancesParameters{
		Filter: &core.InstanceFilter{
			DisplayName:    fmt.Sprintf("%s-*", d.stack),
			LifeCycleState: "running",
		},
	}
	instances, bmcErr := coreClient.ListInstances(options)
	if bmcErr != nil {
		log.Fatal("Fatal: ", bmcErr.Code, " Msg: ", bmcErr.Message)
	}
	backends := []olb.Backend{}
	for _, instance := range instances {
		vNics := coreClient.ListVNic(instance.ID)
		backends = append(backends, olb.Backend{
			IPAddress: vNics[0].PrivateIP,
			Port:      port,
		})
	}
	return backends
}
