package zcfg

import (
	"flag"
	"reflect"
)

type Option interface {
	apply(*Loader)
}

type optionFunc func(*Loader)

func (fn optionFunc) apply(l *Loader) { fn(l) }

func FromFile(path string) Option {
	return optionFunc(func(l *Loader) {
		l.cfgPath = path
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
