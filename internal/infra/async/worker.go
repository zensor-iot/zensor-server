package async

import "context"

type Worker interface {
	Run(context.Context, func())
	Shutdown()
}
