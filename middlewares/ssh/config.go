package ssh

import (
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	"gitlab.com/youtopia.earth/ops/snip/errors"
)

type Config struct {
	Host string
	Port int

	User string

	File string
	Pass string

	MaxRetry int
}

func CreateConfig(vars map[string]string) *Config {
	SSHHost := vars["SSH_HOST"]
	if SSHHost == "" {
		SSHHost = "localhost"
	}

	SSHPortStr := vars["SSH_PORT"]
	var SSHPort int
	if SSHPortStr == "" {
		SSHPort = 22
	} else {
		var err error
		SSHPort, err = strconv.Atoi(SSHPortStr)
		errors.Check(err)
	}

	SSHUser := vars["SSH_USER"]
	if SSHUser == "" {
		currentUser, err := user.Current()
		errors.Check(err)
		SSHUser = currentUser.Username
	}

	SSHPass := vars["SSH_PASS"]

	SSHFile := vars["SSH_FILE"]
	if SSHFile == "" && SSHPass == "" {
		SSHFile = "~/.ssh/id_rsa"
	}
	if strings.HasPrefix(SSHFile, "~/") {
		usr, err := user.Current()
		errors.Check(err)
		dir := usr.HomeDir
		SSHFile = filepath.Join(dir, SSHFile[2:])
	}

	SSHMaxRetryStr := vars["SSH_MAX_RETRY"]
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
		MaxRetry: SSHMaxRetry,
	}

	return cfg
}
