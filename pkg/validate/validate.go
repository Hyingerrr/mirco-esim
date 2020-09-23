package validate

import (
	"reflect"
	"strings"

	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	zh_translations "github.com/go-playground/validator/v10/translations/zh"
	"github.com/pkg/errors"
)

//验证服务接口
type ValidateRepo interface {
	SetTagName(name string)
	ValidateStruct(i interface{}) error
}

//验证实例
type Validate struct {
	validate *validator.Validate
	trans    ut.Translator
}

func NewValidateRepo() ValidateRepo {
	v := new(Validate)
	v.validate = validator.New()
	/*中文翻译*/
	zht := zh.New()
	uni := ut.New(zht, zht)
	v.trans, _ = uni.GetTranslator("zh")
	//语言使用中文
	_ = zh_translations.RegisterDefaultTranslations(v.validate, v.trans)
	// 设置yaml TAG name
	v.SetTagName("mapstructure")
	return v
}

//设置校验返回信息
func (v *Validate) SetTagName(name string) {
	v.validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		tagName := strings.SplitN(fld.Tag.Get(name), ",", 2)[0]
		if tagName == "-" {
			return ""
		}
		return tagName
	})
}

func (v *Validate) ValidateStruct(i interface{}) error {
	err := v.validate.Struct(i)
	if err != nil {
		if _, ok := err.(*validator.InvalidValidationError); ok {
			return errors.Wrapf(err, "结构体规则配置校验失败:[%s]", err)
		}
		for _, err := range err.(validator.ValidationErrors) {
			return errors.New(err.Translate(v.trans))
		}
	}
	return nil
}
