package mrpc

import (
	"context"
	"github.com/NotFound1911/mrpc/message"
)

type Service interface {
	Name() string
}

type Proxy interface {
	Invoke(ctx context.Context, req *message.Request) (*message.Response, error)
}
