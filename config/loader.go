package config

import (
	"fmt"
	"github.com/creasty/defaults"
	"github.com/spf13/viper"
	"reflect"
	"strings"
)

type Loader interface {
	Bind(properties ...Properties)
}

type ViperLoader struct {
	option   Option
	debugLog func(msgFormat string, args ...interface{})
	viper    *viper.Viper
}

func NewLoader(option Option, debugLog func(msgFormat string, args ...interface{})) *ViperLoader {
	setDefaultOption(&option)
	if debugLog == nil {
		debugLog = func(msgFormat string, args ...interface{}) {
			_, _ = fmt.Printf(msgFormat+"\n", args...)
		}
	}
	return &ViperLoader{
		option:   option,
		debugLog: debugLog,
		viper:    loadViper(option, debugLog),
	}
}

func (l *ViperLoader) Bind(propertiesList ...Properties) {
	for _, properties := range propertiesList {
		propertiesName := reflect.TypeOf(properties).String()
		// Run pre-binding life cycle
		if propsPreBind, ok := properties.(PropertiesPreBinding); ok {
			propsPreBind.PreBinding()
		}

		// Unmarshal from config file
		if err := l.viper.UnmarshalKey(properties.Prefix(), properties); err != nil {
			panic(fmt.Sprintf("[GoLib-error] Fatal error when binding config key [%s] to [%s]: %v",
				properties.Prefix(), propertiesName, err))
		}

		// Set default value if its missing
		l.setDefaults(propertiesName, properties)

		// Run post-binding life cycle
		if propsPostBind, ok := properties.(PropertiesPostBinding); ok {
			propsPostBind.PostBinding()
		}
		l.debugLog("[GoLib-debug] Properties [%s] loaded with prefix [%s]", propertiesName, properties.Prefix())
	}
}

func (l *ViperLoader) setDefaults(propertiesName string, properties Properties) {
	if err := defaults.Set(properties); err != nil {
		panic(fmt.Sprintf("[GoLib-error] Fatal error when set default values for [%s]: %v", propertiesName, err))
	}
}

func loadViper(option Option, debugLog func(msgFormat string, args ...interface{})) *viper.Viper {
	debugActiveProfiles := strings.Join(option.ActiveProfiles, ", ")
	debugPaths := strings.Join(option.ConfigPaths, ", ")
	debugLog("[GoLib-debug] Loading active profiles [%s] in paths [%s] with format [%s]",
		debugActiveProfiles, debugPaths, option.ConfigFormat)

	vi := viper.New()
	vi.AutomaticEnv()
	vi.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	for _, activeProfile := range option.ActiveProfiles {
		vi.SetConfigName(activeProfile)
		vi.SetConfigType(option.ConfigFormat)
		for _, path := range option.ConfigPaths {
			vi.AddConfigPath(path)
		}
		if err := vi.MergeInConfig(); err != nil {
			if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
				panic(fmt.Sprintf("[GoLib-error] Fatal error when read active profile [%s] in paths [%s]: %v",
					activeProfile, debugPaths, err))
			}
			debugLog("[GoLib-debug] Config file not found when read active profile [%s] in paths [%s]",
				activeProfile, debugPaths)
			continue
		}
		debugLog("[GoLib-debug] Active profile [%s] was loaded", activeProfile)
	}

	// Workaround solution because viper does not
	// treat env vars the same as other config
	// See https://github.com/spf13/viper/issues/188#issuecomment-399518663
	for _, key := range vi.AllKeys() {
		val := vi.Get(key)
		vi.Set(key, val)
	}
	return vi
}
