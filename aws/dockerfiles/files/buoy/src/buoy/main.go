package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"flag"
	"os"
)
import "github.com/segmentio/analytics-go"

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
	isIdentify := flag.Bool("identify", false, "Do we need to Identify?")
	dockerVersion := flag.String("docker_version", "n/a", "Docker Version")
	swarmID := flag.String("swarm_id", "n/a", "Swarm ID")
	dockerForAWSVersion := os.Getenv("DOCKER_FOR_AWS_VERSION")
	accountID := os.Getenv("ACCOUNT_ID")
	region := os.Getenv("AWS_DEFAULT_REGION")

	flag.Parse()

	// Hash the accountId so we don't know what it is.
	hashedAccountID := computeHmac256(accountID, "ZQM7q96ar8g1y7Id")

	client := analytics.New("0Euz80odMWb07uI6cnhFENW3ohikKpb8")
	client.Size = 1 // We only send one message at a time, no need to cache

	if *isIdentify {
		client.Identify(&analytics.Identify{
			UserId: hashedAccountID,
			Traits: map[string]interface{}{
				"aws_region": region,
			},
		})
	} else {
		client.Track(&analytics.Track{
			Event:  "ping",
			UserId: hashedAccountID,
			Properties: map[string]interface{}{
				"cluster_id":             *swarmID,
				"aws_region":             region,
				"service_count":          *numServices,
				"manager_count":          *numManagers,
				"worker_count":           *numWorkers,
				"docker_version":         *dockerVersion,
				"docker_for_aws_version": dockerForAWSVersion,
			},
		})
	}
	client.Close()
}
