package sshclient

import (
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"gitlab.com/youtopia.earth/ops/snip/errors"
	"golang.org/x/crypto/ssh"
)

type Config struct {
	Host string
	Port int

	User string

	Sock string

	File string
	Pass string

	MaxRetry int
	CacheKey string
}

func CreateConfig(vars map[string]string) *Config {
	SSHHost := vars["host"]
	if SSHHost == "" {
		SSHHost = "localhost"
	}

	SSHPortStr := vars["port"]
	var SSHPort int
	if SSHPortStr == "" {
		SSHPort = 22
	} else {
		var err error
		SSHPort, err = strconv.Atoi(SSHPortStr)
		errors.Check(err)
	}

	SSHUser := vars["user"]
	if SSHUser == "" {
		currentUser, err := user.Current()
		errors.Check(err)
		SSHUser = currentUser.Username
	}

	SSHPass := vars["pass"]
	SSHSock := vars["sock"]
	SSHFile := vars["file"]

	if SSHSock == "" && SSHPass == "" && SSHFile == "" {
		SSHSock = os.Getenv("SSH_AUTH_SOCK")
	}

	if SSHFile == "" && SSHPass == "" && SSHSock == "" {
		SSHFile = "~/.ssh/id_rsa"
	}
	if strings.HasPrefix(SSHFile, "~/") {
		usr, err := user.Current()
		errors.Check(err)
		SSHFile = filepath.Join(usr.HomeDir, SSHFile[2:])
	}

	SSHMaxRetryStr := vars["max_retry"]
	var SSHMaxRetry int
	if SSHMaxRetryStr == "" {
		SSHMaxRetry = 3
	} else {
		var err error
		SSHMaxRetry, err = strconv.Atoi(SSHMaxRetryStr)
		errors.Check(err)
	}

	CacheKey := "host:client:" + SSHHost + ":" + SSHUser

	cfg := &Config{
		Host:     SSHHost,
		Port:     SSHPort,
		User:     SSHUser,
		File:     SSHFile,
		Pass:     SSHPass,
		Sock:     SSHSock,
		MaxRetry: SSHMaxRetry,
		CacheKey: CacheKey,
	}

	return cfg
}

func (cfg *Config) ClientConfig() (ssh.ClientConfig, error) {
	var clientConfig ssh.ClientConfig
	var err error
	if cfg.Sock != "" {
		clientConfig, err = SshAgent(cfg.User, cfg.Sock, ssh.InsecureIgnoreHostKey())
	} else if cfg.File != "" {
		if cfg.Pass != "" {
			clientConfig, err = PrivateKeyWithPassphrase(cfg.User, []byte(cfg.Pass), cfg.File, ssh.InsecureIgnoreHostKey())
		} else {
			clientConfig, err = PrivateKey(cfg.User, cfg.File, ssh.InsecureIgnoreHostKey())
		}
	} else {
		clientConfig, err = PasswordKey(cfg.User, cfg.Pass, ssh.InsecureIgnoreHostKey())
	}
	return clientConfig, err
}
