package minetest

import (
	"bytes"
	"errors"
	"github.com/spf13/pflag"
	"io"
	"os"
	"sync"

	"github.com/go-yaml/yaml"
)

// Commandline Flags:
// ()
var (
	verbose        = pflag.BoolP("verbose", "v", false, "Turn on verbose logging mode")
	configPath     = pflag.StringP("config-path", "c", "config.yml", "Set path of config.yml file")
	confOverwrites = pflag.StringSliceP("config", "o", []string{}, "Overwrite configuration values at startup (gets overwritten by subsequent ReloadConfig calls)")
)

var (
	config   = make(map[string]any)
	configMu sync.RWMutex
)

var loadConfigOnce sync.Once

// LoadConfig ensures the config is loaded
func LoadConfig() {
	loadConfigOnce.Do(func() {
		pflag.Parse()
		loadConfig()

		configMu.Lock()
		defer configMu.Unlock()

		var buf *bytes.Buffer
		var d *yaml.Decoder

		// overlay commandline overwrites
		for _, v := range *confOverwrites {
			buf = bytes.NewBufferString(v)
			key, err := buf.ReadString(byte(':'))
			if err != nil {
				Loggers.Errorf("Error decoding config overwrite: %s", 1, err)
				os.Exit(1)
			}

			d = yaml.NewDecoder(buf)

			var overwrite any

			err = d.Decode(overwrite)
			if err != nil {
				Loggers.Errorf("Error decoding config overwrite '%s': %s", 1, key, err)
				os.Exit(1)
			}

			config[key] = overwrite
		}

		// special overwrites:
		if *verbose {
			config["verbose"] = any(*verbose)
		}
	})
}

// `loadConfig` forceloads the configuration
// DO NOT CONFUSE with `LoadConfig` (capital L)
func loadConfig() {
	f, err := os.OpenFile(*configPath, os.O_CREATE|os.O_RDWR, 0777)
	if err != nil {
		Loggers.Errorf("Failed to open config file '%s': %s", 1, *configPath, err)
		os.Exit(1)
	}

	d := yaml.NewDecoder(f)
	err = d.Decode(&config)

	if err != nil {
		if errors.Is(err, io.EOF) {
			Loggers.Defaultf("EOF while parsing config file '%s'. Ignoring configuration\n", 1, *configPath)
		} else {
			Loggers.Defaultf("Failed to parse config file '%s': %s", 1, *configPath, err)
		}
	}

	if *verbose {
		Loggers.Defaultf("Loaded %d configuration fields!", 1, len(config))
	}
}

// ReloadConfig reloads the config and triggers Config Reload hooks
// may break some modules
func ReloadConfig() {
	LoadConfig()

	loadConfig()
}

func MustGetConfig(key string) any {
	LoadConfig()

	configMu.RLock()
	defer configMu.RUnlock()

	return config[key]
}

func getConfig(key string) (val any, ok bool) {
	LoadConfig()

	configMu.RLock()
	defer configMu.RUnlock()

	val, ok = config[key]
	return
}

// Like GetConfig but does not return ok
func GetConfigV[K any](key string, d K) (val K) {
	val, _ = GetConfig[K](key, d)

	return
}

// Returns value which is is the config field if set or d if not
// ok is set if the config field existed
func GetConfig[K any](key string, d K) (val K, ok bool) {
	v, ok := getConfig(key)
	if !ok {
		return d, true
	}

	val, ok = v.(K)
	if !ok {
		Loggers.Defaultf("WARN: config field %s was requested as %T but is type %T!\n", 1, key, d, v)
		return d, false
	}

	return
}

// ConfigVerbose is a helper function to indicate if verbose logging is turned on
func ConfigVerbose() bool {
	v, _ := GetConfig("log-verbose", false)
	return v
}

// ForInConfig executes f `for k, v := range config`
// if any call of f results in a err != nil, err is returned
func ForInConfig(f func(k string, v any) error) (err error) {
	configMu.RLock()
	defer configMu.RUnlock()

	for k, v := range config {
		err = f(k, v)

		if err != nil {
			return
		}
	}

	return
}
