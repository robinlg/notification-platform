package grpc

import (
	"fmt"

	"github.com/ecodeclub/ekit/syncx"
	"github.com/gotomicro/ego/client/egrpc"
)

type Clients[T any] struct {
	clientMap syncx.Map[string, T]
	creator   func(conn *egrpc.Component) T
}

func NewClients[T any](creator func(conn *egrpc.Component) T) *Clients[T] {
	return &Clients[T]{creator: creator}
}

func (c *Clients[T]) Get(serviceName string) T {
	client, ok := c.clientMap.Load(serviceName)
	if !ok {
		// 初始化client
		// ego如果发现服务失败，会panic
		grpcConn := egrpc.Load("").Build(egrpc.WithAddr(fmt.Sprintf("etcd:///%s", serviceName)))
		client = c.creator(grpcConn)
		c.clientMap.Store(serviceName, client)
	}

	return client
}
