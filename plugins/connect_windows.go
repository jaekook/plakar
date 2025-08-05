package plugins

import (
	"errors"

	"google.golang.org/grpc"
)

func connectPlugin(pluginPath string, args []string) (grpc.ClientConnInterface, error) {
	return nil, errors.ErrUnsupported
}
