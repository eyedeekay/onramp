//go:build gen
// +build gen

package main

import (
	"log"
	"os"
	"os/exec"
)

func main() {
	cmd := exec.Command("goreadme", "-badge-godoc", "-badge-goreportcard", "-title", "Onramp I2P and Tor Library", "-factories", "-methods", "-functions", "-types", "-variabless")
	file, err := os.Create("DOCS.md")
	if err != nil {
		log.Fatal(err)
	}
	cmd.Stdout = file
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		log.Fatal(err)
	}
	cmdEdgar := exec.Command("edgar")
	cmdEdgar.Stdout = os.Stdout
	cmdEdgar.Stderr = os.Stderr
	err = cmdEdgar.Run()
	if err != nil {
		log.Fatal(err)
	}
}
