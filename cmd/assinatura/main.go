package main

import (
	"fmt"
	"os"

	"github.com/jcal1998/ufg-soft-impl-runner/cmd/assinatura/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
		os.Exit(1)
	}
}
