package sshutils

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"sync"

	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"

	"gitlab.com/youtopia.earth/ops/snip/errors"
	"gitlab.com/youtopia.earth/ops/snip/sshclient"
)

var uploadedByHostRegistry map[string]map[string]*UploadState
var once sync.Once
var uploadMutex = &sync.Mutex{}

func GetUploadedByHostRegistry() map[string]map[string]*UploadState {
	once.Do(func() {
		uploadedByHostRegistry = make(map[string]map[string]*UploadState)
	})
	return uploadedByHostRegistry
}

type UploadState struct {
	Uploading bool
	Ready     bool
	Wg        sync.WaitGroup
}

func Upload(cfg *sshclient.Config, src string, dest string, logger *logrus.Entry) error {

	uploadMutex.Lock()
	r := GetUploadedByHostRegistry()
	if r[cfg.Host] == nil {
		r[cfg.Host] = make(map[string]*UploadState)
	}
	if r[cfg.Host][dest] == nil {
		r[cfg.Host][dest] = &UploadState{}
	}
	if r[cfg.Host][dest].Ready {
		uploadMutex.Unlock()
		return nil
	}
	if r[cfg.Host][dest].Uploading {
		r[cfg.Host][dest].Wg.Wait()
		if r[cfg.Host][dest].Ready {
			uploadMutex.Unlock()
			return nil
		}
	}
	r[cfg.Host][dest].Wg.Add(1)
	r[cfg.Host][dest].Uploading = true
	uploadMutex.Unlock()

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

	if err == nil {
		r[cfg.Host][dest].Ready = true
	} else {
		r[cfg.Host][dest].Ready = false
	}
	r[cfg.Host][dest].Uploading = false
	r[cfg.Host][dest].Wg.Done()

	return err
}
func uploadTry(cfg *sshclient.Config, src string, dest string, logger *logrus.Entry) error {

	clientConfig, err := cfg.ClientConfig()
	if err != nil {
		return err
	}
	client := scp.NewClient(cfg.Host+":"+strconv.Itoa(cfg.Port), &clientConfig)

	remotePath := GetRemotePath(cfg.User, dest)

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

	logger.Debugf("uploading script %v", remotePath)

	var mkdirErrB bytes.Buffer
	client.Session.Stderr = &mkdirErrB
	dir := filepath.Dir(remotePath)
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

	err = client.CopyFile(f, remotePath, "0755")

	if err != nil {
		return err
	}

	return nil
}
