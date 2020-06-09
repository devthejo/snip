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
}

func CreateConfig(vars map[string]string) *Config {
	SSHHost := vars["@SSH_HOST"]
	if SSHHost == "" {
		SSHHost = "localhost"
	}

	SSHPortStr := vars["@SSH_PORT"]
	var SSHPort int
	if SSHPortStr == "" {
		SSHPort = 22
	} else {
		var err error
		SSHPort, err = strconv.Atoi(SSHPortStr)
		errors.Check(err)
	}

	SSHUser := vars["@SSH_USER"]
	if SSHUser == "" {
		currentUser, err := user.Current()
		errors.Check(err)
		SSHUser = currentUser.Username
	}

	SSHPass := vars["@SSH_PASS"]
	SSHSock := vars["@SSH_SOCK"]
	SSHFile := vars["@SSH_FILE"]

	if SSHSock == "" && SSHPass == "" && SSHFile == "" {
		SSHSock = os.Getenv("SSH_AUTH_SOCK")
	}

	if SSHFile == "" && SSHPass == "" && SSHSock == "" {
		SSHFile = "~/.ssh/id_rsa"
	}
	if strings.HasPrefix(SSHFile, "~/") {
		usr, err := user.Current()
		errors.Check(err)
		dir := usr.HomeDir
		SSHFile = filepath.Join(dir, SSHFile[2:])
	}

	SSHMaxRetryStr := vars["@SSH_MAX_RETRY"]
	var SSHMaxRetry int
	if SSHMaxRetryStr == "" {
		SSHMaxRetry = 3
	} else {
		var err error
		SSHMaxRetry, err = strconv.Atoi(SSHMaxRetryStr)
		errors.Check(err)
	}

	cfg := &Config{
		Host:     SSHHost,
		Port:     SSHPort,
		User:     SSHUser,
		File:     SSHFile,
		Pass:     SSHPass,
		Sock:     SSHSock,
		MaxRetry: SSHMaxRetry,
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
