package core

import (
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"
)

func EvaluateAndRespond(command *RedisCommand, c io.ReadWriter) error {

	log.Println("Command", command.Command)

	switch command.Command {
	case "PING":
		return evaluatePING(command.Args, c)
	case "SET":
		return evaluateSET(command.Args, c)
	case "GET":
		return evaluateGET(command.Args, c)
	case "TTL":
		return evaluateTTL(command.Args, c)
	case "CLIENT":
		return evaluateCLIENT(command.Args, c)
	default:
		return unknownCommand(command, c)
	}

}

func evaluatePING(args []string, c io.ReadWriter) error {
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

func evaluateSET(args []string, c io.ReadWriter) error {
	if len(args) <= 1 {
		return errors.New("(ERR wrong number of arguments for 'set' command")
	}

	var key, value string
	var expireDurationMs int64 = -1

	key, value = args[0], args[1]

	for i := 2; i < len(args); i++ {
		switch args[i] {
		case "EX", "ex":
			i++
			if i == len(args) {
				return errors.New("ERR syntax error")
			}

			expireDurationSecs, error := strconv.ParseInt(args[i], 10, 64)

			if error != nil {
				return errors.New("ERR value is not an integer or out of range")
			}

			expireDurationMs = expireDurationSecs * 1000
		default:
			return errors.New("ERR syntax error")
		}
	}

	// putting the k and value in map
	Put(key, NewObj(value, expireDurationMs))

	c.Write([]byte("+OK\r\n"))
	return nil
}

func evaluateTTL(args []string, c io.ReadWriter) error {
	if len(args) != 1 {
		return errors.New("ERR wrong number of arguments for 'ttl' command")
	}

	obj := Get(args[0])

	// return RESP -2 if the key does not exist.
	if obj == nil {
		c.Write(Encode(int64(-2), false))
		return nil
	}

	// return RESP -1 if the key exists but has no associated expiration.
	if obj.Expiry == -1 {
		c.Write(Encode(int64(-1), false))
		return nil
	}

	// Key remaining time
	durationMs := obj.Expiry - time.Now().UnixMilli()
	c.Write(Encode(int64(durationMs/1000), false))
	return nil
}

func evaluateGET(args []string, c io.ReadWriter) error {
	if len(args) != 1 {
		return errors.New("ERR wrong number of arguments for 'get' command")
	}

	obj := Get(args[0])

	// if the key does not exist then return nil.
	if obj == nil {
		c.Write(NilRESP)
		return nil
	}

	// if key already expired then return nil.
	if obj.Expiry != -1 && obj.Expiry <= time.Now().UnixMilli() {
		c.Write(NilRESP)
		return nil
	}

	// encode Value
	c.Write(Encode(obj.Value, false))
	return nil
}

func unknownCommand(cmd *RedisCommand, c io.ReadWriter) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("ERR unknown command '%s', with args beginning with:", cmd.Command)
	} else {
		return fmt.Errorf(
			"ERR unknown command '%s', with args beginning with: '%s'",
			cmd.Command,
			strings.Join(cmd.Args, "' '"),
		)
	}
}

func evaluateCLIENT(args []string, c io.ReadWriter) error {
	var buff []byte = Encode("OK", true)
	_, err := c.Write(buff)
	return err
}
