package config

import (
	"errors"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"strings"
)

type (
	// AppConfig общая структура настроек приложения.
	AppConfig struct {
		Http    HttpServer `mapstructure:"http" json:"http" yaml:"http"`
		Sources Sources    `mapstructure:"sources" json:"sources" yaml:"sources"`
		Symbols []string   `mapstructure:"symbols" json:"symbols" yaml:"symbols"`
	}

	// Sources источники данных
	Sources struct {
		Binance string `mapstructure:"binance"  json:"binance" yaml:"binance"`
	}

	// HttpServer содержит настройки для создания HTTP сервера.
	HttpServer struct {
		Listen string `mapstructure:"listen"  json:"listen" yaml:"listen"    validate:"required,hostname_port" example:"0.0.0.0:8080"`
	}
)

const (
	configFlag = "config"
)

// Config - основной объект для работы с конфигурациями.
// Перед использованием необходимо выполнить Init.
var Config *AppConfig

// SetDefaults установка значений по умолчанию.
func (c *AppConfig) setDefaults(v *viper.Viper) {
	v.SetDefault("http_listen", "0.0.0.0:8080")
	v.SetDefault("sources_binance", "wss://stream.binance.com:9443")
}

func (c *AppConfig) parseFlags() (*pflag.FlagSet, error) {
	pflag.StringP(configFlag, "c", "", "config file")

	if !pflag.Parsed() {
		pflag.Parse()
	} else {
		return nil, errors.New("ошибка чтения конфигурации: парсинг флагов был выполнен до инициализации конфигурации")
	}

	err := v.BindPFlags(pflag.CommandLine)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения конфигурации: %w", err)
	}

	return pflag.CommandLine, nil
}

func (c *AppConfig) setDefaultPaths() {
	v.AddConfigPath(".")
}

func (c *AppConfig) setConfigSources() error {
	v.AutomaticEnv()

	if f, err := c.parseFlags(); err == nil {
		if cfgFile, _ := f.GetString(configFlag); len(cfgFile) != 0 {
			v.SetConfigFile(cfgFile)
		} else {
			log.Debug().Msg("В флаге не передан путь до конфигурационного файла, использую стандартные пути для конфига")
			c.setDefaultPaths()
		}
	} else {
		log.Warn().Err(err).Msg("Ошибка парсинга флагов, использую стандартные пути для конфига")
		c.setDefaultPaths()
	}

	return nil
}

var v *viper.Viper = viper.NewWithOptions(
	viper.KeyDelimiter("_"),
	viper.EnvKeyReplacer(strings.NewReplacer(".", "_")),
)

// Init валидирует параметры в конфиге, загружает конфигурацию из файла в структуру
func Init() error {
	Config = new(AppConfig)

	Config.setDefaults(v)

	if err := Config.setConfigSources(); err != nil {
		return err
	}

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Warn().Msg("Конфигурационный файл не найден, использую конфигурацию по умолчанию")
		} else {
			return fmt.Errorf("ошибка чтения конфигурации: %w", err)
		}
	}

	err := v.Unmarshal(Config)
	if err != nil {
		return fmt.Errorf("ошибка чтения конфигурации: %w", err)
	}

	return nil
}
