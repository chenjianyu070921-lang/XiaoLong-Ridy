package main

import (
	"fmt"
	"os"

	"github.com/zeromicro/go-zero/zrpc"
	"gopkg.in/yaml.v3"
)

type Cfg struct {
	HTTPAddr    string             `yaml:"httpAddr"`
	ChatRpc     zrpc.RpcClientConf `yaml:"chatRpc"`
	SigningKeys []string           `yaml:"signingKeys"`
}

func main() {
	data, err := os.ReadFile("api/chat/etc/chat.yaml")
	if err != nil {
		fmt.Println("READ_ERR", err)
		return
	}
	var cfg Cfg
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		fmt.Println("UNMARSHAL_ERR", err)
		return
	}
	fmt.Printf("HTTPAddr   = %q\n", cfg.HTTPAddr)
	fmt.Printf("SigningKeys= %#v\n", cfg.SigningKeys)
	fmt.Printf("ChatRpc.Etcd.Hosts = %#v\n", cfg.ChatRpc.Etcd.Hosts)
	fmt.Printf("ChatRpc.Etcd.Key   = %q\n", cfg.ChatRpc.Etcd.Key)
}
