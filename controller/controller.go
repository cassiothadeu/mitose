package controller

import (
	"context"
	"fmt"
	"time"

	"github.com/luizalabs/mitose/k8s"
)

type Metrics map[string]string

type Colector interface {
	GetMetrics() (Metrics, error)
}

type Cruncher interface {
	CalcDesiredReplicas(Metrics) (int, error)
}

type Controller struct {
	colector   Colector
	cruncher   Cruncher
	namespace  string
	deployment string
}

func (c *Controller) Run(ctx context.Context, interval time.Duration) error {
	fmt.Printf("start controller for deployment %s (namespace %s)\n", c.deployment, c.namespace)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
			if err := c.Exec(); err != nil {
				return err
			}
		}
	}
}

func (c *Controller) Exec() error {
	m, err := c.colector.GetMetrics()
	if err != nil {
		return err
	}
	desiredReplicas, err := c.cruncher.CalcDesiredReplicas(m)
	if err != nil {
		return err
	}
	fmt.Printf(
		"Desired replicas %d for deployment %s (namespace %s)\n",
		desiredReplicas,
		c.deployment,
		c.namespace,
	)
	return c.Autoscale(desiredReplicas)
}

func (c *Controller) Autoscale(desiredPods int) error {
	return k8s.UpdateReplicasCount(c.namespace, c.deployment, desiredPods)
}

func NewController(colector Colector, cruncher Cruncher, namespace, deployment string) *Controller {
	return &Controller{
		colector:   colector,
		cruncher:   cruncher,
		namespace:  namespace,
		deployment: deployment,
	}
}