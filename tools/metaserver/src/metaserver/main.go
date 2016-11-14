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
type Web interface {
	TokenManager(w http.ResponseWriter, r *http.Request)
	TokenWorker(w http.ResponseWriter, r *http.Request)
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

func handleRequests(flavor string) {
	r := mux.NewRouter()

	// 404 catcher
	r.NotFoundHandler = http.HandlerFunc(errorHandler)

	// common handlers
	r.HandleFunc("/", index)
	r.HandleFunc("/token/", tokenIndex)

	// specific to the service
	var w Web
	if flavor == "aws" {
		fmt.Println("AWS service")
		w = AWSWeb{}
	} else if flavor == "azure" {
		fmt.Println("Azure service")
		w = AzureWeb{}
	} else {
		fmt.Printf("%v service\n", flavor)
		panic("-flavor was not a valid option.")
	}

	// used to get the swarm tokens
	r.HandleFunc("/token/manager/", w.TokenManager)
	r.HandleFunc("/token/worker/", w.TokenWorker)

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

func validate(flavor string) {
	// validate that we have everything we need in order to start. If anything is missing.
	// Then exit.
	if flavor == "aws" {
		// make sure our required ENV variables are available, if not fail.
		workerGroupId := os.Getenv("WORKER_SECURITY_GROUP_ID")
		managerGroupId := os.Getenv("MANAGER_SECURITY_GROUP_ID")
		if workerGroupId == "" {
			fmt.Printf("ERROR: Missing environment variable: WORKER_SECURITY_GROUP_ID.")
			os.Exit(1)
		}
		if managerGroupId == "" {
			fmt.Printf("ERROR: Missing environment variable: MANAGER_SECURITY_GROUP_ID")
			os.Exit(1)
		}
	} else if flavor == "azure" {
		// add azure validation code here.

	} else {
		fmt.Printf("ERROR: -flavor %v was not a valid option. please pick either 'aws' or 'azure'", flavor)
		os.Exit(1)
	}

}

func main() {
	// pass in the flavor to determine which IAAS provider we are on.
	// currently defaults to AWS, since that is the only one implemnted
	flavor := flag.String("flavor", "aws", "IAAS Flavor (aws, azure, etc)")
	flag.Parse()

	// make sure we are good to go, before we fully start up.
	validate(*flavor)
	// lets handle those requests
	handleRequests(*flavor)
}
