package hooks

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/jellydator/ttlcache/v3"
	"hash/fnv"
	"mqtt-server/config"
	"net/http"
	"time"

	"github.com/mochi-co/mqtt/v2"
	"github.com/mochi-co/mqtt/v2/packets"
)

type HTTPAuthHook struct {
	mqtt.HookBase
	authCache  *ttlcache.Cache[uint32, bool]
	aclCache   *ttlcache.Cache[uint32, bool]
	httpAuth   string
	httpAcl    string
	httpClient *http.Client
}

// ID returns the ID of the hook.
func (h *HTTPAuthHook) ID() string {
	return "http-auth-hook"
}

// Provides indicates which hook methods this hook provides.
func (h *HTTPAuthHook) Provides(b byte) bool {
	return bytes.Contains([]byte{
		mqtt.OnConnectAuthenticate,
		mqtt.OnACLCheck,
	}, []byte{b})
}

func (h *HTTPAuthHook) Configure() {
	cfg := config.GetConfig()
	h.authCache = ttlcache.New[uint32, bool](
		ttlcache.WithTTL[uint32, bool](time.Duration(cfg.HTTPAuthCacheSeconds) * time.Second),
	)

	h.aclCache = ttlcache.New[uint32, bool](
		ttlcache.WithTTL[uint32, bool](time.Duration(cfg.HTTPAclCacheSeconds) * time.Second),
	)

	h.httpAuth = cfg.HTTPAuthURL
	h.httpAcl = cfg.HTTPAclURL

	h.httpClient = &http.Client{
		Timeout: time.Duration(cfg.HTTPClientTimeoutSeconds) * time.Second,
	}

	go h.authCache.Start()
	go h.aclCache.Start()
}

func hash(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}

// OnConnectAuthenticate returns true/allowed for all requests.
func (h *HTTPAuthHook) OnConnectAuthenticate(cl *mqtt.Client, pk packets.Packet) bool {
	if h.httpAuth == "" {
		return true
	}

	pktHash := hash(fmt.Sprintf("%s%s", cl.Properties.Username, pk.Connect.Password))
	item := h.authCache.Get(pktHash)
	if item != nil && !item.IsExpired() {
		return item.Value()
	}

	data := struct {
		ClientID string `json:"client_id"`
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		ClientID: cl.ID,
		Username: string(cl.Properties.Username),
		Password: string(pk.Connect.Password),
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return false
	}

	request, err := http.NewRequest("POST", h.httpAuth, bytes.NewBuffer(jsonData))
	if err != nil {
		return false
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	resp, err := h.httpClient.Do(request)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		h.authCache.Set(pktHash, true, ttlcache.DefaultTTL)
		return true
	} else {
		h.authCache.Set(pktHash, false, ttlcache.DefaultTTL)
	}
	return false
}

// OnACLCheck returns true/allowed for all checks.
func (h *HTTPAuthHook) OnACLCheck(cl *mqtt.Client, topic string, write bool) bool {
	if h.httpAcl == "" {
		return true
	}
	pktHash := hash(fmt.Sprintf("%s%s%t", cl.Properties.Username, topic, write))
	item := h.aclCache.Get(pktHash)
	if item != nil && !item.IsExpired() {
		return item.Value()
	}

	data := struct {
		ClientID string `json:"client_id"`
		Username string `json:"username"`
		Topic    string `json:"topic"`
		Write    bool   `json:"write"`
	}{
		ClientID: cl.ID,
		Username: string(cl.Properties.Username),
		Topic:    topic,
		Write:    write,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return false
	}

	request, err := http.NewRequest("POST", h.httpAcl, bytes.NewBuffer(jsonData))
	if err != nil {
		return false
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	resp, err := h.httpClient.Do(request)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		h.aclCache.Set(pktHash, true, ttlcache.DefaultTTL)
		return true
	} else {
		h.aclCache.Set(pktHash, false, ttlcache.DefaultTTL)
	}
	return false
}
