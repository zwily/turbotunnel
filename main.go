package main

import (
	"flag"
	"github.com/zwily/turbotunnel/server"
	"io/ioutil"
	"launchpad.net/goyaml"
	"log"
	"os"
	"os/signal"
	"syscall"
)

type TunnelDef struct {
	Name       string
	LocalPort  int    `yaml:"localPort"`
	JumpHost   string `yaml:"jumpHost"`
	RemoteHost string `yaml:"remoteHost"`
	RemotePort int    `yaml:"remotePort"`
	EnvCommand string `yaml:"envCommand"`
}

type Config struct {
	Tunnels []TunnelDef
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	var configPath = flag.String("config", "", "Path to config file")
	flag.Parse()

	configYaml, err := ioutil.ReadFile(*configPath)
	if err != nil {
		log.Fatal(err)
	}

	var config Config
	goyaml.Unmarshal(configYaml, &config)

	for _, t := range config.Tunnels {
		s := server.New(t.Name, t.LocalPort, t.JumpHost, t.RemoteHost, t.RemotePort, t.EnvCommand)
		go s.Listen()
		defer s.Close()
	}

	sigchan := make(chan os.Signal)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)
	<-sigchan
}
