package zconf

import (
	"encoding/json"
	"flag"
	"io"
	"reflect"

	"gopkg.in/yaml.v2"
)

type Option interface {
	apply(*Loader)
}

type optionFunc func(*Loader)

func (fn optionFunc) apply(l *Loader) { fn(l) }

func FromYAML(defaultPath string) Option {
	return optionFunc(func(l *Loader) {
		l.cfgPath = defaultPath
		l.cfgFileDecoder = func(r io.Reader, dst interface{}) error {
			return yaml.NewDecoder(r).Decode(dst)
		}
	})
}

func FromJSON(defaultPath string) Option {
	return optionFunc(func(l *Loader) {
		l.cfgPath = defaultPath
		l.cfgFileDecoder = func(r io.Reader, dst interface{}) error {
			return json.NewDecoder(r).Decode(dst)
		}
	})
}

func UseFlags(flagSet *flag.FlagSet) Option {
	const flagCfgPath = "c"

	return optionFunc(func(l *Loader) {
		l.flagSet = flagSet
		l.flagSet.StringVar(&l.cfgPathOverride, flagCfgPath, "", "config path")

		l.setupFlagSet("", reflect.TypeOf(l.cfg))
	})
}
