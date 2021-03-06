package main

import (
	"fmt"
	"net"
	"strings"
	"io/ioutil"
        "flag"
        "golang.org/x/crypto/ssh"
        "log"
        "os"
)

const version = 0.1

var (
    print_version = flag.Bool("Version", false, "Print the version and exit")
    port = flag.Int("port", 2222, "Port to listen on")
    hostKey = flag.String("hostkey", "host.key", "SSH private host key")
    logger_stdout = log.New(os.Stdout, log.Prefix(), log.Flags())
)

func setupSSHConfig() *ssh.ServerConfig {
    config := &ssh.ServerConfig{
        // Clients are not allowed to connect without authenticating.
        NoClientAuth: false,

        PasswordCallback: func(c ssh.ConnMetadata, pass []byte) (*ssh.Permissions, error) {
            remoteAddr := c.RemoteAddr().String()
            ip := remoteAddr[0:strings.Index(remoteAddr, ":")]
            logger_stdout.Printf("SSH connection from ip=[%s], username=[%s], password=[%s], version=[%s]", ip, c.User(), pass, c.ClientVersion())
            return nil, fmt.Errorf("invalid credentials")
        },

        // Do not allow login using public key authentication.
        PublicKeyCallback: nil,
    }

    privateBytes, err := ioutil.ReadFile(*hostKey)
    if err != nil { log.Fatalf("Failed to load private key %v.  Run ssh-keygen -f %v", *hostKey, *hostKey) }

    private, err := ssh.ParsePrivateKey(privateBytes)
    if err != nil { log.Fatal("Failed to parse private key") }
    config.AddHostKey(private)
    return config
}

func main() {
    flag.Parse()
    config := setupSSHConfig()
    portComplete := fmt.Sprintf(":%v", *port)
    listener, err := net.Listen("tcp", portComplete)
    if err != nil { log.Fatalf("failed to listen on *:%v", *port) }
    log.Printf("listening on %v", *port)
    processConnections(config, listener)

}

func processConnections(sshConfig *ssh.ServerConfig, listener net.Listener) {
    for {
        tcpConn, err := listener.Accept()
            if err != nil {
                log.Println("failed to accept incoming connection (%v)", err)
                continue
            }
            go handleConnection(sshConfig, tcpConn)
    }
}

func handleConnection(sshConfig *ssh.ServerConfig, tcpConn net.Conn) {
	defer tcpConn.Close()

	sshConn, _, _, err := ssh.NewServerConn(tcpConn, sshConfig)
	if err != nil {
		log.Printf("failed to handshake (%v)\n", err)
	} else {
		sshConn.Close()
	}
}
