package main

import (
    "io"
    "log"
    "net"
    "os"
    "flag"

    redis "github.com/go-redis/redis/v7"
    proxyproto "github.com/pires/go-proxyproto"
)

var redisClient *redis.Client
var proxy *string

func main() {
    // Define flags
    localServerHost := flag.String("bind", ":2221", "Address to bind on")
    proxy = flag.String("proxy", "", "Trusted reverse proxy to allow PROXY from")
    logpath := flag.String("log", "selfor.log", "Path of log file")
    redisHost := flag.String("redis", "localhost:6379", "Address of redis instance")
    redisPassword := flag.String("redis-pw", "", "Password of redis database")
    redisDB := flag.Int("redis-db", 0, "Database number of redis")

    // Get flags
    flag.Parse()

    // Setup logging
    logfile, err := os.OpenFile(*logpath, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0600)
    if err != nil {
        log.Fatalf("Error opening file: %v", err)
    }
    defer logfile.Close()
    log.SetOutput(logfile)

    // Listen for connections
    ln, err := net.Listen("tcp", *localServerHost)
    if err != nil {
        log.Fatal(err)
    }
    log.Println("Port forwarding server up and listening on", *localServerHost)

    // Setup PROXY support
    if *proxy != "" {
        ln = &proxyproto.Listener{Listener: ln}
    }
    defer ln.Close()

    // Connect to redis
    redisClient = redis.NewClient(&redis.Options{
        Addr:     *redisHost,
        Password: *redisPassword,
        DB:       *redisDB,
    })

    // Handle each connection in a goroutine
    for {
        conn, err := ln.Accept()
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

    // Check if PROXY is valid
    if *proxy != "" {
        // Check local address
        localAddr := ""
        switch addr := c.LocalAddr().(type) {
        case *net.TCPAddr:
            localAddr = addr.IP.String()
        }

        // Check if local address matches proxy
        if localAddr != *proxy {
            remoteAddr = localAddr
            log.Println("WARN: Direct connection in PROXY mode from:", localAddr)
        }
    }

    // Log new connection
    log.Println("Connection from:", remoteAddr)

    // Get destination address
    data, err := redisClient.HGetAll(remoteAddr).Result()
    if err != nil {
        log.Println("Failed to get value from redis:", remoteAddr, "r:", err)
        c.Close();
        return
    }

    // Get remote host
    remoteHost := data["d"]

    // Check if no host
    if remoteHost == "" {
        log.Println("No mapping found for IP:", remoteAddr)
        c.Close();
        return
    }

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

