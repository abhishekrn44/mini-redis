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
	case "DEL":
		return evaluateDEL(command.Args, c)
	case "EXPIRE":
		return evaluateEXPIRE(command.Args, c)
	case "CLIENT":
		return evaluateCLIENT(c)
	default:
		return unknownCommand(command)
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
		c.Write(Encode(-2, false))
		return nil
	}

	// return RESP -1 if the key exists but has no associated expiration.
	if obj.Expiry == -1 {
		c.Write(Encode(-1, false))
		return nil
	}

	// Key remaining time
	durationMs := obj.Expiry - time.Now().UnixMilli()
	c.Write(Encode(durationMs/1000, false))
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

func unknownCommand(cmd *RedisCommand) error {
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

func evaluateCLIENT(c io.ReadWriter) error {
	var buff []byte = Encode("OK", true)
	_, err := c.Write(buff)
	return err
}

func evaluateDEL(args []string, c io.ReadWriter) error {

	if len(args) == 0 {
		return errors.New("ERR wrong number of arguments for 'del' command")
	}

	delCount := 0

	for _, key := range args {
		if res := DeleteKey(key); res {
			delCount++
		}
	}

	c.Write(Encode(delCount, false))
	return nil
}

func evaluateEXPIRE(args []string, c io.ReadWriter) error {

	if len(args) != 2 {
		return errors.New("ERR wrong number of arguments for 'expire' command")
	}

	key := args[0]

	secs, err := strconv.ParseInt(args[1], 10, 64)

	if err != nil {
		return errors.New("(error) ERR value is not an integer or out of range")
	}

	obj := Get(key)

	// return integet 0 if the timeout was not set; for example, the key doesn't exist, or the operation was skipped because of the provided arguments.
	if obj == nil {
		c.Write([]byte(":0\r\n"))
		return nil
	}

	obj.Expiry = time.Now().UnixMilli() + secs*1000

	// return integer 1 if the timeout was set.
	c.Write([]byte(":1\r\n"))
	return nil
}
