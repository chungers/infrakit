package v0

import (
	"github.com/docker/infrakit/pkg/cli"
	"github.com/docker/infrakit/pkg/controller"
	"github.com/docker/infrakit/pkg/spi/flavor"
	"github.com/docker/infrakit/pkg/spi/group"
	"github.com/docker/infrakit/pkg/spi/instance"
	"github.com/docker/infrakit/pkg/spi/metadata"
	"github.com/docker/infrakit/pkg/spi/resource"

	// v0 loads these packages
	_ "github.com/docker/infrakit/pkg/cli/v0/controller"
	_ "github.com/docker/infrakit/pkg/cli/v0/event"
	_ "github.com/docker/infrakit/pkg/cli/v0/flavor"
	_ "github.com/docker/infrakit/pkg/cli/v0/group"
	_ "github.com/docker/infrakit/pkg/cli/v0/instance"
	_ "github.com/docker/infrakit/pkg/cli/v0/loadbalancer"
	_ "github.com/docker/infrakit/pkg/cli/v0/manager"
	_ "github.com/docker/infrakit/pkg/cli/v0/metadata"
	_ "github.com/docker/infrakit/pkg/cli/v0/resource"
)

func init() {
	cli.Register(controller.InterfaceSpec,
		[]cli.CmdBuilder{
			Info,
		})

	cli.Register(instance.InterfaceSpec,
		[]cli.CmdBuilder{
			Info,
		})
	cli.Register(flavor.InterfaceSpec,
		[]cli.CmdBuilder{
			Info,
		})
	cli.Register(group.InterfaceSpec,
		[]cli.CmdBuilder{
			Info,
		})
	cli.Register(resource.InterfaceSpec,
		[]cli.CmdBuilder{
			Info,
		})
	cli.Register(metadata.InterfaceSpec,
		[]cli.CmdBuilder{
			Info,
		})
}
