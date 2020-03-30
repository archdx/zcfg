package zconf

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

const (
	tagEnv  = "env"
	tagFlag = "flag"
)

const flagCfgPath = "c"

var timeDurationType = reflect.TypeOf(time.Duration(0))

type Loader struct {
	cfg interface{}

	cfgPath         string
	cfgPathOverride string

	cfgFileDecoder func(r io.Reader, dst interface{}) error

	flagSet *flag.FlagSet
}

func (l *Loader) Load() error {
	if l.useFlags() && !l.flagSet.Parsed() {
		l.flagSet.Parse(os.Args[1:])
	}

	if l.useConfigFile() {
		err := l.initConfigFromFile(l.getConfigPath())
		if err != nil {
			return err
		}
	}

	return l.overrideConfig("", "", l.cfg, nil)
}

func (l *Loader) FromYAML(defaultPath string) *Loader {
	l.cfgPath = defaultPath
	l.cfgFileDecoder = func(r io.Reader, dst interface{}) error {
		return yaml.NewDecoder(r).Decode(dst)
	}

	return l
}

func (l *Loader) FromJSON(defaultPath string) *Loader {
	l.cfgPath = defaultPath
	l.cfgFileDecoder = func(r io.Reader, dst interface{}) error {
		return json.NewDecoder(r).Decode(dst)
	}

	return l
}

func (l *Loader) WithFlags(flagSet *flag.FlagSet) *Loader {
	l.flagSet = flagSet
	l.flagSet.StringVar(&l.cfgPathOverride, flagCfgPath, "", "config path")

	l.setupFlagSet("", reflect.TypeOf(l.cfg))

	return l
}

func (l *Loader) initConfigFromFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	defer f.Close()

	return l.cfgFileDecoder(f, l.cfg)
}

func (l *Loader) overrideConfig(flagPath string, env string, node interface{}, before func()) error {
	v := reflect.ValueOf(node)
	for ; v.Kind() == reflect.Ptr; v = v.Elem() {
		if !v.IsNil() {
			continue
		}

		prev := v
		next := reflect.New(v.Type().Elem())
		if before != nil {
			_before := before
			before = func() { _before(); prev.Set(next) }
		} else {
			before = func() { prev.Set(next) }
		}

		v = next
	}

	if v.Kind() == reflect.Struct {
		for i := 0; i < v.Type().NumField(); i++ {
			field := v.Field(i)
			if !field.CanSet() {
				continue
			}

			env, _ = v.Type().Field(i).Tag.Lookup(tagEnv)
			fval, _ := v.Type().Field(i).Tag.Lookup(tagFlag)

			nextFlagPath := buildFlagPath(flagPath, fval)

			err := l.overrideConfig(nextFlagPath, env, field.Addr().Interface(), before)
			if err != nil {
				return err
			}
		}

		return nil
	}

	value := l.lookupOverrideValue(flagPath, env)
	if value == "" {
		return nil
	}

	if before != nil {
		before()
	}

	switch v.Kind() {
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		if v.Type() == timeDurationType {
			dur, err := time.ParseDuration(value)
			if err != nil {
				return err
			}

			v.SetInt(int64(dur))
			break
		}

		num, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}

		v.SetInt(num)

	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		num, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}

		v.SetUint(num)

	case reflect.Float64, reflect.Float32:
		num, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}

		v.SetFloat(num)

	case reflect.Bool:
		bval, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}

		v.SetBool(bval)

	case reflect.String:
		v.SetString(value)

	case reflect.Slice:
		const sep = ","
		vals := strings.Split(value, sep)

		v.Set(reflect.MakeSlice(v.Type(), 0, len(vals)))

		switch t := v.Interface().(type) {
		case []string:
			for _, val := range vals {
				if len(val) > 0 {
					v.Set(reflect.Append(v, reflect.ValueOf(val)))
				}
			}

		default:
			return fmt.Errorf("unsupported slice type: %T", t)
		}

	default:
		return fmt.Errorf("cannot set field with kind: %s", v.Kind())
	}

	return nil
}

func (l *Loader) lookupOverrideValue(flagName, env string) string {
	if l.useFlags() &&
		flagName != "" && flagName != "-" {
		if flag := l.flagSet.Lookup(flagName); flag != nil {
			val := flag.Value.String()
			if val != "" {
				return val
			}
		}
	}

	if env != "" && env != "-" {
		return os.Getenv(env)
	}

	return ""
}

func (l *Loader) setupFlagSet(flagPath string, node reflect.Type) {
	for ; node.Kind() == reflect.Ptr; node = node.Elem() {
	}

	switch node.Kind() {
	case reflect.Struct:
		for i := 0; i < node.NumField(); i++ {
			field := node.Field(i)
			if fval, ok := field.Tag.Lookup(tagFlag); ok {
				l.setupFlagSet(buildFlagPath(flagPath, fval), field.Type)
			}
		}

	default:
		l.flagSet.String(flagPath, "", "")
	}
}

func (l *Loader) useConfigFile() bool {
	return l.cfgFileDecoder != nil
}

func (l *Loader) useFlags() bool {
	return l.flagSet != nil
}

func (l *Loader) getConfigPath() string {
	if l.cfgPathOverride != "" {
		return l.cfgPathOverride
	}

	return l.cfgPath
}

func buildFlagPath(keys ...string) string {
	const sep = "."

	var arr []string
	for _, k := range keys {
		if k != "" {
			arr = append(arr, k)
		}
	}

	return strings.Join(arr, sep)
}

func New(cfg interface{}) *Loader {
	return &Loader{
		cfg: cfg,
	}
}
