package main

import (
    "io"
    "log"
    "net"
    "os"

    redis "github.com/go-redis/redis/v7"
    proxyproto "github.com/pires/go-proxyproto"
)

var localServerHost = ":2221"
var redisClient *redis.Client

func main() {
    // Setup logging
    logfile, err := os.OpenFile("selfor.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0600)
    if err != nil {
        log.Fatalf("Error opening file: %v", err)
    }
    defer logfile.Close()
    log.SetOutput(logfile)

    // Listen for connections
    ln, err := net.Listen("tcp", localServerHost)
    if err != nil {
        log.Fatal(err)
    }
    log.Println("Port forwarding server up and listening on", localServerHost)

    // Wrap listener in a proxyproto listener
    proxyListener := &proxyproto.Listener{Listener: ln}
    defer proxyListener.Close()

    // Connect to redis
    redisClient = redis.NewClient(&redis.Options{
        Addr:     "localhost:6379",
        Password: "",
        DB:       0,
    })

    // Handle each connection in a goroutine
    for {
        conn, err := proxyListener.Accept()
        if err != nil {
            log.Fatal(err)
        }

        go handleConnection(conn)
    }
}

// Copy TCP streams
func forward(src, dest net.Conn) {
    defer src.Close()
    defer dest.Close()
    io.Copy(src, dest)
}

func handleConnection(c net.Conn) {
    // Get remote address
    remoteAddr := ""

    switch addr := c.RemoteAddr().(type) {
    case *net.TCPAddr:
        remoteAddr = addr.IP.String()
    }

    // Log new connection
    log.Println("Connection from:", remoteAddr)

    // Get destination address
    data, err := redisClient.HGetAll(remoteAddr).Result()
    if err != nil {
        log.Println("No mapping found for IP:", remoteAddr)
        c.Close();
        return
    }

    // Get remote host
    remoteHost := data["d"]

    // Dial the remote server
    remote, err := net.Dial("tcp", remoteHost)
    if err != nil {
        log.Println(err)
        c.Close();
        return
    }

    // Connection was successful
    log.Println("Connected", c.RemoteAddr(), "to", remoteHost, "info:", data["i"])

    // Initiate bidrectional communication
    go forward(c, remote)
    go forward(remote, c)
}

