package sshutils

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"

	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"

	"gitlab.com/ytopia/ops/snip/errors"
	"gitlab.com/ytopia/ops/snip/sshclient"
)

func Upload(cfg *sshclient.Config, src string, dest string, logger *logrus.Entry) error {

	retryCount := 0
	var err error
	for {
		err = uploadTry(cfg, src, dest, logger)
		if err == nil {
			break
		}
		logger.Warn(err)
		if retryCount == cfg.MaxRetry {
			break
		}
		retryCount = retryCount + 1
		logger.Warnf("attempt %v/%v failed", retryCount, cfg.MaxRetry)
		logger.Debug("retrying... ")
	}

	return err
}
func uploadTry(cfg *sshclient.Config, src string, dest string, logger *logrus.Entry) error {

	clientConfig, err := cfg.ClientConfig()
	if err != nil {
		return err
	}
	client := scp.NewClient(cfg.Host+":"+strconv.Itoa(cfg.Port), &clientConfig)

	logger.Debugf("connecting to %v via ssh", cfg.Host)
	// err := client.Connect()
	sshClient, err := ssh.Dial("tcp", client.Host, client.ClientConfig)
	if err != nil {
		logger.Warnf("connection to %v via ssh failed", cfg.Host)
		return err
	}
	logger.Debugf("connected to %v", cfg.Host)
	client.Conn = sshClient.Conn
	client.Session, err = sshClient.NewSession()

	if err != nil {
		return err
	}

	f, _ := os.Open(src)
	defer client.Close()
	defer f.Close()

	logger.Debugf("uploading script %v", dest)

	var mkdirErrB bytes.Buffer
	client.Session.Stderr = &mkdirErrB
	dir := filepath.Dir(dest)
	err = client.Session.Run("mkdir -p " + dir)
	errors.Check(err)
	mkdirErr := mkdirErrB.String()
	if mkdirErr != "" {
		logger.Warn(mkdirErr)
	}
	client.Session.Close()

	client.Session, err = sshClient.NewSession()
	if err != nil {
		return err
	}

	err = client.CopyFile(f, dest, "0755")

	if err != nil {
		return err
	}

	return nil
}
