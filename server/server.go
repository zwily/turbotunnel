package server

import (
	"fmt"
	"github.com/tuxychandru/pubsub"
	"log"
	"net"
	"os/exec"
	"time"
)

type Server struct {
	name       string
	localPort  int
	jumpHost   string
	remoteHost string
	remotePort int
	proxyPort  int

	pubsub *pubsub.PubSub
	cmd    *exec.Cmd
}

func New(name string, localPort int, jumpHost string, remoteHost string, remotePort int) *Server {
	s := &Server{
		name:       name,
		localPort:  localPort,
		jumpHost:   jumpHost,
		remoteHost: remoteHost,
		remotePort: remotePort,
	}

	s.pubsub = pubsub.New(0)

	return s
}

func (s *Server) copyConn(from *net.TCPConn, to *net.TCPConn, complete chan bool) {
	var err error
	var bytes []byte = make([]byte, 1024)
	var read int = 0
	for {
		read, err = from.Read(bytes)
		if err != nil {
			complete <- true
			break
		}

		_, err = to.Write(bytes[:read])
		if err != nil {
			complete <- true
			break
		}
	}
}

func (s *Server) proxyConn(lconn *net.TCPConn, rconn *net.TCPConn) {
	// now proxy the connection
	complete := make(chan bool)
	go s.copyConn(lconn, rconn, complete)
	go s.copyConn(rconn, lconn, complete)
	<-complete
	rconn.Close()
	lconn.Close()
}

func (s *Server) connectConn(lconn *net.TCPConn) {
	var rconn *net.TCPConn

	for {
		addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("localhost:%d", s.proxyPort))
		if err != nil {
			log.Fatal(err)
		}

		rconn, err = net.DialTCP("tcp", nil, addr)
		if err == nil {
			break
		}

		// couldn't connect... wait for notice that we're connected
		connectChan := s.pubsub.SubOnce("process-running")
		log.Printf("%s: waiting for process...", s.name)
		<-connectChan
	}

	log.Printf("%s: connected to local port, proxying", s.name)
	s.proxyConn(lconn, rconn)
}

func (s *Server) heartbeat() {
	c := time.Tick(1 * time.Second)
	for _ = range c {
		if s.cmd != nil {
			s.pubsub.Pub(true, "process-running")
		}
	}
}

func (s *Server) handlePending(in <-chan *net.TCPConn) {
	sshDone := make(chan bool)

	go s.heartbeat()

	for {
		select {
		case <-sshDone:
			log.Printf("%s: ssh closed", s.name)
			s.cmd = nil

		case lconn := <-in:
			log.Printf("%s: new connection received", s.name)
			if s.cmd == nil {
				// find an unused local port for the proxying
				l, _ := net.Listen("tcp", "")
				s.proxyPort = l.Addr().(*net.TCPAddr).Port
				log.Printf("%s: using local proxy port %d", s.name, s.proxyPort)
				l.Close()

				notifyCmd := exec.Command("/usr/local/bin/terminal-notifier",
					"-title", s.name,
					"-message", "Tunnel connecting",
				)
				go notifyCmd.Run()

				cmd := exec.Command("/usr/bin/ssh",
					"-N",
					"-o", "ExitOnForwardFailure=true",
					"-L",
					fmt.Sprintf("localhost:%d:%s:%d", s.proxyPort, s.remoteHost, s.remotePort),
					s.jumpHost,
				)

				log.Printf("%s: starting ssh\n", s.name)

				err := cmd.Start()
				if err != nil {
					log.Fatal(err)
				}
				s.cmd = cmd

				go func(cmd *exec.Cmd, done chan bool) {
					cmd.Wait()
					done <- true
				}(cmd, sshDone)
			}

			go s.connectConn(lconn)
		}
	}
}

func (s *Server) Close() {
	if s.cmd != nil {
		log.Printf("%s: killing subprocess", s.name)
		s.cmd.Process.Kill()
	}
}

func (s *Server) Listen() error {
	var err error

	addrString := fmt.Sprintf("localhost:%d", s.localPort)
	log.Println(addrString)

	addr, err := net.ResolveTCPAddr("tcp", addrString)
	if err != nil {
		log.Fatal(err)
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	pending := make(chan *net.TCPConn)
	go s.handlePending(pending)

	log.Printf("%s: listening on %d", s.name, s.localPort)

	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			log.Fatal(err)
		}

		pending <- conn
	}
}
