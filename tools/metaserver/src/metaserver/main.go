package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

func index(w http.ResponseWriter, r *http.Request) {
	GetRequestInfo(r)
	fmt.Fprintf(w, "token/\n")
	fmt.Println("Endpoint Hit: index")
}

func tokenIndex(w http.ResponseWriter, r *http.Request) {
	GetRequestInfo(r)
	fmt.Fprintf(w, "manager/\nworker/\n")
	fmt.Println("Endpoint Hit: tokenIndex")
}

func check(w http.ResponseWriter, r *http.Request) {
	GetRequestInfo(r)
	fmt.Fprintf(w, "manager/\nworker/\n")
	fmt.Println("Endpoint Hit: check")
}

func handleRequests(flavor string) {
	http.HandleFunc("/", index)
	http.HandleFunc("/token/", tokenIndex)

	http.HandleFunc("/check/", check)

	if flavor == "aws" {
		fmt.Println("AWS service")

		http.HandleFunc("/token/manager/", AwsTokenManager)
		http.HandleFunc("/token/worker/", AwsTokenWorker)

		// These are for debugging
		http.HandleFunc("/nodes/", AwsNodes)                //swarm nodes
		http.HandleFunc("/instances/", AwsInstances)        //aws instances
		http.HandleFunc("/check/manager/", AwsCheckManager) //would it pass manager check
		http.HandleFunc("/check/worker/", AwsCheckWorker)   //would it pass worker check
	} else if flavor == "azure" {
		fmt.Println("Azure service")

		http.HandleFunc("/token/manager/", AzureTokenManager)
		http.HandleFunc("/token/worker/", AzureTokenWorker)

		// These are for debugging
		http.HandleFunc("/nodes/", AzureNodes)                //swarm nodes
		http.HandleFunc("/instances/", AzureInstances)        //azure instances
		http.HandleFunc("/check/manager/", AzureCheckManager) //would it pass manager check
		http.HandleFunc("/check/worker/", AzureCheckWorker)   //would it pass worker check
	} else {
		fmt.Printf("%v service\n", flavor)
	}
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func main() {
	flavor := flag.String("flavor", "aws", "IAAS Flavor (aws, azure, etc)")

	handleRequests(*flavor)
}
