package main

import (
	"fmt"
	"net/http"
)

type AzureWeb struct {
}

//TODO: implement these using Azure specific API's

func (a AzureWeb) TokenManager(w http.ResponseWriter, r *http.Request) {
	RequestInfo(r)

	w.WriteHeader(http.StatusNotImplemented)
	fmt.Fprintf(w, "Not implmented Yet")

	fmt.Println("Endpoint Hit: tokenManager")
}

func (a AzureWeb) TokenWorker(w http.ResponseWriter, r *http.Request) {
	RequestInfo(r)

	w.WriteHeader(http.StatusNotImplemented)
	fmt.Fprintf(w, "Not implmented Yet")

	fmt.Println("Endpoint Hit: tokenWorker")
}
