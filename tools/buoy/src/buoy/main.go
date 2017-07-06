package main

import (
	"flag"
	"github.com/docker/editions/tools/buoy/pkg/buoy"
	"os"
	"regexp"
	"strings"
)

// NA represents the default for all non-passed flags
const NA = "n/a"

// given the docker for iaas version get the channel
func getChannel(version string) *string {
	r := regexp.MustCompile("^(?:azure|aws|gcp)-v\\d+.\\d+.\\d+-([0-9a-z-]+)")
	matched := r.FindStringSubmatch(version)
	channel := "test"
	if len(matched) > 1 {
		if strings.Contains(matched[1], "nightly") {
			channel = "nightly"
		} else if strings.Contains(matched[1], "beta") {
			channel = "beta"
		} else {
			channel = "stable"
		}
	}
	return &channel
}

func main() {

	numManagers := flag.Int("managers", 0, "Number of Managers")
	numWorkers := flag.Int("workers", 0, "Number of Workers")
	numServices := flag.Int("services", 0, "Number of Services")
	dockerVersion := flag.String("docker_version", NA, "Docker Version")
	iaasProvider := flag.String("iaas_provider", NA, "IAAS provider (aws, azure, gcp, etc)")
	channel := flag.String("channel", NA, "IAAS Channel (test, beta, stable, etc)")
	swarmID := flag.String("swarm_id", NA, "Swarm ID")
	nodeID := flag.String("node_id", NA, "Node ID")
	edition := flag.String("edition", NA, "Edition (ce, ee)")
	editionOS := flag.String("editionOS", "moby", "Edition OS (centos, oel, ubuntu, ws2016, rhel, sles)")
	editionVersion := flag.String("editionVersion", NA, "Edition Version")
	editionAddOn := flag.String("addon", "base", "Edition Add-On (base, ddc, cloud, etc.)")
	event := flag.String("event", NA, "Event") // identify, init, ping, scale
	// Get Common env vars
	dockerForIAASVersion := os.Getenv("DOCKER_FOR_IAAS_VERSION")
	accountID := os.Getenv("ACCOUNT_ID")
	region := os.Getenv("REGION")
	editionEnv := os.Getenv("EDITION")
	flag.Parse()

	if *channel == NA {
		channel = getChannel(dockerForIAASVersion)
	}

	if *editionVersion == NA {
		editionVersion = &dockerForIAASVersion
	}

	if *edition == NA {
		edition = &editionEnv
	}

	buoyEvent := buoy.BuoyEvent{
		AccountID:      accountID,
		Region:         region,
		Edition:        *edition,
		EditionOS:      *editionOS,
		EditionAddon:   *editionAddOn,
		IaasProvider:   *iaasProvider,
		SwarmID:        *swarmID,
		NodeID:         *nodeID,
		EditionVersion: *editionVersion,
		Channel:        *channel}

	if *event == "identify" {
		// identify the user
		buoyEvent.Identify()
	} else if *event == "swarm:init" {
		// swarm init event, do it on it's own so it doesn't get caught by the
		// swarm: events below.
		buoyEvent.Track("swarm:init")
	} else if strings.HasPrefix(*event, "node:") {
		// node events
		buoyEvent.Track(*event)
	} else if strings.HasPrefix(*event, "swarm:") {
		// swarm events
		buoyEvent.ServiceCount = *numServices
		buoyEvent.ManagerCount = *numManagers
		buoyEvent.WorkerCount = *numWorkers
		buoyEvent.DockerVersion = *dockerVersion

		buoyEvent.Track(*event)
	}
}
