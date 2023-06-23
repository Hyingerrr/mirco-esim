package domainfile

import (
	"bytes"
	"testing"
	"text/template"

	"github.com/Hyingerrr/mirco-esim/pkg"
	"github.com/Hyingerrr/mirco-esim/pkg/templates"

	"github.com/stretchr/testify/assert"
)

func TestDaoTemplate(t *testing.T) {
	tmpl, err := template.New("dao_template").Funcs(templates.EsimFuncMap()).
		Parse(daoTemplate)
	assert.Nil(t, err)

	var imports pkg.Imports
	imports = append(imports, pkg.Import{Name: "time", Path: "time"},
		pkg.Import{Name: "sync", Path: "sync"})

	var buf bytes.Buffer
	daoTmp := newDaoTpl(userStructName)
	daoTmp.Imports = imports
	daoTmp.DataBaseName = database
	daoTmp.TableName = userTable
	daoTmp.PriKeyType = "int"

	err = tmpl.Execute(&buf, daoTmp)
	assert.Nil(t, err)
}
