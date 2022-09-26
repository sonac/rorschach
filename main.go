package main

import (
	"bytes"
	"log"
	"os/exec"
)

var paths []string

func init() {
  paths = []string{"/home/sonjaro/git/notes"}
}

func main() {
  stCmd := []string{"git", "status"}
  for _, p := range paths {
    execCommand(stCmd, p)
  }
}

func execCommand(cmdStr []string, dir string) {
  cmd := exec.Command(cmdStr[0], cmdStr[1])
  var out bytes.Buffer
  cmd.Stdout = &out
  cmd.Dir = dir
  err := cmd.Run()
  
  if err != nil {
    log.Fatal(err)
  }

  log.Println(out.String())
  
}
