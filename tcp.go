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

	"github.com/pkg/errors"
	"github.com/soheilhy/cmux"
)

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
	var data TimeGlob

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
	var data EntryGlob
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

func (e *Endpoint) handleShardGob(rw *bufio.ReadWriter) {
	var data ShardGlob
	dec := gob.NewDecoder(rw)
	log.Println("Decoding viewGob data")
	err := dec.Decode(&data)
	if err != nil {
		log.Println("Error decoding view data")
		return
	}

	log.Println("Updating ShardList - old views: " + e.gossip.shardList.String())
	e.gossip.UpdateShardList(data)
	log.Println("Shards updated: ", data)
}

func (e *Endpoint) handleHelp(rw *bufio.ReadWriter) {
	log.Println("Receive call for help")
	wakeGossip = true
}

func (e *Endpoint) handleDelete(rw *bufio.ReadWriter) {
	log.Println("Handling delete()")
	var data PutRequest
	dec := gob.NewDecoder(rw)
	log.Println("Decoding PutRequest data")
	err := dec.Decode(&data)
	if err != nil {
		log.Println("error decoding request data")
		return
	}

	log.Println("Received request: ", data)

	written := e.gossip.kvs.Delete(data.Key, data.Timestamp, data.Payload)

	var resp string
	if written {
		resp = "true\n"
	} else {
		resp = "false\n"
	}
	log.Println("Writing back response: ", resp)
	n, err := rw.WriteString(resp)
	if err != nil {
		log.Println("error encoding response:", err)
		log.Println("Could not write response data (" + strconv.Itoa(n) + " bytes written)")
	}
	log.Println("flushing buffer")
	err = rw.Flush()
	if err != nil {
		log.Println("error flushing buffer:", err)
	}
}

func (e *Endpoint) handlePut(rw *bufio.ReadWriter) {
	log.Println("Handling put()")
	var data PutRequest
	dec := gob.NewDecoder(rw)
	log.Println("Decoding PutRequest data")
	err := dec.Decode(&data)
	if err != nil {
		log.Println("error decoding request data")
		return
	}

	log.Println("Received request: ", data)

	written := e.gossip.kvs.Put(data.Key, data.Value, data.Timestamp, data.Payload)

	var resp string
	if written {
		resp = "true\n"
	} else {
		resp = "false\n"
	}
	log.Println("Writing back response: ", resp)
	n, err := rw.WriteString(resp)
	if err != nil {
		log.Println("error encoding response:", err)
		log.Println("Could not write response data (" + strconv.Itoa(n) + " bytes written)")
		return
	}
	log.Println("flushing buffer")
	err = rw.Flush()
	if err != nil {
		log.Println("error flushing buffer:", err)
	}
}

func (e *Endpoint) handleGet(rw *bufio.ReadWriter) {
	log.Println("Handling get()")
	var data GetRequest
	dec := gob.NewDecoder(rw)
	log.Println("Decoding GetRequest data")
	err := dec.Decode(&data)
	if err != nil {
		log.Println("error decoding request data")
		return
	}

	log.Println("Received request: ", data)

	value, payload := e.gossip.kvs.Get(data.Key, data.Payload)

	resp := GetResponse{Value: value, Payload: payload}

	enc := gob.NewEncoder(rw)
	log.Println("Encoding respoonse back: ", resp)
	err = enc.Encode(resp)
	if err != nil {
		log.Println("error encoding response:", err)
		return
	}
	log.Println("flushing buffer")
	err = rw.Flush()
	if err != nil {
		log.Println("error flushing buffer:", err)
	}
}

func (e *Endpoint) handleContains(rw *bufio.ReadWriter) {
	log.Println("Handling contains()")
	var data GetRequest
	dec := gob.NewDecoder(rw)
	log.Println("Decoding GetRequest data")
	err := dec.Decode(&data)
	if err != nil {
		log.Println("error decoding request data")
		return
	}

	log.Println("Received request: ", data)

	alive, version := e.gossip.kvs.Contains(data.Key)

	resp := ContainsResponse{Alive: alive, Version: version}

	enc := gob.NewEncoder(rw)
	log.Println("Encoding respoonse back: ", resp)
	err = enc.Encode(resp)
	if err != nil {
		log.Println("error encoding response:", err)
		return
	}
	log.Println("flushing buffer")
	err = rw.Flush()
	if err != nil {
		log.Println("error flushing buffer:", err)
	}
}

// client is called if the app is called with -connect=`ip addr`.
func sendTimeGlob(ip string, tg TimeGlob) (*TimeGlob, error) {
	// Open a connection to the server.
	var out TimeGlob

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

func sendEntryGlob(ip string, eg EntryGlob) error {

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

func sendShardGob(ip string, s ShardGlob) error {
	rw, err := Open(ip)
	if err != nil {
		return errors.Wrap(err, "Client: failed to open connection to "+ip)
	}
	enc := gob.NewEncoder(rw)
	log.Println("Sending command initialization: 'shard'")
	n, err := rw.WriteString("shard\n")
	if err != nil {
		return errors.Wrap(err, "Could not write shard data ("+strconv.Itoa(n)+" bytes written)")
	}

	log.Println("Encoding shard struct")
	err = enc.Encode(s)
	if err != nil {
		return errors.Wrapf(err, "Encode failed for shard list: %#v", s)

	}
	log.Println("Flushing buffer")
	err = rw.Flush()
	if err != nil {
		return errors.Wrap(err, "Flush failed")
	}
	return nil
}
func sendDeleteRequest(ip string, p PutRequest) (bool, error) {
	rw, err := Open(ip)
	if err != nil {
		return false, errors.Wrap(err, "Client: failed to open connection to "+ip)
	}
	enc := gob.NewEncoder(rw)
	log.Println("Sending command initialization: 'delete'")
	n, err := rw.WriteString("delete\n")
	if err != nil {
		return false, errors.Wrap(err, "Could not write delete data ("+strconv.Itoa(n)+" bytes written)")
	}

	log.Println("Encoding PutRequest")
	err = enc.Encode(p)
	if err != nil {
		return false, errors.Wrapf(err, "Encode failed for PutRequest: %#v", p)

	}
	log.Println("Flushing buffer")
	err = rw.Flush()
	if err != nil {
		return false, errors.Wrap(err, "Flush failed")
	}

	log.Println("Reading response")
	resp, err := rw.ReadString('\n')
	if err != nil {
		log.Println("Error decoding response data:", err)
		return false, err
	}
	resp = strings.Trim(resp, "\n ")
	if resp == "true" {
		return true, nil
	}

	return false, nil
}

func sendPutRequest(ip string, p PutRequest) (bool, error) {
	rw, err := Open(ip)
	if err != nil {
		return false, errors.Wrap(err, "Client: failed to open connection to "+ip)
	}
	enc := gob.NewEncoder(rw)
	log.Println("Sending command initialization: 'put'")
	n, err := rw.WriteString("put\n")
	if err != nil {
		return false, errors.Wrap(err, "Could not write put data ("+strconv.Itoa(n)+" bytes written)")
	}

	log.Println("Encoding PutRequest")
	err = enc.Encode(p)
	if err != nil {
		return false, errors.Wrapf(err, "Encode failed for PutRequest: %#v", p)

	}
	log.Println("Flushing buffer")
	err = rw.Flush()
	if err != nil {
		return false, errors.Wrap(err, "Flush failed")
	}

	log.Println("Reading response")
	resp, err := rw.ReadString('\n')
	if err != nil {
		log.Println("Error decoding response data:", err)
		return false, err
	}
	resp = strings.Trim(resp, "\n ")
	if resp == "true" {
		return true, nil
	}

	return false, nil
}

func sendContainsRequest(ip string, g GetRequest) (ContainsResponse, error) {
	rw, err := Open(ip)
	if err != nil {
		return ContainsResponse{}, errors.Wrap(err, "Client: failed to open connection to "+ip)
	}
	enc := gob.NewEncoder(rw)
	log.Println("Sending command initialization: 'contains'")
	n, err := rw.WriteString("contains\n")
	if err != nil {
		return ContainsResponse{}, errors.Wrap(err, "Could not write contains data ("+strconv.Itoa(n)+" bytes written)")
	}

	log.Println("Encoding GetRequest")
	err = enc.Encode(g)
	if err != nil {
		return ContainsResponse{}, errors.Wrapf(err, "Encode failed for GetRequest: %#v", g)

	}
	log.Println("Flushing buffer")
	err = rw.Flush()
	if err != nil {
		return ContainsResponse{}, errors.Wrap(err, "Flush failed")
	}

	var data ContainsResponse

	// Create a decoder that decodes directly into a struct variable.
	dec := gob.NewDecoder(rw)
	log.Println("Reading ContainsResponse")
	err = dec.Decode(&data)
	if err != nil {
		log.Println("Error decoding GOB data:", err)
		return ContainsResponse{}, err
	}

	return data, nil
}

func sendGetRequest(ip string, g GetRequest) (GetResponse, error) {
	rw, err := Open(ip)
	if err != nil {
		return GetResponse{}, errors.Wrap(err, "Client: failed to open connection to "+ip)
	}
	enc := gob.NewEncoder(rw)
	log.Println("Sending command initialization: 'get'")
	n, err := rw.WriteString("get\n")
	if err != nil {
		return GetResponse{}, errors.Wrap(err, "Could not write get data ("+strconv.Itoa(n)+" bytes written)")
	}

	log.Println("Encoding GetRequest")
	err = enc.Encode(g)
	if err != nil {
		return GetResponse{}, errors.Wrapf(err, "Encode failed for GetRequest: %#v", g)

	}
	log.Println("Flushing buffer")
	err = rw.Flush()
	if err != nil {
		return GetResponse{}, errors.Wrap(err, "Flush failed")
	}

	var data GetResponse

	// Create a decoder that decodes directly into a struct variable.
	dec := gob.NewDecoder(rw)
	log.Println("Reading GetResponse")
	err = dec.Decode(&data)
	if err != nil {
		log.Println("Error decoding GOB data:", err)
		return GetResponse{}, err
	}

	return data, nil
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

	// Add functions to it
	addHandlers(endpoint)
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

func addHandlers(e *Endpoint) {
	e.AddHandleFunc("time", e.handleTimeGob)
	e.AddHandleFunc("entry", e.handleEntryGob)
	e.AddHandleFunc("shard", e.handleShardGob)
	e.AddHandleFunc("help", e.handleHelp)
	e.AddHandleFunc("contains", e.handleContains)
	e.AddHandleFunc("get", e.handleGet)
	e.AddHandleFunc("put", e.handlePut)
	e.AddHandleFunc("delete", e.handleDelete)
}

func register() {
	// Register types for gob
	gob.Register(TimeGlob{})
	gob.Register(EntryGlob{})
	gob.Register(Entry{})
	gob.Register(GetRequest{})
	gob.Register(PutRequest{})
	gob.Register(GetResponse{})
	gob.Register(ContainsResponse{})
	gob.Register(ShardGlob{})
}
