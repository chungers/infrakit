package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// Each IAAS service needs to implement these 8 methods.
// AWS already implemnted in aws.go
// Azure stubbed out in azure.go
type Web interface {
	TokenManager(w http.ResponseWriter, r *http.Request)
	TokenWorker(w http.ResponseWriter, r *http.Request)
	CheckManager(w http.ResponseWriter, r *http.Request)
	CheckWorker(w http.ResponseWriter, r *http.Request)
	Instances(w http.ResponseWriter, r *http.Request)
	Nodes(w http.ResponseWriter, r *http.Request)
	ManagerInstances(w http.ResponseWriter, r *http.Request)
	WorkerInstances(w http.ResponseWriter, r *http.Request)
}

func index(w http.ResponseWriter, r *http.Request) {
	// print the home index page.
	GetRequestInfo(r)
	fmt.Fprintf(w, "token/\ncheck/\nnodes/\ninstances/\n")
}

func instanceIndex(w http.ResponseWriter, r *http.Request) {
	// print the instance index page.
	GetRequestInfo(r)
	fmt.Fprintf(w, "all/\nmanagers/\nworkers/\n")
}

func tokenIndex(w http.ResponseWriter, r *http.Request) {
	// print the token index page.
	GetRequestInfo(r)
	fmt.Fprintf(w, "manager/\nworker/\n")
}

func check(w http.ResponseWriter, r *http.Request) {
	// print the check index page
	GetRequestInfo(r)
	fmt.Fprintf(w, "manager/\nworker/\n")
}

func errorHandler(w http.ResponseWriter, r *http.Request) {
	// handle the errors
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprint(w, "404: Page not found")
}

func handleRequests(flavor string) {
	r := mux.NewRouter()

	// 404 catcher
	r.NotFoundHandler = http.HandlerFunc(errorHandler)

	// common handlers
	r.HandleFunc("/", index)
	r.HandleFunc("/instances/", instanceIndex)
	r.HandleFunc("/token/", tokenIndex)
	r.HandleFunc("/check/", check)

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

	// These are for debugging
	r.HandleFunc("/nodes/", w.Nodes)                         //swarm nodes
	r.HandleFunc("/instances/all/", w.Instances)             //provider instances
	r.HandleFunc("/instances/managers/", w.ManagerInstances) //provider instances
	r.HandleFunc("/instances/workers/", w.WorkerInstances)   //provider instances
	r.HandleFunc("/check/manager/", w.CheckManager)          //would it pass manager check
	r.HandleFunc("/check/worker/", w.CheckWorker)            //would it pass worker check

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

func main() {
	// pass in the flavor to determine which IAAS provider we are on.
	// currently defaults to AWS, since that is the only one implemnted
	flavor := flag.String("flavor", "aws", "IAAS Flavor (aws, azure, etc)")
	flag.Parse()

	// lets handle those requests
	handleRequests(*flavor)
}
