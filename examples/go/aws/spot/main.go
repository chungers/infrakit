package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/docker/infrakit/pkg/api"
	"github.com/docker/infrakit/pkg/api/compute"
	"github.com/docker/infrakit/pkg/api/scope"

	_ "github.com/docker/infrakit/pkg/api/driver/aws"
)

func InstallDocker(os string) compute.ShellScript {
	return compute.ShellScript("wget -qO https://get.docker.com | sh")
}

func InstallGit(os string) compute.ShellScript {
	return compute.ShellScript("apt-get install -y git-core")
}

func InstallGo(os string) compute.ShellScript {
	return compute.ShellScript("apt-get install -y golang")
}

func main() {

	scope, err := scope.Connect("local://",
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

	machine, err := compute.Customize(profile)
	if err != nil {
		// Err if profile is not a machine profile; can't down-cast.
		panic(err)
	}

	machine.AddInit(InstallDocker("ubuntu")).
		AddInit(InstallGit("ubuntu")).
		AddInit(InstallGo("ubuntu")).
		AddTag("foo", "bar")

	for _, id := range []string{"hello", "world"} {

		machine.SetLogicalID(id).Set("subnetID", "1234")

		meta, err := scope.Enforce(machine)
		if err != nil {
			panic(err)
		}

		fmt.Println("created instance %v", meta)
	}

	fmt.Println("bye")
}
