package buoy

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"

	"github.com/segmentio/analytics-go"
)

// segment client code
const clientCode = "jLwurYoMosZliChljnSNq7mCAOOd8Vnn"

// random hash code for hashing the account IDs
const hashCode = "ZQM7q96ar8g1y7Id"

// compute the hmac255 value for the message
func computeHmac256(message string, secret string) string {
	key := []byte(secret)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(message))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// Hash the accountId so we don't know what it is.
func hashAccountID(accountID string) string {
	return computeHmac256(accountID, hashCode)
}

// get the analytics client
func analyticsClient() *analytics.Client {
	client := analytics.New(clientCode)
	client.Size = 1 // We only send one message at a time, no need to cache
	return client
}

// BuoyEvent used to track different events
type BuoyEvent struct {
	AccountID      string
	SwarmID        string
	NodeID         string
	Region         string
	ServiceCount   int
	ManagerCount   int
	WorkerCount    int
	DockerVersion  string
	Edition        string
	EditionOS      string
	EditionVersion string
	EditionAddon   string
	Channel        string
	IaasProvider   string
}

// New returns a new analytics tracker with `config`.
func (buoyEvent *BuoyEvent) Properties() map[string]interface{} {
	return map[string]interface{}{
		"swarm_id":        buoyEvent.SwarmID,
		"node_id":         buoyEvent.NodeID,
		"region":          buoyEvent.Region,
		"service_count":   buoyEvent.ServiceCount,
		"manager_count":   buoyEvent.ManagerCount,
		"worker_count":    buoyEvent.WorkerCount,
		"docker_version":  buoyEvent.DockerVersion,
		"edition":         buoyEvent.Edition,
		"edition_os":      buoyEvent.EditionOS,
		"edition_version": buoyEvent.EditionVersion,
		"edition_addon":   buoyEvent.EditionAddon,
		"channel":         buoyEvent.Channel,
		"iaas_provider":   buoyEvent.IaasProvider,
	}
}

// Identify the user in segment
func (buoyEvent *BuoyEvent) Identify() {
	client := analyticsClient()
	client.Identify(&analytics.Identify{
		UserId: hashAccountID(buoyEvent.AccountID),
		Traits: buoyEvent.Properties(),
	})
	client.Close()
}

// Track: Send track event to segment
func (buoyEvent *BuoyEvent) Track(event string) {
	client := analyticsClient()
	client.Track(&analytics.Track{
		Event:      event,
		UserId:     hashAccountID(buoyEvent.AccountID),
		Properties: buoyEvent.Properties(),
	})
	client.Close()
}
