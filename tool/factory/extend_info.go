package factory

type extendFieldInfo struct {
	ftype string

	name string

	importPath string
}

func newLoggerFieldInfo() extendFieldInfo {
	efi := extendFieldInfo{
		ftype:      "log.Logger",
		name:       "logger",
		importPath: "github.com/Hyingerrr/mirco-esim/log",
	}

	return efi
}

func newConfigFieldInfo() extendFieldInfo {
	efi := extendFieldInfo{
		ftype:      "config.Config",
		name:       "conf",
		importPath: "github.com/Hyingerrr/mirco-esim/config",
	}

	return efi
}
