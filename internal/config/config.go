package config

import (
    "encoding/json"
    "os"
)

type Config struct {
    Database DatabaseConfig `json:"database"`
    Asterisk AsteriskConfig `json:"asterisk"`
    Router   RouterConfig   `json:"router"`
    Trunks   TrunksConfig   `json:"trunks"`
}

type DatabaseConfig struct {
    Host     string `json:"host"`
    Port     int    `json:"port"`
    Username string `json:"username"`
    Password string `json:"password"`
    Name     string `json:"name"`
}

type AsteriskConfig struct {
    AMI AMIConfig `json:"ami"`
    ARI ARIConfig `json:"ari"`
}

type AMIConfig struct {
    Host     string `json:"host"`
    Port     int    `json:"port"`
    Username string `json:"username"`
    Password string `json:"password"`
}

type ARIConfig struct {
    URL      string `json:"url"`
    Username string `json:"username"`
    Password string `json:"password"`
}

type RouterConfig struct {
    DIDSelectionMode string `json:"did_selection_mode"` // "random" or "sequential"
    MaxRetries       int    `json:"max_retries"`
    RetryDelay       int    `json:"retry_delay"` // seconds
    CallTimeout      int    `json:"call_timeout"` // seconds
}

type TrunksConfig struct {
    ToS1 string `json:"to_s1"`
    ToS3 string `json:"to_s3"`
    ToS4 string `json:"to_s4"`
}

func LoadConfig(path string) (*Config, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    var config Config
    decoder := json.NewDecoder(file)
    if err := decoder.Decode(&config); err != nil {
        return nil, err
    }

    // Set defaults
    if config.Router.DIDSelectionMode == "" {
        config.Router.DIDSelectionMode = "random"
    }
    if config.Router.MaxRetries == 0 {
        config.Router.MaxRetries = 3
    }
    if config.Router.RetryDelay == 0 {
        config.Router.RetryDelay = 5
    }
    if config.Router.CallTimeout == 0 {
        config.Router.CallTimeout = 300
    }

    return &config, nil
}
