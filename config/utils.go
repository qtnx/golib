package config

import (
	"github.com/mitchellh/mapstructure"
	"gitlab.com/golibs-starter/golib/utils"
	"reflect"
)

func MapStructurePlaceholderValueHook() mapstructure.DecodeHookFunc {
	return func(f reflect.Type, t reflect.Type, data interface{}) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		return utils.ReplacePlaceholder(data.(string))
	}
}
