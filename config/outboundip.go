package config

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

// OutboundIP is a ready-to-use CIDR scoped to this process's own public
// (post-NAT) IPv4 address, as seen from the internet. A container has no
// local way to learn this (its network interfaces only expose private
// addresses), so it's detected once at startup via an external echo service
// and cached here as "<ip>/32". RDB instance firewall rules use it instead
// of 0.0.0.0/0. Falls back to 0.0.0.0/0 if detection fails.
var OutboundIP = "0.0.0.0/0"

const ipEchoURL = "https://checkip.amazonaws.com"

// InitOutboundIP detects and caches OutboundIP. Call once at startup.
// On failure OutboundIP is left at its 0.0.0.0/0 fallback.
func InitOutboundIP() error {
	client := &http.Client{Timeout: 5 * time.Second}

	resp, err := client.Get(ipEchoURL)
	if err != nil {
		return fmt.Errorf("failed to detect outbound public IP: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read outbound IP response: %w", err)
	}

	ip := strings.TrimSpace(string(body))
	if ip == "" {
		return fmt.Errorf("empty response from %s", ipEchoURL)
	}

	OutboundIP = ip + "/32"
	log.Info().Str("outboundCIDR", OutboundIP).Msg("detected outbound public IP")
	return nil
}
