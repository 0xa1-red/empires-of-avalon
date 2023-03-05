package gamecluster

import "github.com/asynkron/protoactor-go/cluster"

var c *cluster.Cluster

func GetC() *cluster.Cluster {
	return c
}

func SetC(cc *cluster.Cluster) {
	c = cc
}
