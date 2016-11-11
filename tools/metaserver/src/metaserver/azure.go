package main

import (
	"fmt"
	"net/http"
)

type AzureWeb struct {
}

//TODO: implement these using Azure specific API's

func (a AzureWeb) TokenManager(w http.ResponseWriter, r *http.Request) {
	GetRequestInfo(r)

	w.WriteHeader(http.StatusNotImplemented)
	fmt.Fprintf(w, "Not implmented Yet\n")

	fmt.Println("Endpoint Hit: tokenManager")
}

func (a AzureWeb) TokenWorker(w http.ResponseWriter, r *http.Request) {
	GetRequestInfo(r)

	w.WriteHeader(http.StatusNotImplemented)
	fmt.Fprintf(w, "Not implmented Yet\n")

	fmt.Println("Endpoint Hit: tokenWorker")
}

func (a AzureWeb) CheckManager(w http.ResponseWriter, r *http.Request) {
	GetRequestInfo(r)

	w.WriteHeader(http.StatusNotImplemented)
	fmt.Fprintf(w, "Not implmented Yet\n")

}

func (a AzureWeb) CheckWorker(w http.ResponseWriter, r *http.Request) {
	GetRequestInfo(r)

	w.WriteHeader(http.StatusNotImplemented)
	fmt.Fprintf(w, "Not implmented Yet\n")

}

func (a AzureWeb) Instances(w http.ResponseWriter, r *http.Request) {
	GetRequestInfo(r)

	w.WriteHeader(http.StatusNotImplemented)
	fmt.Fprintf(w, "Not implmented Yet\n")

}

func (a AzureWeb) ManagerInstances(w http.ResponseWriter, r *http.Request) {
	GetRequestInfo(r)

	w.WriteHeader(http.StatusNotImplemented)
	fmt.Fprintf(w, "Not implmented Yet\n")

}

func (a AzureWeb) WorkerInstances(w http.ResponseWriter, r *http.Request) {
	GetRequestInfo(r)

	w.WriteHeader(http.StatusNotImplemented)
	fmt.Fprintf(w, "Not implmented Yet\n")

}

func (a AzureWeb) Nodes(w http.ResponseWriter, r *http.Request) {
	GetRequestInfo(r)

	w.WriteHeader(http.StatusNotImplemented)
	fmt.Fprintf(w, "Not implmented Yet\n")

}
