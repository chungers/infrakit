package types

import (
	logutil "github.com/docker/infrakit/pkg/log"
	"github.com/docker/infrakit/pkg/run/depends"
	"github.com/docker/infrakit/pkg/spi/controller"
	"github.com/docker/infrakit/pkg/types"
)

var (
	log     = logutil.New("module", "controller/script/types")
	debugV  = logutil.V(500)
	debugV2 = logutil.V(1000)
)

func init() {
	depends.Register("script", types.InterfaceSpec(controller.InterfaceSpec), ResolveDependencies)
}
