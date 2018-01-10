package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/docker/infrakit/pkg/api"
	"github.com/docker/infrakit/pkg/api/types"

	_ "github.com/docker/infrakit/pkg/api/driver/aws"
)

func InstallDocker(os string) types.ShellScript {
	return types.ShellScript("wget -qO https://get.docker.com | sh")
}

func InstallGit(os string) types.ShellScript {
	return types.ShellScript("apt-get install -y git-core")
}

func InstallGo(os string) types.ShellScript {
	return types.ShellScript("apt-get install -y golang")
}

func main() {

	scope, err := api.Connect("local://",
		api.Options{
			ProfilePaths: strings.Split(os.Getenv("PROFILES_PATH"), ","),
		})
	if err != nil {
		panic(err)
	}

	api.Logger.Info("Started")

	// list all the known instance profiles
	profiles, err := scope.Profiles()
	if err != nil {
		panic(err)
	}

	profileName := "spot-small"

	profile, has := profiles[profileName]
	if !has {
		panic("cannot find profile")
	}

	profile.AddTag("foo", "bar").
		AddInit(InstallDocker("ubuntu")).
		AddInit(InstallGit("ubuntu")).
		AddInit(InstallGo("ubuntu"))

	for _, id := range []string{"hello", "world"} {

		profile.SetLogicalID(id).Set("subnetID", "1234")

		meta, err := scope.Provision(profile)
		if err != nil {
			panic(err)
		}

		fmt.Println("created instance %v", meta)
	}

	fmt.Println("bye")
}
