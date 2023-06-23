package factory

//nolint:lll
var rpcPluginTemplate = `package main

import (
	"unsafe"
	"sort"
	"encoding/json"
	"reflect"
	"strings"
	"github.com/hashicorp/go-plugin"
	"github.com/Hyingerrr/mirco-esim/tool/factory"
	"github.com/Hyingerrr/mirco-esim/pkg"
	"github.com/Hyingerrr/mirco-esim/log"
	"github.com/Hyingerrr/mirco-esim/pkg/file-dir"
)


type InitFieldsReturn struct{
	Fields []string
	SpecFields pkg.Fields
}

type Return struct{
	Fields pkg.Fields
	Size int
}

func (ModelImp) Sort() string {

	{{.StructName | tolower}} := {{.StructName}}{}

	originSize := unsafe.Sizeof({{.StructName | tolower}})

	getType := reflect.TypeOf({{.StructName | tolower}})

	var fields pkg.Fields

	{{range $i, $field := .Fields}}
	field{{$i}} := pkg.Field{}
	field{{$i}}.Name = "{{$field.Name}}"
	field{{$i}}.Field = "{{$field.Field}}"
	field{{$i}}.Size = int(getType.Field({{$i}}).Type.Size())
	fields = append(fields, field{{$i}})

	{{end}}

	sort.Sort(fields)

	re := &Return{}
	re.Fields = fields
	re.Size = int(originSize)

	by, _ := json.Marshal(re)
	return string(by)

}


func (ModelImp) InitField() string {
		{{.StructName | tolower}} := {{.StructName}}{}

		initReturn := &InitFieldsReturn{}
	 	fields := &pkg.Fields{}

		getType := reflect.TypeOf({{.StructName | tolower}})

		writer := filedir.NewEsimWriter()
		rpcPlugin := factory.NewRPCPluginStructField(writer, log.NewLogger())
		structFields := rpcPlugin.GenInitFieldStr(getType, "{{.StructName | snakeToCamelLower | shorten}}", "{{.StructName | snakeToCamelLower | shorten}}", fields)

		initReturn.SpecFields = *fields
		initReturn.Fields = structFields
		j, _ := json.Marshal(initReturn)
		return string(j)
	}




type ModelImp struct{}

func main() {

	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: factory.HandshakeConfig,
		Plugins: map[string]plugin.Plugin{
			"model": &factory.ModelPlugin{Impl: &ModelImp{}},
		},

		// A non-nil value here enables gRPC serving for this plugin...
		GRPCServer: plugin.DefaultGRPCServer,
	})
}
`
