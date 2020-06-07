package ssh

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

var uploadedByHostRegistry map[string]map[string]bool
var once sync.Once
var uploadMutex = &sync.Mutex{}

func GetUploadedByHostRegistry() map[string]map[string]bool {
	once.Do(func() {
		uploadedByHostRegistry = make(map[string]map[string]bool)
	})
	return uploadedByHostRegistry
}

func Upload(cfg *sshclient.Config, localPath string) error {

	uploadMutex.Lock()
	r := GetUploadedByHostRegistry()
	if r[cfg.Host] == nil {
		r[cfg.Host] = make(map[string]bool)
	}
	if r[cfg.Host][localPath] {
		uploadMutex.Unlock()
		return nil
	}
	r[cfg.Host][localPath] = true
	uploadMutex.Unlock()

	retryCount := 0
	var err error
	for {
		err = uploadTry(cfg, localPath)
		if err == nil {
			break
		}
		logrus.Warn(err)
		if retryCount == cfg.MaxRetry {
			break
		}
		retryCount = retryCount + 1
		logrus.Warnf("attempt %v/%v failed", retryCount, cfg.MaxRetry)
		logrus.Debug("retrying... ")
	}

	if err != nil {
		uploadMutex.Lock()
		r[cfg.Host][localPath] = false
		uploadMutex.Unlock()
	}

	return err
}
func uploadTry(cfg *sshclient.Config, localPath string) error {

	clientConfig, err := cfg.ClientConfig()
	if err != nil {
		return err
	}
	client := scp.NewClient(cfg.Host+":"+strconv.Itoa(cfg.Port), &clientConfig)

	remotePath := "/home/" + cfg.User + "/.snip/" + localPath

	logrus.Debugf("connecting to %v via ssh", cfg.Host)
	// err := client.Connect()
	sshClient, err := ssh.Dial("tcp", client.Host, client.ClientConfig)
	if err != nil {
		logrus.Warnf("connection to %v via ssh failed", cfg.Host)
		return err
	}
	logrus.Debugf("connected to %v", cfg.Host)
	client.Conn = sshClient.Conn
	client.Session, err = sshClient.NewSession()

	if err != nil {
		return err
	}

	f, _ := os.Open(localPath)
	defer client.Close()
	defer f.Close()

	logrus.Debugf("uploading script %v", remotePath)

	var mkdirErrB bytes.Buffer
	client.Session.Stderr = &mkdirErrB
	dir := filepath.Dir(remotePath)
	err = client.Session.Run("mkdir -p " + dir)
	errors.Check(err)
	mkdirErr := mkdirErrB.String()
	if mkdirErr != "" {
		logrus.Warn(mkdirErr)
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
