package main

import (
	"fmt"
	"os"

	"github.com/simonswine/kube2consul/pkg/kube2consul"
)

func main() {
	k := kube2consul.New()
	if err := k.RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}
