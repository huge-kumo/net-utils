package ssh

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
)

type Client struct {
	Servers []*sshClient
}

func NewClient(path string) (cli *Client, err error) {
	var body []byte
	body, err = ioutil.ReadFile(path)
	if err != nil {
		return
	}

	cli = &Client{}
	if err = json.Unmarshal(body, &cli.Servers); err != nil {
		return
	}
	return
}

func (c *Client) Execute(ctx context.Context, command string, tags ...string) (outputs []string, err error) {
	for _, server := range c.Servers {
		if len(tags) != 0 && !server.checkTagsExist(tags...) {
			continue
		}

		var output string
		if output, err = server.execute(ctx, command); err != nil {
			return
		}
		outputs = append(outputs, output)
	}
	return
}

type sshClient struct {
	Name string   `json:"name,omitempty"`
	Tags []string `json:"tags,omitempty"`

	HostName     string `json:"hostName,omitempty"`     // 服务器地址
	Port         string `json:"port,omitempty"`         // 服务器端口
	UserName     string `json:"userName,omitempty"`     // 服务器用户
	IdentityFile string `json:"identityFile,omitempty"` // 服务器私钥
}

func (c *sshClient) execute(ctx context.Context, cli string) (output string, err error) {
	var args []string
	args = append(args, "-i", c.IdentityFile)
	args = append(args, fmt.Sprintf("%s@%s", c.UserName, c.HostName))
	args = append(args, "-p", c.Port)
	args = append(args, cli)
	cmd := exec.CommandContext(ctx, "ssh", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err = cmd.Run(); err != nil {
		log.Printf("[警告] %s服务器执行命令失败 %s", c.Name, err.Error())
		err = nil
	}

	if stderr.Len() != 0 {
		err = errors.New(stderr.String())
		return
	}
	output = stdout.String()
	return
}

func (c *sshClient) checkTagsExist(tags ...string) bool {
	for _, tag := range tags {
		for _, subTag := range c.Tags {
			if tag == subTag {
				return true
			}
		}
	}
	return false
}
