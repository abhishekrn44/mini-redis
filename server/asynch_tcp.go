package server

import (
	"log"
	"mini-redis/config"
	"mini-redis/core"
	"net"
	"strconv"
	"syscall"
)

var con_clients int = 0

func StartAsyncTCPServer() error {
	log.Println("TCP Server starting", config.Host+":"+strconv.Itoa(config.Port))

	max_clients := 10000

	// Create slice of EpollEvents to store events objects
	var events []syscall.EpollEvent = make([]syscall.EpollEvent, max_clients)

	// Create a kernel-level IPv4 TCP socket in non-blocking mode.
	// - AF_INET       : IPv4 address family
	// - SOCK_STREAM   : TCP (connection-oriented, reliable stream)
	// - O_NONBLOCK    : All I/O operations return immediately (required for epoll)
	// - protocol = 0  : Let the kernel choose TCP by default
	serverFD, error := syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)

	log.Println("Main server socket file descriptor", serverFD)

	if error != nil {
		return error
	}

	defer syscall.Close(serverFD)

	syscall.SetNonblock(serverFD, true)

	// Set the socket operate in Non-Blocking mode
	if error := syscall.SetNonblock(serverFD, true); error != nil {
		return error
	}

	// Bind IP and Port
	ip4 := net.ParseIP(config.Host)

	if error := syscall.Bind(serverFD, &syscall.SockaddrInet4{
		Port: config.Port,
		Addr: [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]},
	}); error != nil {
		return error
	}

	// Start listening
	if error := syscall.Listen(serverFD, max_clients); error != nil {
		return error
	}

	// Asynchronous Magic starts here!

	// Create Epoll instance
	epollFD, err := syscall.EpollCreate1(0)

	if err != nil {
		log.Fatal(err)
	}

	defer syscall.Close(epollFD)

	// Specify the events we want to get hints about and set the socket on which

	var socketServerEvents syscall.EpollEvent = syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     int32(serverFD),
	}

	// Listen to read events on the server itself
	if err := syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, serverFD, &socketServerEvents); err != nil {
		return err
	}

	for {
		// see if any FD is ready for an IO
		no_of_events, e := syscall.EpollWait(epollFD, events[:], -1)

		if e != nil {
			continue
		}

		for i := 0; i < no_of_events; i++ {
			// if the socket server is ready for an IO
			if int(events[i].Fd) == serverFD {
				// accept the incoming connection from a client
				fd, _, err := syscall.Accept(serverFD)
				if err != nil {
					log.Println("error", err)
					continue
				}

				// increase the number of concurrent clients count
				con_clients++

				log.Println("concurrent clients count", con_clients)

				syscall.SetNonblock(serverFD, true)

				// add this new TCP connection to be monitored
				var socketClientEvent syscall.EpollEvent = syscall.EpollEvent{
					Events: syscall.EPOLLIN,
					Fd:     int32(fd),
				}

				if error := syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, fd, &socketClientEvent); error != nil {
					log.Fatal(error)
				}
			} else {
				comm := core.FDComm{
					Fd: int(events[i].Fd),
				}

				cmd, err := ReadCommand(comm)
				if err != nil {
					syscall.Close(int(events[i].Fd))
					con_clients -= 1
					continue
				}
				Respond(cmd, comm)
			}
		}
	}

}
