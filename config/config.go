package config

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config App config struct
type Config struct {
	SiteURL  string
	Server   ServerConfig
	Mongo    MongoConfig
	Metrics  Metrics
	Logger   Logger
	Jaeger   Jaeger
	AssetClient AssetClient
	//Auth     []AuthConfig
	//JWT      JWT
}

// ServerConfig is Server config struct
type ServerConfig struct {
	AppVersion        string
	Port              string
	Mode              string // if mode is not "Production", the reflection is be enabled
	MaxConnectionIdle time.Duration
	Timeout           time.Duration
	MaxConnectionAge  time.Duration
	Time              time.Duration
}

// Logger config
type Logger struct {
	DisableCaller     bool
	DisableStacktrace bool
	Encoding          string
	Level             string
}


type MongoConfig struct {
	Uri string
}

// Metrics config
type Metrics struct {
	URL         string
	ServiceName string
}

// Jaeger is config for Jaeger
type Jaeger struct {
	Host          string
	ServiceName   string
	LogSpans      bool
	SamplingRatio float64
}

type AssetClient struct {
	ServerAddr string
}

/* type AuthConfig struct {
	Method string
	Role   []string
}

type JWT struct {
	ExpIn        int
	JwtSecretKey string
} */

// LoadViperConfig file from given path
func LoadViperConfig() (*viper.Viper, error) {
	v := viper.New()

	v.SetConfigName("./config/config")
	v.AddConfigPath(".")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// lists all files in ./config
			files, err := os.ReadDir("./config")
			if err != nil {
				var filenames []string
				for _, file := range files {
					if file.Type().IsRegular() {
						filenames = append(filenames, file.Name())
					}
				}
				log.Println(filenames)
			}

			return nil, fmt.Errorf("config file ['./config/config'] not found")
		}
		return nil, err
	}

	return v, nil
}

// LoadViperConfigFromString file from given path
func LoadViperConfigFromString(input string) (*viper.Viper, error) {
	v := viper.New()
	v.SetConfigType("yaml")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	if err := v.ReadConfig(strings.NewReader(input)); err != nil {
		return nil, err
	}

	return v, nil
}

// ParseConfig is to parse config file with Viper
func ParseConfig(v *viper.Viper) (*Config, error) {
	var c Config

	err := v.Unmarshal(&c)
	if err != nil {
		log.Printf("unable to decode into struct, %v", err)
		return nil, err
	}

	return &c, nil
}

// GetServiceConfig is to get config
func GetServiceConfig() (*Config, error) {
	cfgViper, err := LoadViperConfig()
	if err != nil {
		return nil, err
	}

	cfg, err := ParseConfig(cfgViper)
	if err != nil {
		return nil, err
	}

	log.Println("Success parsed config - App version:", cfg.Server.AppVersion)
	return cfg, nil
}

func (c Config) GetTracerName() string {
	return fmt.Sprintf("%s-tracer", c.Jaeger.ServiceName)
}
