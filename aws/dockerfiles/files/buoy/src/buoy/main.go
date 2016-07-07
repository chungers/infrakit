package main

import (
	"flag"
	"os"
)
import "github.com/segmentio/analytics-go"

func main() {

	numManagers := flag.Int("managers", 0, "Number of Managers")
	numWorkers := flag.Int("workers", 0, "Number of Workers")
	numServices := flag.Int("services", 0, "Number of Services")
	isIdentify := flag.Bool("identify", false, "Do we need to Identify?")
	dockerVersion := flag.String("docker_version", "n/a", "Docker Version")
	dockerForAWSVersion := os.Getenv("DOCKER_FOR_AWS_VERSION")
	stackID := os.Getenv("STACK_ID")
	accountID := os.Getenv("ACCOUNT_ID")
	region := os.Getenv("AWS_DEFAULT_REGION")

	flag.Parse()
	// fmt.Println("Services:", *numServices)
	// fmt.Println("Managers:", *numManagers)
	// fmt.Println("Workers:", *numWorkers)
	// fmt.Println("isIdentify:", *isIdentify)
	// fmt.Println("dockerVersion:", *dockerVersion)
	// fmt.Println("stackID:", stackID)
	// fmt.Println("accountID:", accountID)
	// fmt.Println("dockerForAWSVersion:", dockerForAWSVersion)

	client := analytics.New("0Euz80odMWb07uI6cnhFENW3ohikKpb8")
	client.Size = 1 // We only send one message at a time, no need to cache

	if *isIdentify {
		// fmt.Println("Sending Identify")
		client.Identify(&analytics.Identify{
			UserId: accountID,
			Traits: map[string]interface{}{
				"aws_region": region,
			},
		})
	} else {
		// fmt.Println("Sending Ping")
		client.Track(&analytics.Track{
			Event:  "ping",
			UserId: accountID,
			Properties: map[string]interface{}{
				"cluster_id":             stackID,
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
