/*
 * Copyright (c) 2022.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0.
 *
 * If a copy of the MPL was not distributed with
 * this file, You can obtain one at
 *
 *   http://mozilla.org/MPL/2.0/
 */

package telemetry

import (
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"net"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/ezodude/machineid"
	"github.com/go-logr/logr"
	"github.com/posthog/posthog-go"
	core "k8s.io/api/core/v1"
)

// PostHogToken is PostHog API token.
const PostHogToken = "_uzX-_WJoVmE_tsLvu0OFD2tpd0HGz72D5sU1zM2hbs"

// event describes a telemetry event.
type event struct {
	Name       string
	Properties map[string]interface{}
}

// protectedDistinctId returns a hashed ID in a cryptographically secure way.
func protectedDistinctId(cfg Config) (string, error) {
	return machineid.ProtectedID(cfg.AppName)
}

// hashEncode returns a hashed value for a provided string.
func hashEncode(v string) string {
	h := sha1.New()
	h.Write([]byte(v))
	hashed := h.Sum(nil)
	return base64.StdEncoding.EncodeToString(hashed)
}

// getIPAddress gets the ip address using the preferred outbound ip of this machine
func getIPAddress() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return ""
	}
	defer func(conn net.Conn) {
		_ = conn.Close()
	}(conn)

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return string(localAddr.IP)
}

// noopClient is no-op telemetry client used when telemetry is disabled.
type noopClient struct{}

func (n *noopClient) Close() error                                              { return nil }
func (n *noopClient) Enqueue(_ posthog.Message) error                           { return nil }
func (n *noopClient) IsFeatureEnabled(_ string, _ string, _ bool) (bool, error) { return true, nil }
func (n *noopClient) ReloadFeatureFlags() error                                 { return nil }
func (n *noopClient) GetFeatureFlags() ([]posthog.FeatureFlag, error)           { return nil, nil }

// NewClient returns a telemetry client based on the telemetry configuration.
// This could either be a no-op client or a PostHog client.
func NewClient(tCfg Config) (posthog.Client, error) {
	if tCfg.Disable {
		return &noopClient{}, nil
	}
	return posthog.NewWithConfig(PostHogToken, posthog.Config{})
}

// enqueue enqueues a telemetry event.
func enqueue(client posthog.Client, config Config, event event, logger logr.Logger) error {
	properties := buildProperties(event.Properties, config)
	if config.Debug {
		debugProperties(properties, logger)
		return nil
	}

	distinctId, err := protectedDistinctId(config)
	if err != nil {
		return err
	}

	if err := client.Enqueue(posthog.Capture{
		DistinctId: distinctId,
		Event:      event.Name,
		Properties: properties,
	}); err != nil {
		return err
	}

	return nil
}

// debugProperties logs telemetry events using a logger.
func debugProperties(props map[string]interface{}, logger logr.Logger) {
	for k, v := range props {
		logger.Info("ARTILLERY_TELEMETRY_DEBUG=true", k, v)
	}
}

// buildProperties returns default telemetry properties along with any extra provided properties.
func buildProperties(extra map[string]interface{}, cfg Config) map[string]interface{} {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = ""
	}

	properties := map[string]interface{}{
		"source":       cfg.AppName,
		"version":      cfg.Version,
		"containerOS":  strings.ToLower(runtime.GOOS),
		"workerImage":  cfg.WorkerImage,
		"ipHash":       hashEncode(getIPAddress()),
		"hostnameHash": hashEncode(hostname),
		"$ip":          nil,
	}

	for key, val := range extra {
		properties[key] = val
	}

	return properties
}

// Config defines telemetry configuration.
type Config struct {
	Disable     bool
	Debug       bool
	AppName     string
	Version     string
	WorkerImage string
}

// NewConfig return a new telemetry config, that include environment settings.
func NewConfig(appName, version, workerImage string, logger logr.Logger) Config {
	result := Config{
		AppName:     appName,
		Version:     version,
		WorkerImage: workerImage,
	}

	if getDisableConfig(logger) {
		result.Disable = true
	}

	if getDebugConfig(logger) {
		result.Debug = true
	}

	return result
}

// ToK8sEnvVar converts telemetry config to K8s env var settings.
func (t Config) ToK8sEnvVar() []core.EnvVar {
	return []core.EnvVar{
		{
			Name:  "ARTILLERY_DISABLE_TELEMETRY",
			Value: strconv.FormatBool(t.Disable),
		},
		{
			Name:  "ARTILLERY_TELEMETRY_DEBUG",
			Value: strconv.FormatBool(t.Debug),
		},
		// This a serialised JSON object that will be propagated
		// on every worker event
		{
			Name:  "ARTILLERY_TELEMETRY_DEFAULTS",
			Value: fmt.Sprintf(`{"testRunner": "%s"}`, t.AppName),
		},
	}
}

func getDisableConfig(logger logr.Logger) bool {
	disable, ok := os.LookupEnv("ARTILLERY_DISABLE_TELEMETRY")
	if !ok {
		if logger != nil {
			logger.Info("ARTILLERY_DISABLE_TELEMETRY was not set!")
		}
		return false
	}

	parsedDisable, err := strconv.ParseBool(disable)
	if err != nil {
		if logger != nil {
			logger.Info("ARTILLERY_DISABLE_TELEMETRY was not set with boolean type value. TELEMETRY REMAINS ENABLED")
		}
		return false
	}

	return parsedDisable
}

func getDebugConfig(logger logr.Logger) bool {
	debug, ok := os.LookupEnv("ARTILLERY_TELEMETRY_DEBUG")
	if !ok {
		if logger != nil {
			logger.Info("ARTILLERY_TELEMETRY_DEBUG was not set!")
		}
		return false
	}

	parsedDebug, err := strconv.ParseBool(debug)
	if err != nil {
		if logger != nil {
			logger.Info("ARTILLERY_TELEMETRY_DEBUG was not set with boolean type value. TELEMETRY DEBUG REMAINS DISABLED")
		}
		return false
	}

	return parsedDebug
}
