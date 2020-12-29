package config

const (
	DeployEnvDev  DeployEnv = "dev" // local develop
	DeployEnvQa   DeployEnv = "qa"  // test
	DeployEnvPre  DeployEnv = "pre" // pre release
	DeployEnvProd DeployEnv = "pro" // production
)

type DeployEnv string

func (env DeployEnv) GetEnv() {

}
