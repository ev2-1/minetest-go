package minetest

import (
	"bytes"
	"errors"
	"github.com/spf13/pflag"
	"io"
	"log"
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
				log.Fatalf("Error decoding config overwrite: %s", err)
			}

			d = yaml.NewDecoder(buf)

			var overwrite any

			err = d.Decode(overwrite)
			if err != nil {
				log.Fatalf("Error decoding config overwrite '%s': %s", key, err)
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
		log.Fatalf("Failed to open config file '%s': %s", *configPath, err)
	}

	d := yaml.NewDecoder(f)
	err = d.Decode(&config)

	if err != nil {
		if errors.Is(err, io.EOF) {
			log.Printf("EOF while parsing config file '%s'. Ignoring configuration\n", *configPath)
		} else {
			log.Fatalf("Failed to parse config file '%s': %s", *configPath, err)
		}
	}

	if *verbose {
		log.Printf("Loaded %d configuration fields!", len(config))
	}
}

// ReloadConfig reloads the config and triggers Config Reload hooks
// may break some modules
func ReloadConfig() {
	LoadConfig()

	loadConfig()
}

func GetConfig(key string) any {
	LoadConfig()

	configMu.RLock()
	defer configMu.RUnlock()

	return config[key]
}

// ConfigVerbose is a helper function to indicate if verbose logging is turned on
func ConfigVerbose() bool {
	return GetConfig("verbose").(bool)
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
