package args

import (
	"strings"
)

type option = func(*FlagSet)

func Env() option { return EnvPrefix("") }
func EnvPrefix(prefix string) option {
	return func(f *FlagSet) {
		if prefix != "" {
			if strings.HasSuffix(prefix, "_") {
				prefix = strings.ToUpper(prefix)
			} else {
				prefix = strings.ToUpper(prefix) + "_"
			}
		}
		f.envPrefix = &prefix
	}
}

func Json(filepath *string) option { return UseProvider(NewJsonProvider(filepath)) }
func Yaml(filepath *string) option { return UseProvider(NewYamlProvider(filepath)) }
func UseProvider(provider Provider) option {
	return func(f *FlagSet) {
		f.providers = append(f.providers, provider)
	}
}
