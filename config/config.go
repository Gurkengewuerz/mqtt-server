package config

import (
	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
)

var cfg Config

type Config struct {
	PathAuth        string `env:"PATH_AUTH_FILE"`
	PathPersistence string `env:"PATH_PERSISTENCE_FILE"`

	TCPEnabled bool `env:"ENABLE_TCP" envDefault:"true"`
	TCPPort    int  `env:"PORT_TCP" envDefault:"1883"`

	WSEnabled bool `env:"ENABLE_WS" envDefault:"true"`
	WSPort    int  `env:"PORT_WS" envDefault:"1882"`

	HTTPAuthURL              string `env:"HTTP_URL_AUTH"`
	HTTPAclURL               string `env:"HTTP_URL_ACL"`
	HTTPAclCacheSeconds      int    `env:"HTTP_AUTH_CACHE" envDefault:"30"`
	HTTPAuthCacheSeconds     int    `env:"HTTP_ACL_CACHE" envDefault:"600"`
	HTTPClientTimeoutSeconds int    `env:"HTTP_CLIENT_TIMEOUT" envDefault:"5"`

	MQTTMaximumQos                   int8   `env:"MQTT_MAX_QOS" envDefault:"-1"`
	MQTTMinimumProtocolVersion       int8   `env:"MQTT_MIN_VERSION" envDefault:"-1"`
	MQTTMaximumPacketSize            uint32 `env:"MQTT_MAX_PACKET_SIZE" envDefault:"0"`
	MQTTMaximumMessageExpiryInterval int64  `env:"MQTT_MAX_MESSAGE_EXPIRY" envDefault:"-1"`
	MQTTReceiveMaximum               int32  `env:"MQTT_MAX_RECEIVE" envDefault:"-1"`
	MQTTServerKeepAlive              int32  `env:"MQTT_KEEP_ALIVE" envDefault:"-1"`
	MQTTRetainAvailable              bool   `env:"MQTT_ENABLE_RETAIN" envDefault:"true"`
	MQTTWildcardSubAvailable         bool   `env:"MQTT_ENABLE_WILDCARD" envDefault:"true"`
	MQTTSharedSubAvailable           bool   `env:"MQTT_ENABLE_SHARED" envDefault:"true"`
}

func Init() error {
	err := godotenv.Load()
	if err != nil {
		log.Warn().Msg("no .env file found - skipping")
	}

	cfg = Config{}
	if err := env.Parse(&cfg); err != nil {
		return err
	}

	return nil
}

func GetConfig() *Config {
	return &cfg
}
