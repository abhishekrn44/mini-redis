package core

import (
	"bufio"
	"net"
	"os"
	"testing"
	"time"
)

func serverAddr() string {
	if v := os.Getenv("REDIS_ADDR"); v != "" {
		return v
	}
	return "127.0.0.1:6379"
}

func dial(t *testing.T) net.Conn {
	t.Helper()

	conn, err := net.DialTimeout("tcp", serverAddr(), 2*time.Second)
	if err != nil {
		t.Fatalf("dial failed: %v", err)
	}

	// Avoid hanging forever if the server replies only once / stalls.
	_ = conn.SetDeadline(time.Now().Add(2 * time.Second))
	return conn
}

func mustReadLine(t *testing.T, r *bufio.Reader) string {
	t.Helper()

	line, err := r.ReadString('\n') // reads until LF
	if err != nil {
		t.Fatalf("read failed: %v", err)
	}
	return line
}

// RESP command payloads (exact bytes).
const (
	cmdSetInfoLibName = "*4\r\n$6\r\nCLIENT\r\n$7\r\nSETINFO\r\n$8\r\nlib-name\r\n$7\r\nLettuce\r\n"
	cmdSetInfoLibVer  = "*4\r\n$6\r\nCLIENT\r\n$7\r\nSETINFO\r\n$7\r\nlib-ver\r\n$5\r\n6.3.2\r\n"
)

// Test 1: pipelining (two commands back-to-back, one write)
func TestRESP_Pipelining_TwoClientSetInfo(t *testing.T) {
	conn := dial(t)
	defer conn.Close()

	payload := []byte(cmdSetInfoLibName + cmdSetInfoLibVer)

	n, err := conn.Write(payload)
	if err != nil {
		t.Fatalf("write failed: %v", err)
	}
	if n != len(payload) {
		// Non-fatal warning: on blocking TCP it should usually write all.
		t.Fatalf("partial write: wrote %d of %d bytes", n, len(payload))
	}

	r := bufio.NewReader(conn)

	got1 := mustReadLine(t, r)
	if got1 != "+OK\r\n" {
		t.Fatalf("reply1 mismatch: got %q, want %q", got1, "+OK\r\n")
	}

	got2 := mustReadLine(t, r)
	if got2 != "+OK\r\n" {
		t.Fatalf("reply2 mismatch: got %q, want %q", got2, "+OK\r\n")
	}
}

// Test 2: fragmentation (one command split into multiple writes)
func TestRESP_Fragmentation_SingleCommandInParts(t *testing.T) {
	conn := dial(t)
	defer conn.Close()

	// Split mid-frame intentionally to ensure your server supports stream parsing.
	part1 := []byte("*4\r\n$6\r\nCLIENT\r\n$7\r\nSETINFO\r\n")
	part2 := []byte("$8\r\nlib-name\r\n")
	part3 := []byte("$7\r\nLettuce\r\n")

	if _, err := conn.Write(part1); err != nil {
		t.Fatalf("write part1 failed: %v", err)
	}
	time.Sleep(50 * time.Millisecond)

	if _, err := conn.Write(part2); err != nil {
		t.Fatalf("write part2 failed: %v", err)
	}
	time.Sleep(50 * time.Millisecond)

	if _, err := conn.Write(part3); err != nil {
		t.Fatalf("write part3 failed: %v", err)
	}

	r := bufio.NewReader(conn)
	got := mustReadLine(t, r)
	if got != "+OK\r\n" {
		t.Fatalf("reply mismatch: got %q, want %q", got, "+OK\r\n")
	}
}
