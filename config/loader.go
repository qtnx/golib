package config

import (
	"bytes"
	"fmt"
	"github.com/creasty/defaults"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
	"os"
	"reflect"
	"strings"
)

var keyDelimiter = "."

type Loader interface {
	Bind(properties ...Properties) error
}

type ViperLoader struct {
	viper            *viper.Viper
	option           Option
	properties       []Properties
	groupPropsConfig map[string]interface{}
}

func NewLoader(option Option, properties []Properties) (Loader, error) {
	setDefaultOption(&option)
	vi, err := loadViper(option, properties)
	if err != nil {
		return nil, err
	}
	return &ViperLoader{
		viper:            vi,
		option:           option,
		properties:       properties,
		groupPropsConfig: groupPropertiesValues(vi, properties),
	}, nil
}

func (l *ViperLoader) Bind(propertiesList ...Properties) error {
	for _, props := range propertiesList {
		propsName := reflect.TypeOf(props).String()
		// Run pre-binding life cycle
		if propsPreBind, ok := props.(PropertiesPreBinding); ok {
			if err := propsPreBind.PreBinding(); err != nil {
				return err
			}
		}

		decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
			Metadata:         nil,
			Result:           props,
			WeaklyTypedInput: true,
			DecodeHook: mapstructure.ComposeDecodeHookFunc(
				mapstructure.StringToTimeDurationHookFunc(),
				mapstructure.StringToSliceHookFunc(","),
			),
		})
		if err != nil {
			return fmt.Errorf("[GoLib-error] Fatal error when init decoder for key [%s] to [%s]: %v",
				props.Prefix(), propsName, err)
		}
		if err := decoder.Decode(l.groupPropsConfig[props.Prefix()]); err != nil {
			return fmt.Errorf("[GoLib-error] Fatal error when binding config key [%s] to [%s]: %v",
				props.Prefix(), propsName, err)
		}

		// Run post-binding life cycle
		if propsPostBind, ok := props.(PropertiesPostBinding); ok {
			if err := propsPostBind.PostBinding(); err != nil {
				return err
			}
		}
		l.option.DebugFunc("[GoLib-debug] LoggingProperties [%s] loaded with prefix [%s]", propsName, props.Prefix())
	}
	return nil
}

func setDefaults(propertiesName string, properties Properties) error {
	if err := defaults.Set(properties); err != nil {
		return fmt.Errorf("[GoLib-error] Fatal error when set default values for [%s]: %v", propertiesName, err)
	}
	return nil
}

func loadViper(option Option, propertiesList []Properties) (*viper.Viper, error) {
	option.DebugFunc("[GoLib-debug] Loading active profiles [%s] in paths [%s] with format [%s]",
		strings.Join(option.ActiveProfiles, ", "), strings.Join(option.ConfigPaths, ", "), option.ConfigFormat)

	vi := viper.NewWithOptions(viper.KeyDelimiter(keyDelimiter))
	vi.SetEnvKeyReplacer(strings.NewReplacer(keyDelimiter, "_"))
	vi.AutomaticEnv()

	if err := discoverDefaultValue(vi, propertiesList, option.DebugFunc); err != nil {
		return nil, err
	}

	if err := discoverActiveProfiles(vi, option); err != nil {
		return nil, err
	}

	// High priority for environment variable.
	// This is workaround solution because viper does not
	// treat env vars the same as other config
	// See https://github.com/spf13/viper/issues/188#issuecomment-399518663
	//
	// Notes: Currently vi.AllKeys() doesn't support key for array item, such as: foo.bar.0.username,
	// so environment variable cannot overwrite these values, replace placeholder also not working
	// (using PropertiesPostBinding to replace placeholder as a workaround solution).
	// TODO Improve it or wait for viper in next version
	//for _, key := range vi.AllKeys() {
	//	val := vi.Get(key)
	//	if newVal, err := ReplacePlaceholderValue(val); err != nil {
	//		return nil, err
	//	} else {
	//		val = newVal
	//	}
	//	err := vi.BindEnv(key)
	//	if err != nil {
	//		return nil, err
	//	}
	//}
	return vi, nil
}

func groupPropertiesValues(vi *viper.Viper, propertiesList []Properties) map[string]interface{} {
	allSettings := vi.AllSettings()
	group := make(map[string]interface{})
	for _, props := range propertiesList {
		m := deepSearchInMap(allSettings, props.Prefix())
		correctedVal, exists := correctSliceValues(vi, props.Prefix(), m)
		if exists {
			group[props.Prefix()] = correctedVal
		} else {
			group[props.Prefix()] = m
		}
	}
	return group
}

func correctSliceValues(vi *viper.Viper, prefix string, val interface{}) (interface{}, bool) {
	if slice, ok := val.([]interface{}); ok {
		for k, v := range slice {
			correctedVal, exists := correctSliceValues(vi, fmt.Sprintf("%s%s%d", prefix, keyDelimiter, k), v)
			if exists {
				slice[k] = correctedVal
			}
		}
	} else if m, ok := val.(map[interface{}]interface{}); ok {
		for k, v := range m {
			correctedVal, exists := correctSliceValues(vi, fmt.Sprintf("%s%s%s", prefix, keyDelimiter, k), v)
			if exists {
				m[k] = correctedVal
			}
		}
	} else if m, ok := val.(map[string]interface{}); ok {
		for k, v := range m {
			correctedVal, exists := correctSliceValues(vi, fmt.Sprintf("%s%s%s", prefix, keyDelimiter, k), v)
			if exists {
				m[k] = correctedVal
			}
		}
	} else {
		if correctedVal := vi.Get(prefix); correctedVal != nil {
			return correctedVal, true
		}
	}
	return nil, false
}

func deepSearchInMap(m map[string]interface{}, key string) map[string]interface{} {
	parts := strings.Split(key, keyDelimiter)
	for _, part := range parts {
		val, ok := m[part]
		if !ok {
			return make(map[string]interface{})
		}
		m, ok = val.(map[string]interface{})
		if !ok {
			return make(map[string]interface{})
		}
	}
	return m
}

// discoverDefaultValue Discover default values for multiple properties at once
func discoverDefaultValue(vi *viper.Viper, propertiesList []Properties, debugFunc DebugFunc) error {
	for _, props := range propertiesList {
		propsName := reflect.TypeOf(props).String()

		// Set default value if its missing
		if err := setDefaults(propsName, props); err != nil {
			return err
		}

		// set default values in viper.
		// Viper needs to know if a key exists in order to override it.
		// https://github.com/spf13/viper/issues/188
		b, err := yaml.Marshal(convertSliceToNestedMap(strings.Split(props.Prefix(), keyDelimiter), props, nil))
		if err != nil {
			return err
		}
		vi.SetConfigType("yaml")
		if err := vi.MergeConfig(bytes.NewReader(b)); err != nil {
			return fmt.Errorf("[GoLib-error] Error when discover default value for properties [%s]: %v", propsName, err)
		}
		debugFunc("[GoLib-debug] Default value was discovered for properties [%s]", propsName)
	}
	return nil
}

// discoverActiveProfiles Discover values for multiple active profiles at once
func discoverActiveProfiles(vi *viper.Viper, option Option) error {
	debugPaths := strings.Join(option.ConfigPaths, ", ")
	for _, activeProfile := range option.ActiveProfiles {
		vi.SetConfigName(activeProfile)
		vi.SetConfigType(option.ConfigFormat)
		for _, path := range option.ConfigPaths {
			vi.AddConfigPath(path)
		}
		if err := vi.MergeInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				return fmt.Errorf("[GoLib-error] Error when read active profile [%s] in paths [%s]: %v",
					activeProfile, debugPaths, err)
			}
			return fmt.Errorf("[GoLib-debug] Config file not found when read active profile [%s] in paths [%s]",
				activeProfile, debugPaths)
		}
		option.DebugFunc("[GoLib-debug] Active profile [%s] was loaded", activeProfile)
	}
	return nil
}

func convertSliceToNestedMap(paths []string, endVal interface{}, inMap map[interface{}]interface{}) map[interface{}]interface{} {
	if inMap == nil {
		inMap = map[interface{}]interface{}{}
	}
	if len(paths) == 0 {
		return inMap
	}
	if len(paths) == 1 {
		inMap[paths[0]] = endVal
		return inMap
	}
	inMap[paths[0]] = convertSliceToNestedMap(paths[1:], endVal, map[interface{}]interface{}{})
	return inMap
}

// ReplacePlaceholderValue Replaces a value in placeholder format
// by new value configured in environment variable.
//
// Placeholder format: ${EXAMPLE_VAR}
func ReplacePlaceholderValue(val interface{}) (interface{}, error) {
	strVal, ok := val.(string)
	if !ok {
		return val, nil
	}
	// Make sure the value starts with ${ and end with }
	if !strings.HasPrefix(strVal, "${") || !strings.HasSuffix(strVal, "}") {
		return val, nil
	}
	key := strings.TrimSuffix(strings.TrimPrefix(strVal, "${"), "}")
	if len(key) == 0 {
		return nil, fmt.Errorf("invalid config placeholder format. Expected ${EX_ENV}, got [%s]", strVal)
	}
	res, present := os.LookupEnv(key)
	if !present {
		return nil, fmt.Errorf("mandatory env variable not found [%s]", key)
	}
	return res, nil
}
