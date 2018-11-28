// tcp.go
//
// CMPS 128 Fall 2018
//
// Lawrence Lawson     lelawson
// Pete Wilcox         pcwilcox
// Annie Shen          ashen7
// Victoria Tran       vilatran
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
	"github.com/soheilhy/cmux"
)

// A timeGlob is a map of keys to timestamps and lets the gossip module figure out which ones need to be updated
type timeGlob struct {
	List map[string]time.Time
}

// An entryGlob is a map of keys to entries which allowes the gossip module to enter into conflict resolution and update the required keys
type entryGlob struct {
	Keys map[string]Entry
}

// Open connects to a TCP Address.
// It returns a TCP connection armed with a timeout and wrapped into a buffered ReadWriter.
func Open(addr string) (*bufio.ReadWriter, error) {
	// Trim the address since we're only using the IP
	s := strings.Split(addr, ":")[0]
	s = s + port
	// Dial the remote process.
	log.Println("Dial " + s)
	conn, err := net.Dial("tcp", s)
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
	listener net.Listener          // The listener that this endpoint is attached to
	handler  map[string]HandleFunc // The handlers that this endpoint uses to process requests
	gossip   GossipVals            // The gossip module the endpoint uses
	m        sync.RWMutex          // A lock for the handler map
}

// NewEndpoint creates a new endpoint.
func NewEndpoint() *Endpoint {
	// Create a new Endpoint with an empty list of handler funcs.
	return &Endpoint{
		handler: map[string]HandleFunc{},
	}
}

// AddHandleFunc adds a new function for handling incoming data. The name is the
// string passed in at the start of the connection stream, and the handleFunc
// is the function used to handle the request.
func (e *Endpoint) AddHandleFunc(name string, f HandleFunc) {
	e.m.Lock()
	e.handler[name] = f
	e.m.Unlock()
}

// Listen starts listening on the endpoint port on all interfaces.
// At least one handler function must have been added
// through AddHandleFunc() before.
func (e *Endpoint) Listen() error {
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
		log.Println("Received command: '" + cmd + "'")

		// Fetch the appropriate handler function from the 'handler' map and call it.
		e.m.RLock()
		handleCommand, ok := e.handler[cmd]
		e.m.RUnlock()
		if !ok {
			log.Println("Command '" + cmd + "' is not registered.")
			return
		}

		// Call the appropriate function
		handleCommand(rw)
	}
}

// handleTimeGob reads the timeGob out of the request and passes it to the gossip
// module, then returns the result to the client
func (e *Endpoint) handleTimeGob(rw *bufio.ReadWriter) {
	log.Print("Receive Time Gob data:")

	// Create an empty timeGlob
	var data timeGlob

	// Create a decoder that decodes directly into the empty glob
	dec := gob.NewDecoder(rw)
	err := dec.Decode(&data)
	if err != nil {
		log.Println("Error decoding GOB data:", err)
		return
	}

	log.Printf("Decoding timeGlob: %#v\n", data)

	// Pass the data glob to the gossip module and get the result
	data = e.gossip.ClockPrune(data)

	// Create an encoder on the stream
	enc := gob.NewEncoder(rw)
	log.Printf("Encoding response timeGlob back to buffer: %#v\n", data)
	// Encode the response back to the client
	err = enc.Encode(data)
	if err != nil {
		log.Println("Encode failed for struct: ", data)
	}

	// Flush the buffer to ensure it has all been read before closing it
	log.Println("Flushing buffer")
	err = rw.Flush()
	if err != nil {
		log.Println("Flush failed.")
	}

	log.Printf("timeGlob struct: \n%v", data)
}

func (e *Endpoint) handleEntryGob(rw *bufio.ReadWriter) {
	log.Println("Receive entryGlob data:")
	var data entryGlob
	// Create a decoder that decodes directly into a struct variable.
	dec := gob.NewDecoder(rw)
	err := dec.Decode(&data)
	if err != nil {
		log.Println("Error decoding GOB data:", err)
		return
	}

	log.Println("Decoding entryGlob: ", data)
	log.Println("Updating KVS")
	e.gossip.UpdateKVS(data)
	// Print the complexData struct and the nested one, too, to prove
	// that both travelled across the wire.
	log.Printf("Outer complexData struct: \n%#v\n", data)
}

func (e *Endpoint) handleViewGob(rw *bufio.ReadWriter) {
	var data []string
	dec := gob.NewDecoder(rw)
	log.Println("Decoding viewGob data")
	err := dec.Decode(&data)
	if err != nil {
		log.Println("Error decoding view data")
		return
	}

	log.Println("Updating viewList - old views: " + e.gossip.view.String())
	e.gossip.UpdateViews(data)
	log.Println("Views updated: ", data)
}

func (e *Endpoint) handleHelp(rw *bufio.ReadWriter) {
	log.Println("Receive call for help")
	wakeGossip = true
}

// client is called if the app is called with -connect=`ip addr`.
func sendTimeGlob(ip string, tg timeGlob) (*timeGlob, error) {
	// Open a connection to the server.
	var out timeGlob

	rw, err := Open(ip)
	if err != nil {
		return nil, errors.Wrap(err, "Client: Failed to open connection to "+ip)
	}

	enc := gob.NewEncoder(rw)
	log.Println("Sending command initialization: 'time'")
	n, err := rw.WriteString("time\n")

	if err != nil {
		return nil, errors.Wrap(err, "Could not write GOB data ("+strconv.Itoa(n)+" bytes written)")
	}

	log.Println("Encoding timeGlob")
	err = enc.Encode(tg)
	if err != nil {
		return nil, errors.Wrapf(err, "Encode failed for struct: %#v", tg)
	}
	log.Println("Flushing buffer")
	err = rw.Flush()
	if err != nil {
		return nil, errors.Wrap(err, "Flush failed.")
	}

	// Create a decoder that decodes directly into a struct variable.
	dec := gob.NewDecoder(rw)
	log.Println("Reading timeGlob response")
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
	log.Println("Sending command initialization: 'entry'")
	n, err := rw.WriteString("entry\n")

	if err != nil {
		return errors.Wrap(err, "Could not write GOB data ("+strconv.Itoa(n)+" bytes written)")
	}
	log.Println("Encoding entryGlob: ", eg)
	err = enc.Encode(eg)
	if err != nil {
		return errors.Wrapf(err, "Encode failed for struct: %#v", eg)
	}
	log.Println("Flushing buffer")
	err = rw.Flush()
	if err != nil {
		return errors.Wrap(err, "Flush failed.")
	}

	return nil
}

func sendViewList(ip string, v []string) error {
	rw, err := Open(ip)
	if err != nil {
		return errors.Wrap(err, "Client: failed to open connection to "+ip)
	}
	enc := gob.NewEncoder(rw)
	log.Println("Sending command initialization: 'view'")
	n, err := rw.WriteString("view\n")
	if err != nil {
		return errors.Wrap(err, "Could not write view data ("+strconv.Itoa(n)+" bytes written)")
	}

	log.Println("Encoding view []string")
	err = enc.Encode(v)
	if err != nil {
		return errors.Wrapf(err, "Encode failed for slice: %#v", v)

	}
	log.Println("Flushing buffer")
	err = rw.Flush()
	if err != nil {
		return errors.Wrap(err, "Flush failed")
	}
	return nil
}

func askForHelp(ip string) error {
	rw, err := Open(ip)
	if err != nil {
		return errors.Wrap(err, "Client: failed to open connection to "+ip)
	}
	log.Println("Sending command initialization: 'help'")
	n, err := rw.WriteString("help\n")
	if err != nil {
		return errors.Wrap(err, "Could not write view data ("+strconv.Itoa(n)+" bytes written)")
	}

	log.Println("Flushing buffer")
	err = rw.Flush()
	return err
}

// server listens for incoming requests and dispatches them to
// registered handler functions.
func server(a App, g GossipVals) {
	// Register types for gob
	gob.Register(timeGlob{})
	gob.Register(entryGlob{})
	gob.Register(Entry{})

	// Create a  listener
	l, err := net.Listen("tcp", port)
	if err != nil {
		log.Fatalln(err)
	}
	// Create a cmux
	m := cmux.New(l)

	// Set up a matcher for HTTP
	httpl := m.Match(cmux.HTTP1())

	// Create a matcher for anything else
	tcpl := m.Match(cmux.Any())

	// Create the TCP endpoint
	endpoint := NewEndpoint()
	// Add HandleTimeGob
	endpoint.AddHandleFunc("time", endpoint.handleTimeGob)
	// Add HandleEntryGob
	endpoint.AddHandleFunc("entry", endpoint.handleEntryGob)
	// Add HandleViewListGob
	endpoint.AddHandleFunc("view", endpoint.handleViewGob)
	// Add HandleHelp
	endpoint.AddHandleFunc("help", endpoint.handleHelp)

	endpoint.listener = tcpl
	endpoint.gossip = g
	log.Println("Server has initialized")

	// Run the two listeners
	go a.Initialize(httpl)
	go endpoint.Listen()
	if err := m.Serve(); !strings.Contains(err.Error(), "use of closed network connection") {
		log.Fatalln(err)
	}
}
