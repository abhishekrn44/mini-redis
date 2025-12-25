package core

import (
	"errors"
	"io"
	"log"
)

func EvaluateAndRespond(command *RedisCommand, c io.ReadWriter) error {

	log.Println("command :: ", command.Command)

	switch command.Command {
	case "PING":
		return EvaluatePING(command.Args, c)
	case "CLIENT":
		return EvaluateCLIENT(command.Args, c)
	default:
		return EvaluatePING(command.Args, c)
	}

}

func EvaluatePING(args []string, c io.ReadWriter) error {
	var buff []byte

	if len(args) >= 2 {
		return errors.New("ERR wrong number of arguments for 'ping' command")
	}

	if len(args) == 0 {
		buff = Encode("PONG", true)
	} else {
		buff = Encode(args[0], false)
	}

	_, err := c.Write(buff)
	return err
}

func EvaluateCLIENT(args []string, c io.ReadWriter) error {
	var buff []byte = Encode("OK", true)

	_, err := c.Write(buff)
	return err
}
