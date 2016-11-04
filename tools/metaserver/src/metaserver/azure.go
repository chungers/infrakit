package main

import (
	"fmt"
	"net/http"
)

func AzureTokenManager(w http.ResponseWriter, r *http.Request) {
	GetRequestInfo(r)

	w.WriteHeader(http.StatusNotImplemented)
	fmt.Fprintf(w, "Not implmented Yet\n")

	fmt.Println("Endpoint Hit: tokenManager")
}

func AzureTokenWorker(w http.ResponseWriter, r *http.Request) {
	GetRequestInfo(r)

	w.WriteHeader(http.StatusNotImplemented)
	fmt.Fprintf(w, "Not implmented Yet\n")

	fmt.Println("Endpoint Hit: tokenWorker")
}

func AzureCheckManager(w http.ResponseWriter, r *http.Request) {
	GetRequestInfo(r)

	w.WriteHeader(http.StatusNotImplemented)
	fmt.Fprintf(w, "Not implmented Yet\n")

}

func AzureCheckWorker(w http.ResponseWriter, r *http.Request) {
	GetRequestInfo(r)

	w.WriteHeader(http.StatusNotImplemented)
	fmt.Fprintf(w, "Not implmented Yet\n")

}

func AzureInstances(w http.ResponseWriter, r *http.Request) {
	GetRequestInfo(r)

	w.WriteHeader(http.StatusNotImplemented)
	fmt.Fprintf(w, "Not implmented Yet\n")

}

func AzureNodes(w http.ResponseWriter, r *http.Request) {
	GetRequestInfo(r)

	w.WriteHeader(http.StatusNotImplemented)
	fmt.Fprintf(w, "Not implmented Yet\n")

}
