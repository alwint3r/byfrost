package main

import (
	"log"
	"net"
	"os"
	"os/signal"
)

var signalChan = make(chan os.Signal, 1)

const outputDir = "./output"

func main() {
	addr := net.TCPAddr{
		IP:   net.IPv4(0, 0, 0, 0),
		Port: 8191,
	}
	tcp, err := net.ListenTCP("tcp", &addr)
	if err != nil {
		log.Panic(err)
	}

	signal.Notify(signalChan, os.Interrupt)

	go func() {
		<-signalChan

		log.Println("shutting down")
		tcp.Close()
	}()

	log.Println("listening on", addr)

	for {
		conn, err := tcp.Accept()
		if err != nil {
			switch err.(type) {
			case *net.OpError:
				opError := err.(*net.OpError)
				if opError.Err == net.ErrClosed {
					log.Println("Connection closed")
				}
				return
			default:
				log.Panic(err)
			}
		}

		log.Printf("Accepted connection from %s", conn.RemoteAddr())

		go process(conn)
	}
}

func process(conn net.Conn) {
	defer conn.Close()

	fServer := InitFileServerContext()

	for {
		buf := make([]byte, 1)
		_, err := conn.Read(buf)
		if err != nil {
			if err == os.ErrClosed {
				log.Println("Connection closed")
			}
			return
		}

		nextState, err := fServer.Process(buf[0])
		if err != nil {
			log.Println(err)
			return
		}

		if nextState == Finished {
			log.Printf("File: %s, Size: %d", fServer.FileName, len(fServer.FileContent))

			f, err := os.Create(outputDir + "/" + fServer.FileName)
			if err != nil {
				log.Println(err)
				return
			}

			written, err := f.Write(fServer.FileContent)
			if err != nil {
				log.Println(err)
				return
			}

			log.Printf("Wrote %d bytes", written)

			f.Close()

			fServer.ResetState()
		}
	}
}
