package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"flag"
	"os"
	"regexp"
	"strings"

	"github.com/segmentio/analytics-go"
)

func computeHmac256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func main() {

	numManagers := flag.Int("managers", 0, "Number of Managers")
	numWorkers := flag.Int("workers", 0, "Number of Workers")
	numServices := flag.Int("services", 0, "Number of Services")
	dockerVersion := flag.String("docker_version", "n/a", "Docker Version")
	iaas_provider := flag.String("iaas_provider", "n/a", "IAAS provider (aws, azure, gcp, etc)")
	channel := flag.String("channel", "n/a", "IAAS Channel (test, beta, stable, etc)")
	swarmID := flag.String("swarm_id", "n/a", "Swarm ID")
	nodeID := flag.String("node_id", "n/a", "Node ID")
	edition := flag.String("edition", "n/a", "Edition (ce, ee)")
	editionOS := flag.String("editionOS", "moby", "Edition OS (centos, oel, ubuntu, ws2016, rhel, sles)")
	editionVersion := flag.String("editionVersion", "n/a", "Edition Version")
	event := flag.String("event", "n/a", "Event") // identify, init, ping, scale
	dockerForIAASVersion := os.Getenv("DOCKER_FOR_IAAS_VERSION")
	accountID := os.Getenv("ACCOUNT_ID")
	region := os.Getenv("REGION")

	flag.Parse()

	if *channel == "n/a" {
		channel = getChannel(dockerForIAASVersion)
	}

	if *editionVersion == "n/a" {
		editionVersion = dockerForIAASVersion
	}

	if *edition == "n/a" {
		edition = os.Getenv("EDITION")
	}

	// Hash the accountId so we don't know what it is.
	hashedAccountID := computeHmac256(accountID, "ZQM7q96ar8g1y7Id")
	clientCode := "jLwurYoMosZliChljnSNq7mCAOOd8Vnn"

	client := analytics.New(clientCode)
	client.Size = 1 // We only send one message at a time, no need to cache

	if *event == "identify" {
		client.Identify(&analytics.Identify{
			UserId: hashedAccountID,
			Traits: map[string]interface{}{
				"region":           region,
				"edition":          *edition,
				"edition_os":       *editionOS,
				"iaas_provider":    *iaas_provider,
			},
		})
	} else if *event == "swarm:init" {
		client.Track(&analytics.Track{
			Event:  "swarm:init",
			UserId: hashedAccountID,
			Properties: map[string]interface{}{
				"swarm_id":         *swarmID,
				"node_id":          *nodeID,
				"region":           region,
				"edition_version":  *editionVersion,
				"channel":          *channel,
			},
		})
	} else if strings.HasPrefix(*event, "node:") {
		client.Track(&analytics.Track{
			Event:  *event,
			UserId: hashedAccountID,
			Properties: map[string]interface{}{
				"swarm_id":         *swarmID,
				"node_id":          *nodeID,
				"region":           region,
				"edition_version":  *editionVersion,
				"channel":          *channel,
			},
		})
	} else if strings.HasPrefix(*event, "swarm:") {
		client.Track(&analytics.Track{
			Event:  *event,
			UserId: hashedAccountID,
			Properties: map[string]interface{}{
				"swarm_id":                *swarmID,
				"node_id":                 *nodeID,
				"region":                  region,
				"service_count":           *numServices,
				"manager_count":           *numManagers,
				"worker_count":            *numWorkers,
				"docker_version":          *dockerVersion,
				"edition_version":         *editionVersion,
				"channel":                 *channel,
			},
		})
	}
	client.Close()
}

func getChannel(version string) *string {
	r := regexp.MustCompile("^(?:azure|aws)-v\\d+.\\d+.\\d+-([0-9a-z-]+)")
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
