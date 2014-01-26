package config

import (
    "code.google.com/p/gcfg"
)

type Config struct {
    Server struct {
        Name string
        Port string
    }

    Websocket struct {
        ReadBufSize    int
        WriteBufSize   int
        MaxMessageSize int64
        WriteWait      int
    }
}

func Conf() (*Config, error) {
    var cfg Config
    err := gcfg.ReadFileInto(&cfg, "config/app.ini")

    return &cfg, err
}
