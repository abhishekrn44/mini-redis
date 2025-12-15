package server

import (
	"io"
	"log"
	"mini-redis/config"
	"net"
	"strconv"
)

func StartSyncTCPServer() {
	log.Println("TCP Server starting", config.Host, config.Port)

	listener, error := net.Listen("tcp", config.Host+":"+strconv.Itoa(config.Port))

	if error != nil {
		panic(error)
	}

	for {
		connection, error := listener.Accept()

		if error != nil {
			panic(error)
		}

		log.Println("Client connected with address ", connection.RemoteAddr())

		for {
			command, error := readCommand(connection)

			if error != nil {
				connection.Close()
				log.Println("Client disconnected", connection.RemoteAddr())
				log.Println("Error", error)

				if error == io.EOF {
					break
				}

			}

			if error = respond(connection, command); error != nil {
				log.Println("Error", error)
			}
		}
	}

}

func readCommand(connection net.Conn) (string, error) {
	var buff []byte = make([]byte, 512)

	count, error := connection.Read(buff[:])

	if error != nil {
		return "", error
	}

	return string(buff[:count]), nil
}

func respond(connection net.Conn, command string) error {

	_, error := connection.Write([]byte(command))

	if error != nil {
		return error
	}

	return nil
}
