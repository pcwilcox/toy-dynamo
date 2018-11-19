// tcp.go
//
// CMPS 128 Fall 2018
//
// Lawrence Lawson         lelawson
// Pete Wilcox             pcwilcox
// Annie Shen              ashen7
// Victoria Tran           ???
//
// Defines a module for communicating between replicas by setting up TCP connections and using
// them to send KVS entries as messages.
//
// The structure and design of this code is based on this blog post: https://appliedgo.net/networking/

package main

import (
	"bufio"
	"encoding/gob"
	"io"
	"log"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
)

//Dummy functions, will be removed when gossip is complete
func ClockPrune(tg timeGlob) timeGlob {
	return tg
}

//Dummy functions, will be removed when gossip is complete
func UpdatingNewEntries(eg entryGlob) {
}

// A timeGlob is a map of keys to timestamps and lets the gossip module figure out which ones need to be updated
type timeGlob struct {
	List map[string]time.Time
}

// An entryGlob is a map of keys to entries which allowes the gossip module to enter into conflict resolution and update the required keys
type entryGlob struct {
	Keys map[string]KeyEntry
}

// Open connects to a TCP Address.
// It returns a TCP connection armed with a timeout and wrapped into a buffered ReadWriter.
func Open(addr string) (*bufio.ReadWriter, error) {
	// Dial the remote process.
	log.Println("Dial " + addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, errors.Wrap(err, "Dialing "+addr+" failed")
	}
	return bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)), nil
}

// HandleFunc is a function that handles an incoming command.
// It receives the open connection wrapped in a `ReadWriter` interface.
type HandleFunc func(*bufio.ReadWriter)

// Endpoint provides an endpoint to other processess
// that they can send data to.
type Endpoint struct {
	listener net.Listener
	handler  map[string]HandleFunc

	// Maps are not threadsafe, so we need a mutex to control access.
	m sync.RWMutex
}

// NewEndpoint creates a new endpoint. Too keep things simple,
// the endpoint listens on a fixed port number.
func NewEndpoint() *Endpoint {
	// Create a new Endpoint with an empty list of handler funcs.
	return &Endpoint{
		handler: map[string]HandleFunc{},
	}
}

// AddHandleFunc adds a new function for handling incoming data.
func (e *Endpoint) AddHandleFunc(name string, f HandleFunc) {
	e.m.Lock()
	e.handler[name] = f
	e.m.Unlock()
}

// Listen starts listening on the endpoint port on all interfaces.
// At least one handler function must have been added
// through AddHandleFunc() before.
func (e *Endpoint) Listen() error {
	var err error
	e.listener, err = net.Listen("tcp", port)
	if err != nil {
		return errors.Wrapf(err, "Unable to listen on port %s\n", port)
	}
	log.Println("Listen on", e.listener.Addr().String())
	for {
		log.Println("Accept a connection request.")
		conn, err := e.listener.Accept()
		if err != nil {
			log.Println("Failed accepting a connection request:", err)
			continue
		}
		log.Println("Handle incoming messages.")
		go e.handleMessages(conn)
	}
}

// handleMessages reads the connection up to the first newline.
// Based on this string, it calls the appropriate HandleFunc.
func (e *Endpoint) handleMessages(conn net.Conn) {
	// Wrap the connection into a buffered reader for easier reading.
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	defer conn.Close()

	// Read from the connection until EOF. Expect a command name as the
	// next input. Call the handler that is registered for this command.
	for {
		log.Print("Receive command '")
		cmd, err := rw.ReadString('\n')
		switch {
		case err == io.EOF:
			log.Println("Reached EOF - close this connection.\n   ---")
			return
		case err != nil:
			log.Println("\nError reading command. Got: '"+cmd+"'\n", err)
			return
		}
		// Trim the request string - ReadString does not strip any newlines.
		cmd = strings.Trim(cmd, "\n ")
		log.Println(cmd + "'")

		// Fetch the appropriate handler function from the 'handler' map and call it.
		e.m.RLock()
		handleCommand, ok := e.handler[cmd]
		e.m.RUnlock()
		if !ok {
			log.Println("Command '" + cmd + "' is not registered.")
			return
		}
		handleCommand(rw)
	}
}

// handleGob handles the "GOB" request. It decodes the received GOB data
// into a struct.
func handleTimeGob(rw *bufio.ReadWriter) {
	log.Print("Receive Time Gob data:")
	var data timeGlob
	// Create a decoder that decodes directly into a struct variable.
	dec := gob.NewDecoder(rw)
	err := dec.Decode(&data)
	if err != nil {
		log.Println("Error decoding GOB data:", err)
		return
	}

	data = ClockPrune(data)

	enc := gob.NewEncoder(rw)
	n, err := rw.WriteString("GOB\n")

	if err != nil {
		log.Println("Could not write GOB data (" + strconv.Itoa(n) + " bytes written)")
	}
	err = enc.Encode(data)
	if err != nil {
		log.Println("Encode failed for struct: ", data)
	}
	err = rw.Flush()
	if err != nil {
		log.Println("Flush failed.")
	}

	log.Printf("Outer complexData struct: \n%v", data)
}

func handleEntryGob(rw *bufio.ReadWriter) {
	log.Print("Receive GOB data:")
	var data entryGlob
	// Create a decoder that decodes directly into a struct variable.
	dec := gob.NewDecoder(rw)
	err := dec.Decode(&data)
	if err != nil {
		log.Println("Error decoding GOB data:", err)
		return
	}

	UpdatingNewEntries(data)
	// Print the complexData struct and the nested one, too, to prove
	// that both travelled across the wire.
	log.Printf("Outer complexData struct: \n%#v\n", data)
}

/*
## The client and server functions

With all this in place, we can now set up client and server functions.

The client function connects to the server and sends STRING and GOB requests.

The server starts listening for requests and triggers the appropriate handlers.
*/

// client is called if the app is called with -connect=`ip addr`.
func sendTimeGlob(ip string, tg timeGlob) (*timeGlob, error) {
	// Open a connection to the server.

	var out timeGlob

	rw, err := Open(ip)
	if err != nil {
		return nil, errors.Wrap(err, "Client: Failed to open connection to "+ip)
	}

	enc := gob.NewEncoder(rw)
	n, err := rw.WriteString("GOB\n")

	if err != nil {
		return nil, errors.Wrap(err, "Could not write GOB data ("+strconv.Itoa(n)+" bytes written)")
	}

	err = enc.Encode(tg)
	if err != nil {
		return nil, errors.Wrapf(err, "Encode failed for struct: %#v", tg)
	}
	err = rw.Flush()
	if err != nil {
		return nil, errors.Wrap(err, "Flush failed.")
	}

	// Create a decoder that decodes directly into a struct variable.
	dec := gob.NewDecoder(rw)
	err = dec.Decode(&out)
	if err != nil {
		log.Println("Error decoding GOB data:", err)
		return nil, err
	}

	return &out, nil
}

func sendEntryGlob(ip string, eg entryGlob) error {
	// Open a connection to the server.

	rw, err := Open(ip)
	if err != nil {
		return errors.Wrap(err, "Client: Failed to open connection to "+ip)
	}

	// Send the request name.

	enc := gob.NewEncoder(rw)
	n, err := rw.WriteString("GOB\n")

	if err != nil {
		return errors.Wrap(err, "Could not write GOB data ("+strconv.Itoa(n)+" bytes written)")
	}
	err = enc.Encode(eg)
	if err != nil {
		return errors.Wrapf(err, "Encode failed for struct: %#v", eg)
	}
	err = rw.Flush()
	if err != nil {
		return errors.Wrap(err, "Flush failed.")
	}

	return nil
}

// server listens for incoming requests and dispatches them to
// registered handler functions.
func server() error {
	endpoint := NewEndpoint()
	// Add HandleTimeGob
	endpoint.AddHandleFunc("time", handleTimeGob)
	// Add HandleEntryGob
	endpoint.AddHandleFunc("entry", handleEntryGob)
	// Start listening.
	return endpoint.Listen()
}

/*
## Main

Main starts either a client or a server, depending on whether the `connect`
flag is set. Without the flag, the process starts as a server, listening
for incoming requests. With the flag the process starts as a client and connects
to the host specified by the flag value.

Try "localhost" or "127.0.0.1" when running both processes on the same machine.

*/

// main
/*
func main() {
	connect := flag.String("connect", "", "IP address of process to join. If empty, go into listen mode.")
	flag.Parse()

	// If the connect flag is set, go into client mode.
	if *connect != "" {
		err := client(*connect)
		if err != nil {
			log.Println("Error:", errors.WithStack(err))
		}
		log.Println("Client done.")
		return
	}

	// Else go into server mode.
	err := server()
	if err != nil {
		log.Println("Error:", errors.WithStack(err))
	}

	log.Println("Server done.")
}
*/
// The Lshortfile flag includes file name and line number in log messages.
func init() {
	log.SetFlags(log.Lshortfile)
}

/*
## How to get and run the code

Step 1: `go get` the code. Note the `-d` flag that prevents auto-installing
the binary into `$GOPATH/bin`.

    go get -d github.com/appliedgo/networking

Step 2: `cd` to the source code directory.

    cd $GOPATH/src/github.com/appliedgo/networking

Step 3. Run the server.

    go run networking.go

Step 4. Open another shell, `cd` to the source code (see Step 2), and
run the client.

    go run networking.go -connect localhost
*/
