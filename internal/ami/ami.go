package ami

import (
    "fmt"
    "log"
    "sync"
    "time"

    "github.com/CyCoreSystems/ari/v5"
    "github.com/CyCoreSystems/ari/v5/client/native"
    
    "github.com/asterisk-call-routing/internal/config"
)

var (
    client *ari.Client
    mu     sync.RWMutex
)

// Initialize connects to Asterisk AMI/ARI
func Initialize(cfg *config.Config) error {
    // Connect to ARI
    ariClient, err := native.Connect(&native.Options{
        Application:  "call-routing",
        Username:     cfg.Asterisk.ARI.Username,
        Password:     cfg.Asterisk.ARI.Password,
        URL:          cfg.Asterisk.ARI.URL,
        WebsocketURL: fmt.Sprintf("ws://%s/ari/events", cfg.Asterisk.ARI.URL[7:]),
    })
    
    if err != nil {
        return fmt.Errorf("failed to connect to ARI: %v", err)
    }
    
    mu.Lock()
    client = ariClient
    mu.Unlock()
    
    log.Println("Connected to Asterisk ARI")
    return nil
}

// Close disconnects from Asterisk
func Close() {
    mu.Lock()
    defer mu.Unlock()
    
    if client != nil {
        client.Close()
    }
}

// OriginateCall creates a new outbound call
func OriginateCall(endpoint, callerId, extension, context string) (string, error) {
    mu.RLock()
    c := client
    mu.RUnlock()
    
    if c == nil {
        return "", fmt.Errorf("ARI client not initialized")
    }
    
    channelID := fmt.Sprintf("call_%d", time.Now().UnixNano())
    
    channel, err := c.Channel().Originate(nil, ari.OriginateRequest{
        Endpoint:  endpoint,
        Extension: extension,
        Context:   context,
        Priority:  1,
        CallerID:  callerId,
        Timeout:   30,
        Variables: map[string]string{
            "CHANNEL_ID": channelID,
        },
        ChannelID: channelID,
    })
    
    if err != nil {
        return "", fmt.Errorf("failed to originate call: %v", err)
    }
    
    return channel.ID(), nil
}

// GetChannelVariable retrieves a channel variable
func GetChannelVariable(channelID, variable string) (string, error) {
    mu.RLock()
    c := client
    mu.RUnlock()
    
    if c == nil {
        return "", fmt.Errorf("ARI client not initialized")
    }
    
    value, err := c.Channel().GetVariable(ari.NewKey(ari.ChannelKey, channelID), variable)
    if err != nil {
        return "", err
    }
    
    return value, nil
}

// SetChannelVariable sets a channel variable
func SetChannelVariable(channelID, variable, value string) error {
    mu.RLock()
    c := client
    mu.RUnlock()
    
    if c == nil {
        return fmt.Errorf("ARI client not initialized")
    }
    
    return c.Channel().SetVariable(ari.NewKey(ari.ChannelKey, channelID), variable, value)
}

// HangupChannel terminates a channel
func HangupChannel(channelID string, reason string) error {
    mu.RLock()
    c := client
    mu.RUnlock()
    
    if c == nil {
        return fmt.Errorf("ARI client not initialized")
    }
    
    return c.Channel().Hangup(ari.NewKey(ari.ChannelKey, channelID), reason)
}
