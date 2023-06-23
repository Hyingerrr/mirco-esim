//nolint
package infra

import (
	"sync"

	"github.com/Hyingerrr/mirco-esim/container"
	"github.com/Hyingerrr/mirco-esim/redis"

	"github.com/google/wire"
)

var infraOnce sync.Once
var onceInfra *Infra

type Infra struct {

	// Esim
	*container.Esim

	// redis
	Redis *redis.Client
}

var infraSet = wire.NewSet(
	wire.Struct(new(Infra), "*"),
)

func NewInfra() *Infra {
	infraOnce.Do(func() {
	})

	return onceInfra
}

// Close close the infra when app stop
func (inf *Infra) Close() {
}

func (inf *Infra) HealthCheck() []error {
	var errs []error
	return errs
}
