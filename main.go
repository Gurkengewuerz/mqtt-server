package main

import (
	"fmt"
	"github.com/mochi-co/mqtt/v2"
	"github.com/mochi-co/mqtt/v2/hooks/storage/badger"
	"mqtt-server/config"
	"mqtt-server/hooks"
	"os"
	"os/signal"
	"syscall"

	"github.com/mochi-co/mqtt/v2/hooks/auth"
	"github.com/mochi-co/mqtt/v2/listeners"
	"github.com/rs/zerolog/log"
)

func bool2byte(b bool) byte {
	if b {
		return 1
	}
	return 0
}

func main() {
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)

	if err := config.Init(); err != nil {
		log.Panic().Err(err).Msg("failed to parse init")
	}

	cfg := config.GetConfig()
	if cfg.TCPEnabled == false && cfg.WSEnabled == false {
		log.Panic().Msg("at least one protocol must be enabled")
	}

	options := &mqtt.Options{
		Capabilities: mqtt.DefaultServerCapabilities,
	}

	if cfg.MQTTMaximumQos != -1 {
		options.Capabilities.MaximumQos = byte(cfg.MQTTMaximumQos)
	}

	if cfg.MQTTMinimumProtocolVersion != -1 {
		options.Capabilities.MinimumProtocolVersion = byte(cfg.MQTTMinimumProtocolVersion)
	}

	if cfg.MQTTMaximumMessageExpiryInterval != -1 {
		options.Capabilities.MaximumMessageExpiryInterval = cfg.MQTTMaximumMessageExpiryInterval
	}

	if cfg.MQTTReceiveMaximum != -1 {
		options.Capabilities.ReceiveMaximum = uint16(cfg.MQTTReceiveMaximum)
	}

	if cfg.MQTTServerKeepAlive != -1 {
		options.Capabilities.ServerKeepAlive = uint16(cfg.MQTTServerKeepAlive)
	}

	options.Capabilities.MaximumPacketSize = cfg.MQTTMaximumPacketSize
	options.Capabilities.RetainAvailable = bool2byte(cfg.MQTTRetainAvailable)
	options.Capabilities.WildcardSubAvailable = bool2byte(cfg.MQTTWildcardSubAvailable)
	options.Capabilities.SharedSubAvailable = bool2byte(cfg.MQTTSharedSubAvailable)

	server := mqtt.New(options)

	if cfg.HTTPAuthURL != "" || cfg.HTTPAclURL != "" {
		httpHook := new(hooks.HTTPAuthHook)
		httpHook.Configure()
		err := server.AddHook(httpHook, nil)
		if err != nil {
			server.Log.Panic().Err(err).Msg("failed to add http hook")
		}
	} else {
		if cfg.PathAuth == "" {
			err := server.AddHook(new(auth.AllowHook), nil)
			if err != nil {
				server.Log.Panic().Err(err).Msg("failed to add allow all hook")
			}
		} else {
			// Get ledger from yaml file
			data, err := os.ReadFile(cfg.PathAuth)
			if err != nil {
				server.Log.Panic().Err(err).Msg("failed to read auth file")
			}

			err = server.AddHook(new(auth.Hook), &auth.Options{
				Data: data, // build ledger from byte slice, yaml or json
			})
			if err != nil {
				server.Log.Panic().Err(err).Msg("failed to parse auth file")
			}
		}
	}

	if cfg.PathPersistence != "" {
		err := server.AddHook(new(badger.Hook), &badger.Options{
			Path: cfg.PathPersistence,
		})
		if err != nil {
			server.Log.Panic().Err(err).Msg("failed to add persistence storage")
		}
	}

	if cfg.TCPEnabled {
		tcp := listeners.NewTCP("t1", fmt.Sprintf(":%d", cfg.TCPPort), nil)
		err := server.AddListener(tcp)
		if err != nil {
			server.Log.Panic().Err(err).Msg("failed to initialize tcp")
		}
	}

	if cfg.WSEnabled {
		ws := listeners.NewWebsocket("ws1", fmt.Sprintf(":%d", cfg.WSPort), nil)
		err := server.AddListener(ws)
		if err != nil {
			server.Log.Panic().Err(err).Msg("failed to initialize websocket")
		}
	}

	go func() {
		err := server.Serve()
		if err != nil {
			server.Log.Panic().Err(err).Msg("failed to serve server")
		}
	}()

	server.Log.Info().Int("pid", os.Getpid()).Msg("server started")
	<-sigint
	server.Log.Warn().Msg("caught signal, stopping...")
	server.Close()
	server.Log.Info().Msg("main.go finished")
}
