package zcfg

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	tagEnv   = "env"
	tagFlag  = "flag"
	tagUsage = "usage"
)

var timeDurationType = reflect.TypeOf(time.Duration(0))

type Loader struct {
	cfg interface{}

	cfgPath         string
	cfgPathOverride string

	flagSet *flag.FlagSet
}

func (l *Loader) Load() error {
	if l.useFlags() && !l.flagSet.Parsed() {
		l.flagSet.Parse(os.Args[1:])
	}

	if cfgPath := l.getConfigPath(); cfgPath != "" {
		err := l.initConfigFromFile(cfgPath)
		if err != nil {
			return err
		}
	}

	return l.overrideConfig(traverseContext{}, reflect.ValueOf(l.cfg), nil)
}

func (l *Loader) initConfigFromFile(path string) error {
	ext := strings.TrimPrefix(filepath.Ext(path), ".")

	decodeFunc, ok := fileDecoders[ext]
	if !ok {
		return fmt.Errorf("unsupported file extension: %s", ext)
	}

	f, err := os.Open(path)
	if err != nil {
		return err
	}

	defer f.Close()

	return decodeFunc(f, l.cfg)
}

type traverseContext struct {
	Tags     reflect.StructTag
	FullFlag string
}

func (l *Loader) overrideConfig(tctx traverseContext, node reflect.Value, bindNode func()) error {
	for ; node.Kind() == reflect.Ptr; node = node.Elem() {
		if !node.IsNil() {
			continue
		}

		prev := node
		next := reflect.New(prev.Type().Elem())

		if bindNode != nil {
			_bindNode := bindNode
			bindNode = func() { _bindNode(); prev.Set(next) }
		} else {
			bindNode = func() { prev.Set(next) }
		}

		node = next
	}

	if node.Kind() == reflect.Struct {
		for i := 0; i < node.Type().NumField(); i++ {
			field := node.Field(i)
			if !field.CanSet() {
				continue
			}

			tags := node.Type().Field(i).Tag

			flag, _ := tags.Lookup(tagFlag)

			err := l.overrideConfig(traverseContext{
				Tags:     tags,
				FullFlag: joinFlags(tctx.FullFlag, flag),
			}, field.Addr(), bindNode)

			if err != nil {
				return err
			}
		}

		return nil
	}

	value := l.lookupOverrideValue(tctx)
	if value == "" {
		return nil
	}

	switch node.Kind() {
	case reflect.Int, reflect.Int64, reflect.Int32, reflect.Int16, reflect.Int8:
		if node.Type() == timeDurationType {
			dur, err := time.ParseDuration(value)
			if err != nil {
				return err
			}

			node.SetInt(int64(dur))
			break
		}

		num, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}

		node.SetInt(num)

	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Uint16, reflect.Uint8:
		num, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}

		node.SetUint(num)

	case reflect.Float64, reflect.Float32:
		num, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}

		node.SetFloat(num)

	case reflect.Bool:
		bval, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}

		node.SetBool(bval)

	case reflect.String:
		node.SetString(value)

	case reflect.Slice:
		const sep = ","
		vals := strings.Split(value, sep)

		node.Set(reflect.MakeSlice(node.Type(), 0, len(vals)))

		switch t := node.Interface().(type) {
		case []string:
			for _, val := range vals {
				if len(val) > 0 {
					node.Set(reflect.Append(node, reflect.ValueOf(val)))
				}
			}

		default:
			return fmt.Errorf("unsupported slice type: %T", t)
		}

	default:
		return fmt.Errorf("cannot set field with kind: %s", node.Kind())
	}

	if bindNode != nil {
		bindNode()
	}

	return nil
}

func (l *Loader) lookupOverrideValue(tctx traverseContext) string {
	if l.useFlags() && tctx.FullFlag != "" {
		flag := l.flagSet.Lookup(tctx.FullFlag)
		if flag != nil {
			val := flag.Value.String()
			if val != "" {
				return val
			}
		}
	}

	if env, ok := tctx.Tags.Lookup(tagEnv); ok {
		return os.Getenv(env)
	}

	return ""
}

func (l *Loader) setupFlagSet(tctx traverseContext, node reflect.Type) {
	for ; node.Kind() == reflect.Ptr; node = node.Elem() {
	}

	switch node.Kind() {
	case reflect.Struct:
		for i := 0; i < node.NumField(); i++ {
			field := node.Field(i)
			tags := field.Tag

			if flag, ok := tags.Lookup(tagFlag); ok || field.Anonymous {
				l.setupFlagSet(traverseContext{
					Tags:     tags,
					FullFlag: joinFlags(tctx.FullFlag, flag),
				}, field.Type)
			}
		}

	default:
		usage, _ := tctx.Tags.Lookup(tagUsage)

		l.flagSet.String(tctx.FullFlag, "", usage)
	}
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

func joinFlags(keys ...string) string {
	const sep = "."

	var arr []string
	for _, k := range keys {
		if k != "" {
			arr = append(arr, k)
		}
	}

	return strings.Join(arr, sep)
}

func New(cfg interface{}, opts ...Option) *Loader {
	loader := Loader{
		cfg: cfg,
	}

	for _, opt := range opts {
		opt.apply(&loader)
	}

	return &loader
}
