package sshclient

import (
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
)

type Client struct {
	Host         string
	ClientConfig *ssh.ClientConfig
	Conn         ssh.Conn
	Client       *ssh.Client
	Timeout      time.Duration
}

func CreateClient(cfg *Config) (*Client, error) {
	var err error
	var clientConfig ssh.ClientConfig
	clientConfig, err = cfg.ClientConfig()
	c := &Client{}
	if err == nil {
		c.Host = cfg.Host + ":" + strconv.Itoa(cfg.Port)
		c.ClientConfig = &clientConfig
	}
	return c, err
}

func (c *Client) Connect() error {
	client, err := ssh.Dial("tcp", c.Host, c.ClientConfig)
	if err != nil {
		return err
	}

	c.Client = client
	c.Conn = client.Conn
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) NewSession() (*ssh.Session, error) {
	return c.Client.NewSession()
}

func (c *Client) Close() {
	c.Conn.Close()
}
