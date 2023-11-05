package args

import "strings"

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

func Json(path *string) option {
	return func(f *FlagSet) {
		f.providers = append(f.providers, (&JsonProvider{path: path}).Parse)
	}
}

func Yaml(path *string) option {
	return func(f *FlagSet) {
		f.providers = append(f.providers, (&YamlProvider{path: path}).Parse)
	}
}
