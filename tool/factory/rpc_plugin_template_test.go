package factory

import (
	"testing"

	"github.com/Hyingerrr/mirco-esim/pkg"
	"github.com/Hyingerrr/mirco-esim/pkg/templates"

	"github.com/stretchr/testify/assert"
)

func TestExecuteRPCPluginTemplate(t *testing.T) {
	Field1 := pkg.Field{}
	Field1.Field = "a int"

	Field2 := pkg.Field{}
	Field2.Field = "b string"

	fields := make([]pkg.Field, 0)
	fields = append(fields, Field1, Field2)

	data := struct {
		StructName string
		Fields     []pkg.Field
	}{
		testStructName,
		fields,
	}

	tpl := templates.NewTextTpl()
	tmpl, err := tpl.Execute("rpc_plugin", rpcPluginTemplate, data)
	assert.Nil(t, err)
	assert.NotEmpty(t, tmpl)
}
