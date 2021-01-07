package xenv

const (
	DeployEnvDev  DeployEnv = "dev" // local develop
	DeployEnvQa   DeployEnv = "qa"  // test
	DeployEnvPre  DeployEnv = "pre" // pre release
	DeployEnvProd DeployEnv = "pro" // production
)

// ----------------------------------------
//  运行环境类型定义
// ----------------------------------------
type (
	DeployEnv string

	RunMode struct {
		mode DeployEnv
	}
)

var (
	env DeployEnv
)

func SetRunMode(mode string) *RunMode {
	env = DeployEnv(mode)
	return &RunMode{mode: env}
}

func (run RunMode) GetEnv() DeployEnv {
	return run.mode
}

func (run *RunMode) IsValid() bool {
	switch run.mode {
	case DeployEnvDev, DeployEnvQa, DeployEnvPre, DeployEnvProd:
		return true
	}

	return false
}

func IsPro() bool {
	return env == DeployEnvProd
}

func IsDev() bool {
	return env == DeployEnvDev
}

func IsQa() bool {
	return env == DeployEnvQa
}

func IsPre() bool {
	return env == DeployEnvPre
}
