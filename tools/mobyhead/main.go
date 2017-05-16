package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/github"
)

var (
	// Username Github User
	Username string
	// Token Github Personal Access Token
	Token string
)

func main() {
	tp := github.BasicAuthTransport{
		Username: strings.TrimSpace(Username),
		Password: strings.TrimSpace(Token),
	}

	client := github.NewClient(tp.Client())
	ctx := context.Background()
	ref, _, err := client.Git.GetRef(ctx, "docker", "editions.moby", "heads/master")
	if err != nil {
		fmt.Printf("\nerror: %v\n", err)
		return
	}
	fmt.Println(*ref.Object.SHA)
}
