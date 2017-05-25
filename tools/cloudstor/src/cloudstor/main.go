package main

import (
	"os"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/go-plugins-helpers/volume"
)

const (
	volumeDriverName = "cloudstor"
	metadataRoot     = "/etc/docker/plugins/cloudstor/volumes"
	socketAddress    = "/run/docker/plugins/cloudstor.sock"
	mountPoint       = "/mnt/cloudstor"
)

func main() {
	debug := os.Getenv("DEBUG")
	if ok, _ := strconv.ParseBool(debug); ok {
		log.SetLevel(log.DebugLevel)
	}
	cloudEnv := os.Getenv("CLOUD_PLATFORM")
	if cloudEnv == "AZURE" {
		accountName := os.Getenv("AZURE_STORAGE_ACCOUNT")
		accountKey := os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")
		storEndPt := os.Getenv("AZURE_STORAGE_ENDPOINT")
		if accountName == "" || accountKey == "" {
			log.Fatal("azure storage account name and key must be provided.")
		}

		log.WithFields(log.Fields{
			"accountName": accountName,
		}).Debug("Starting Azure server.")

		driver, err := newAZFSDriver(accountName, accountKey, metadataRoot, storEndPt)
		if err != nil {
			log.Fatal(err)
		}
		h := volume.NewHandler(driver)
		log.Fatal(h.ServeUnix(socketAddress, 0))

	} else if cloudEnv == "AWS" {
		efsIDRegular := os.Getenv("EFS_ID_REGULAR")
		efsIDMaxIO := os.Getenv("EFS_ID_MAXIO")
		stackID := os.Getenv("AWS_STACK_ID")
		efsSupported, err := strconv.ParseBool(os.Getenv("EFS_SUPPORTED"))
		if err != nil {
			log.Fatal(err)
		}

		log.WithFields(log.Fields{
			"Regular EFS ID": efsIDRegular,
			"MaxIO EFS ID":   efsIDMaxIO,
			"AWS Stack ID":   stackID,
		}).Debug("Starting server.")

		driver, err := newAWSDriver(efsIDRegular, efsIDMaxIO, metadataRoot, stackID, efsSupported)
		if err != nil {
			log.Fatal(err)
		}
		h := volume.NewHandler(driver)
		log.Fatal(h.ServeUnix(socketAddress, 0))

	} else if cloudEnv == "GCP" {
		log.Debug("Starting GCP server.")
		driver, err := newGCPDriver(metadataRoot)
		if err != nil {
			log.Fatal(err)
		}
		h := volume.NewHandler(driver)
		log.Fatal(h.ServeUnix(socketAddress, 0))
	} else {
		log.Fatal("Failed to initialize Cloudstor: unsupported cloud platform")
	}
}
