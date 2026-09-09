package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"XiaoLong-Ridy/agent/react"
)

func main() {
	question := "帮我查商品1001多少钱"
	if len(os.Args) > 1 {
		question = os.Args[1]
	}
	tools, err := react.NewDefaultToolRegistry()
	if err != nil {
		panic(err)
	}
	agent, err := react.NewAgent(context.Background(), react.ScriptedModel{}, tools, 3)
	if err != nil {
		panic(err)
	}
	state, err := agent.Run(context.Background(), question)
	if err != nil {
		panic(err)
	}
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		panic(err)
	}
	fmt.Println(string(data))
}
