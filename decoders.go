package zcfg

import (
	"encoding/json"
	"io"

	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v2"
)

var fileDecoders = map[string]FileDecoder{
	"yml":  yamlDecoder,
	"yaml": yamlDecoder,
	"json": jsonDecoder,
	"toml": tomlDecoder,
}

var (
	yamlDecoder = func(r io.Reader, dst interface{}) error { return yaml.NewDecoder(r).Decode(dst) }
	jsonDecoder = func(r io.Reader, dst interface{}) error { return json.NewDecoder(r).Decode(dst) }
	tomlDecoder = func(r io.Reader, dst interface{}) error { return toml.NewDecoder(r).Decode(dst) }
)

type FileDecoder func(r io.Reader, dst interface{}) error

func RegisterFileDecoder(ext string, dec FileDecoder) {
	if _, ok := fileDecoders[ext]; !ok {
		fileDecoders[ext] = dec
	}
}
