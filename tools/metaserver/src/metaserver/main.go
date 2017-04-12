package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
)

// Each IAAS service needs to implement these 8 methods.
// AWS already implemnted in aws.go
// Azure stubbed out in azure.go

// WebInstance instance struct returned by the different cloud providers
type WebInstance struct {
	ID               string
	InstanceName     string
	InstanceID       string
	InstanceType     string
	InstanceNic      string
	PublicIPAddress  string
	PrivateIPAddress string
	InstanceState    string
	InstanceAZ       string
}

// Web interface for all cloud providers
type Web interface {
	TokenManager(w http.ResponseWriter, r *http.Request)
	TokenWorker(w http.ResponseWriter, r *http.Request)
	Managers() []WebInstance
	Workers() []WebInstance
	Instances(w http.ResponseWriter, r *http.Request)
	DtrInfo(w http.ResponseWriter, r *http.Request)
}

func index(w http.ResponseWriter, r *http.Request) {
	// print the home index page.
	RequestInfo(r)
	fmt.Fprintln(w, "token/")
}

func tokenIndex(w http.ResponseWriter, r *http.Request) {
	// print the token index page.
	RequestInfo(r)
	fmt.Fprintln(w, "manager/\nworker/")
}

func errorHandler(w http.ResponseWriter, r *http.Request) {
	// handle the errors
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintln(w, "404: Page not found")
}

func handleRequests(iaasProvider string) {
	r := mux.NewRouter()

	// 404 catcher
	r.NotFoundHandler = http.HandlerFunc(errorHandler)

	// common handlers
	r.HandleFunc("/", index)
	r.HandleFunc("/token/", tokenIndex)

	// specific to the service
	var w Web
	if iaasProvider == "aws" {
		fmt.Println("AWS service")
		w = AWSWeb{}
	} else if iaasProvider == "azure" {
		fmt.Println("Azure service")
		w = AzureWeb{}
	} else {
		fmt.Printf("%v service\n", iaasProvider)
		panic("-iaas_provider was not a valid option.")
	}

	// used to get the swarm tokens
	r.HandleFunc("/token/manager/", w.TokenManager) // Return manager token
	r.HandleFunc("/token/worker/", w.TokenWorker)   // Return worker token
	r.HandleFunc("/instances/all/", w.Instances)    // Return all provider instances
	r.HandleFunc("/info/dtr/", w.DtrInfo)           // Return DTR image version

	// setup the custom server.
	srv := &http.Server{
		Handler: r,
		Addr:    ":8080",

		// it's a good idea to set timeouts.
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	}

	// lets try and listen and serve, if not, log a fatal.
	log.Fatal(srv.ListenAndServe())
}

func validate(iaasProvider string) {
	// validate that we have everything we need in order to start. If anything is missing.
	// Then exit.
	if iaasProvider == "aws" {
		// make sure our required ENV variables are available, if not fail.
		if !checkEnvVars("WORKER_SECURITY_GROUP_ID", "MANAGER_SECURITY_GROUP_ID") {
			os.Exit(1)
		}
	} else if iaasProvider == "azure" {
		// make sure our required ENV variables are available, if not fail.
		if !checkEnvVars("APP_ID", "APP_SECRET", "ACCOUNT_ID", "TENANT_ID", "GROUP_NAME", "VMSS_MGR", "VMSS_WRK") {
			os.Exit(1)
		}
	} else {
		fmt.Printf("ERROR: -iaas_provider %v was not a valid option. please pick either 'aws' or 'azure'", iaasProvider)
		os.Exit(1)
	}
}

func checkEnvVars(envVars ...string) bool {
	for _, envVar := range envVars {
		if os.Getenv(envVar) == "" {
			fmt.Printf("ERROR: Missing environment variable: %s.\n", envVar)
			return false
		}
	}
	return true
}

func main() {
	// pass in the iaas_provider to determine which IAAS provider we are on.
	// currently defaults to AWS, since that is the only one implemnted
	iaasProvider := flag.String("iaas_provider", "aws", "IAAS provider (aws, azure, etc)")
	flag.Parse()

	// make sure we are good to go, before we fully start up.
	validate(*iaasProvider)
	// lets handle those requests
	handleRequests(*iaasProvider)
}
