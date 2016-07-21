package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"flag"
	"log"
	"os"
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
	flavor := flag.String("flavor", "", "IAAS Flavor (aws, azure, etc)")
	swarmID := flag.String("swarm_id", "n/a", "Swarm ID")
	nodeID := flag.String("node_id", "", "Node ID")
	event := flag.String("event", "n/a", "Event") // identify, init, ping, scale
	dockerForIAASVersion := os.Getenv("DOCKER_FOR_IAAS_VERSION")
	accountID := os.Getenv("ACCOUNT_ID")
	region := os.Getenv("REGION")

	flag.Parse()

	// Hash the accountId so we don't know what it is.
	hashedAccountID := computeHmac256(accountID, "ZQM7q96ar8g1y7Id")
	var clientCode string

	if *flavor == "aws" {
		clientCode = "0Euz80odMWb07uI6cnhFENW3ohikKpb8"
	} else if *flavor == "azure" {
		clientCode = "1og3BGfY1Dt2aRBSf2SQrUL2VlGBoW5v"
	} else {
		err := errors.New("unknown flavor")
		log.Fatal(err)
		os.Exit(1) // error
	}

	client := analytics.New(clientCode)
	client.Size = 1 // We only send one message at a time, no need to cache

	if *event == "identify" {
		client.Identify(&analytics.Identify{
			UserId: hashedAccountID,
			Traits: map[string]interface{}{
				"region": region,
			},
		})
		client.Track(&analytics.Track{
			Event:  "swarm:init",
			UserId: hashedAccountID,
			Properties: map[string]interface{}{
				"swarm_id": *swarmID,
				"region":   region,
			},
		})
	} else if *event == "swarm:init" {
		client.Track(&analytics.Track{
			Event:  "swarm:init",
			UserId: hashedAccountID,
			Properties: map[string]interface{}{
				"swarm_id": *swarmID,
				"region":   region,
			},
		})
	} else if strings.HasPrefix(*event, "node:") {
		client.Track(&analytics.Track{
			Event:  *event,
			UserId: hashedAccountID,
			Properties: map[string]interface{}{
				"swarm_id": *swarmID,
				"node_id":  *nodeID,
				"region":   region,
			},
		})
	} else if strings.HasPrefix(*event, "swarm:") {
		client.Track(&analytics.Track{
			Event:  *event,
			UserId: hashedAccountID,
			Properties: map[string]interface{}{
				"swarm_id":                *swarmID,
				"region":                  region,
				"service_count":           *numServices,
				"manager_count":           *numManagers,
				"worker_count":            *numWorkers,
				"docker_version":          *dockerVersion,
				"docker_for_iaas_version": dockerForIAASVersion,
			},
		})
	}
	client.Close()
}
