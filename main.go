package main

import (
	"bytes"
	"context"
	"flag"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

var paths []string
var debug *bool

func init() {
  readConfig()
  log.Println(paths)
}

func readConfig() {
  homeDir, err := os.UserHomeDir()
  if err != nil {
    log.Fatal(err)
  }
  configPath := homeDir + "/.rorschach"
  configData, err := os.ReadFile(configPath)
  if err != nil {
    log.Fatal(err)
  }

  paths = strings.Split(string(configData), "\n")
}

func main() {
  debug = flag.Bool("d", false, "enable debug log level")
  flag.Parse()

  ctx, cancel := context.WithCancel(context.Background())
  sigCh := make(chan os.Signal)
  signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

  var wg sync.WaitGroup

  wg.Add(2)
  go commitWorker(&wg, ctx)
  go pushWorker(&wg, ctx)

  <-sigCh
  log.Println("shutting down")
  cancel()
  wg.Wait()
  os.Exit(0)
}

func commitWorker(wg *sync.WaitGroup, ctx context.Context) {
  defer wg.Done()
  for {
    select {
      case <- ctx.Done():
        log.Println("gracefully shutting down commit worker")
        return;
      case <- time.After(15 * time.Second):
        log.Println("doing regular commit")
        commit()
    }
  }
}

func pushWorker(wg *sync.WaitGroup, ctx context.Context) {
  defer wg.Done()
  for {
    select {
    case <- ctx.Done():
      log.Println("doing last push before shutting down")
      push()
      log.Println("gracefully shutting down push worker")
      return;
    case <- time.After(1 * time.Minute):
      log.Println("doing regular push")
      push()
    }
  }
}

func commit() {
  stCmd := []string{"git", "status"}
  addCmd := []string{"git", "add", "-A"}
  cmCmd := []string{"git", "commit", "-m", "regular rorschach commit"}
  for _, p := range paths {
    if hasNewInfo(execCommand(stCmd, p)) {
      execCommand(addCmd, p)
      execCommand(cmCmd, p)
    }
  }
}

func push() {
  pushCmd := []string{"git", "push"}
  for _, p := range paths {
    execCommand(pushCmd, p)
  }
}

func hasNewInfo(cmdRes string) bool {
  return strings.Contains(cmdRes, "git add") 
}

func execCommand(cmdStr []string, dir string) string {
  cmd := exec.Command(cmdStr[0], cmdStr[1:]...)
  var out bytes.Buffer
  cmd.Stdout = &out
  cmd.Dir = dir
  err := cmd.Run()
  
  if err != nil {
    log.Fatal(err)
  }

  cmdRes := out.String()

  if *debug {
    log.Println(cmdRes)
  }

  return cmdRes
}
