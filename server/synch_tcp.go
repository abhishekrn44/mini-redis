package server

import (
	"fmt"
	"io"
	"log"
	"mini-redis/config"
	"mini-redis/core"
	"net"
	"strconv"
	"strings"
)

func StartSyncTCPServer() {
	log.Println("TCP Server starting", config.Host+":"+strconv.Itoa(config.Port))

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
			command, error := ReadCommand(connection)

			if error != nil {
				connection.Close()
				log.Println("Client disconnected", connection.RemoteAddr())
				log.Println("Error", error)

				if error == io.EOF {
					break
				}

			}

			Respond(command, connection)
		}
	}

}

// func ReadCommand(connection net.Conn) (*core.RedisCommand, error) {
// 	var buff []byte = make([]byte, 512)

// 	count, error := connection.Read(buff[:])

// 	if error != nil {
// 		return nil, error
// 	}

// 	tokens, error := core.DecodeArrayString(buff[:count])

// 	if error != nil {
// 		return nil, error
// 	}

// 	return &core.RedisCommand{
// 		Command: strings.ToUpper(tokens[0]),
// 		Args:    tokens[1:],
// 	}, nil
// }

// func Respond(connection net.Conn, command *core.RedisCommand) {
// 	err := core.EvaluateAndRespond(command, connection)
// 	if err != nil {
// 		RespondError(err, connection)
// 	}
// }

// func RespondError(err error, c net.Conn) {
// 	c.Write([]byte(fmt.Sprintf("-%s\r\n", err)))
// }

func ReadCommand(c io.ReadWriter) (*core.RedisCommand, error) {
	var buff []byte = make([]byte, 512)

	count, error := c.Read(buff[:])

	if error != nil {
		return nil, error
	}

	tokens, error := core.DecodeArrayString(buff[:count])

	log.Println("cmd", tokens)

	if error != nil {
		return nil, error
	}

	return &core.RedisCommand{
		Command: strings.ToUpper(tokens[0]),
		Args:    tokens[1:],
	}, nil
}

func Respond(command *core.RedisCommand, c io.ReadWriter) {
	err := core.EvaluateAndRespond(command, c)
	if err != nil {
		RespondError(err, c)
	}
}

func RespondError(err error, c io.ReadWriter) {
	c.Write([]byte(fmt.Sprintf("-%s\r\n", err)))
}
