package cmd

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"

	scp "github.com/bramvdbogaerde/go-scp"
	"github.com/bramvdbogaerde/go-scp/auth"
	shellquote "github.com/kballard/go-shellquote"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"

	"gitlab.com/youtopia.earth/ops/snip/errors"
)

func CopyToRemoteViaSSH(app App, localPath string) error {
	cfg := app.GetConfig()

	retryCount := 0
	for {
		err := CopyToRemoteViaSSHTry(app, localPath)
		if err == nil {
			break
		}
		logrus.Warn(err)
		if retryCount == cfg.SSHRetryMax {
			return err
		}
		logrus.Info("retrying...")
		retryCount = retryCount + 1
	}
	return nil
}
func CopyToRemoteViaSSHTry(app App, localPath string) error {

	cfg := app.GetConfig()

	sshUser := cfg.SSHUser
	if sshUser == "" {
		currentUser, err := user.Current()
		errors.Check(err)
		sshUser = currentUser.Username
	}
	var clientConfig ssh.ClientConfig
	if cfg.SSHFile != "" {
		SSHFile := cfg.SSHFile
		if strings.HasPrefix(SSHFile, "~/") {
			usr, err := user.Current()
			errors.Check(err)
			dir := usr.HomeDir
			SSHFile = filepath.Join(dir, SSHFile[2:])
		}
		if cfg.SSHPass != "" {
			var err error
			clientConfig, err = auth.PrivateKeyWithPassphrase(sshUser, []byte(cfg.SSHPass), SSHFile, ssh.InsecureIgnoreHostKey())
			errors.Check(err)
		} else {
			var err error
			clientConfig, err = auth.PrivateKey(sshUser, SSHFile, ssh.InsecureIgnoreHostKey())
			errors.Check(err)
		}
	} else {
		var err error
		clientConfig, err = auth.PasswordKey(sshUser, cfg.SSHPass, ssh.InsecureIgnoreHostKey())
		errors.Check(err)
	}
	var sshPort string
	if cfg.SSHPort != 0 {
		sshPort = strconv.Itoa(cfg.SSHPort)
	} else {
		sshPort = "22"
	}
	client := scp.NewClient(cfg.SSHHost+":"+sshPort, &clientConfig)

	remotePath := "/home/" + sshUser + "/.snip/" + localPath

	// err := client.Connect()
	sshClient, err := ssh.Dial("tcp", client.Host, client.ClientConfig)
	if err != nil {
		return err
	}
	client.Conn = sshClient.Conn
	client.Session, err = sshClient.NewSession()

	if err != nil {
		return err
	}

	f, _ := os.Open(localPath)
	defer client.Close()
	defer f.Close()

	var mkdirErrB bytes.Buffer
	client.Session.Stderr = &mkdirErrB
	dir := filepath.Dir(remotePath)
	err = client.Session.Run("mkdir -p " + dir)
	errors.Check(err)
	mkdirErr := mkdirErrB.String()
	if mkdirErr != "" {
		logrus.Warn(mkdirErr)
	}

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

func CmdMiddlewareSSH(app App) *cobra.Command {

	var cmd = &cobra.Command{
		Use:   "ssh",
		Short: "ssh middleware",
		Args:  cobra.ExactArgs(0),
		PreRun: func(cmd *cobra.Command, args []string) {
			app.OnPreRun(cmd)
		},
		Run: func(cmd *cobra.Command, args []string) {
			cfg := app.GetConfig()

			info, err := os.Stdin.Stat()
			errors.Check(err)

			if info.Mode()&os.ModeNamedPipe == 0 {
				fmt.Println("The command is intended to work with pipes.")
				fmt.Println("Usage: echo my-command-to-wrap | snip middleware ssh")
				return
			}

			data, err := ioutil.ReadAll(os.Stdin)
			errors.Check(err)

			commandPath := string(data)
			commandParts, err := shellquote.Split(commandPath)
			errors.Check(err)
			commandBin := commandParts[0]

			if strings.Contains(commandBin, "/") {
				err := CopyToRemoteViaSSH(app, commandBin)
				errors.Check(err)
			}

			var sshPort string
			if cfg.SSHPort != 0 {
				sshPort = strconv.Itoa(cfg.SSHPort)
			} else {
				sshPort = "22"
			}

			sshUser := cfg.SSHUser
			if sshUser == "" {
				currentUser, err := user.Current()
				errors.Check(err)
				sshUser = currentUser.Username
			}

			var s []string
			s = append(s, "ssh")
			s = append(s, "-f")
			s = append(s, "-p "+sshPort)
			if cfg.SSHFile != "" {
				s = append(s, "-i "+cfg.SSHFile)
			}
			s = append(s, "-o StrictHostKeyChecking=no")
			s = append(s, sshUser+"@"+cfg.SSHHost)
			s = append(s, "~/.snip/%s")

			fmt.Printf(strings.Join(s, " "), string(data))

		},
	}

	return cmd
}
