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

func main() {
	// pass in the flavor to determine which IAAS provider we are on.
	// currently defaults to AWS, since that is the only one implemnted
	flavor := flag.String("flavor", "aws", "IAAS Flavor (aws, azure, etc)")
	flag.Parse()

	// lets handle those requests
	handleRequests(*flavor)
}
