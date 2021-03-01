package new

var (
	infrafc1 = &FileContent{
		FileName: "infra.go",
		Dir:      "internal/infra",
		Content: `package infra

import (
	"sync"
	"github.com/google/wire"
	"github.com/jukylin/esim/container"
	"github.com/jukylin/esim/mysql"
	"github.com/jukylin/esim/grpc"
	"github.com/jukylin/esim/pkg/uid"
	"github.com/jukylin/esim/pkg/validate"
	"{{.ProPath}}{{.ServerName}}/internal/infra/repo"
)

// Do not change the function name and var name
//  infraOnce
//  onceInfra
//  infraSet
//  NewInfra

var infraOnce sync.Once
var onceInfra *Infra

type Infra struct {
	*container.Esim

	DB *mysql.Client

	GrpcClient *grpc.Client

	Validate validate.ValidateRepo

	Uid uid.UIDRepo

	UserRepo repo.UserRepo
}

//nolint:deadcode,unused,varcheck
var infraSet = wire.NewSet(
	wire.Struct(new(Infra), "*"),
	provideDb,
	provideValidate,
	provideUid,
	provideUserRepo,
)


func NewInfra() *Infra {
	infraOnce.Do(func() {
		esim  := container.NewEsim()
		onceInfra = initInfra(esim, provideGrpcClient())
	})

	return onceInfra
}

func NewStubsInfra(grpcClient *grpc.Client) *Infra {
	infraOnce.Do(func() {
		esim  := container.NewEsim()
		onceInfra = initInfra(esim, grpcClient)
	})

	return onceInfra
}

// Close close the infra when app stop
func (infraer *Infra) Close()  {
	infraer.DB.Close()
}

func (infraer *Infra) HealthCheck() []error {
	var errs []error

	dbErrs := infraer.DB.Ping()
	if dbErrs != nil{
		errs = append(errs, dbErrs...)
	}

	return errs
}


func provideDb() *mysql.Client {
	return mysqlClent := mysql.NewClient()
}


func provideUserRepo(esim *container.Esim) repo.UserRepo {
	return repo.NewDBUserRepo(esim.Logger)
}

func provideValidate() validate.ValidateRepo {
	return validate.NewValidateRepo()
}

func provideUid() uid.UIDRepo {
	return uid.NewUIDRepo()
}

func provideGrpcClient() *grpc.Client {
	return grpc.NewClient(grpc.NewClientOptions())
}

`,
	}

	infrafc2 = &FileContent{
		FileName: "wire.go",
		Dir:      "internal/infra",
		Content: `//+build wireinject

package infra

import (
	"github.com/google/wire"
	"github.com/jukylin/esim/grpc"
	"github.com/jukylin/esim/container"
)


func initInfra(esim *container.Esim,grpc *grpc.Client) *Infra {
	wire.Build(infraSet)
	return nil
}
`,
	}

	infrafc3 = &FileContent{
		FileName: "wire_gen.go",
		Dir:      "internal/infra",
		Content: `// Code generated by Wire. DO NOT EDIT.

//go:generate wire
//+build !wireinject

package infra

import (
	"github.com/jukylin/esim/container"
	"github.com/jukylin/esim/grpc"
)

// Injectors from wire.go:
func initInfra(esim *container.Esim, grpc2 *grpc.Client) *Infra {
	mysqlClient := provideDb()
	userRepo := provideUserRepo(esim)
	infra := &Infra{
		Esim:     esim,
		DB:       mysqlClient,
		GrpcClient: grpc2,
		UserRepo: userRepo,
	}
	return infra
}
`,
	}
)

func initInfraFiles() {
	Files = append(Files, infrafc1, infrafc2, infrafc3)
}
