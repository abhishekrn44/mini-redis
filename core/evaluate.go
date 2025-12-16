package core

import (
	"errors"
	"log"
	"net"
)

func EvaluateAndRespond(command *RedisCommand, conn net.Conn) error {

	log.Println("command :: ", command.Command)

	switch command.Command {
	case "PING":
		return EvaluatePING(command.Args, conn)
	default:
		return EvaluatePING(command.Args, conn)
	}

}

func EvaluatePING(args []string, conn net.Conn) error {
	var buff []byte

	if len(args) >= 2 {
		return errors.New("ERR wrong number of arguments for 'ping' command")
	}

	if len(args) == 0 {
		buff = Encode("PONG", true)
	} else {
		buff = Encode(args[0], false)
	}

	_, err := conn.Write(buff)
	return err
}
