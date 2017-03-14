package sse // import "github.com/digivava/go-sse-json"

import (
	"bufio"
	"fmt"
	"io"
	"net/http"

	"github.com/antonholmquist/jason"
)

//SSE name constants
const (
	eName = "event"
	dName = "data"
)

var (
	//ErrNilChan will be returned by Notify if it is passed a nil channel
	ErrNilChan = fmt.Errorf("nil channel given")
)

//Client is the default client used for requests.
var Client = &http.Client{}

func liveReq(verb, uri string, body io.Reader) (*http.Request, error) {
	req, err := GetReq(verb, uri, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")

	return req, nil
}

//Event is a go representation of an http server-sent event
type Event struct {
	*jason.Object
}

//GetReq is a function to return a single request. It will be used by notify to
//get a request and can be replaces if additional configuration is desired on
//the request. The "Accept" header will necessarily be overwritten.
var GetReq = func(verb, uri string, body io.Reader) (*http.Request, error) {
	return http.NewRequest(verb, uri, body)
}

//Notify takes the uri of an SSE stream and channel, and will send an Event
//down the channel when received, until the stream is closed.
// This is blocking, and so you will likely want to call this
// in a new goroutine (via `go sse.Notify(..)`)
func Notify(uri string, evCh chan<- *Event) error {
	fmt.Println("Notify method called!")

	if evCh == nil {
		return ErrNilChan
	}

	req, err := liveReq("GET", uri, nil)
	if err != nil {
		return fmt.Errorf("error getting sse request: %v", err)
	}

	res, err := Client.Do(req)
	if err != nil {
		return fmt.Errorf("error performing request for %s: %v", uri, err)
	}

	br := bufio.NewReader(res.Body)
	defer res.Body.Close()

	var currEvent *Event

	for {
		bs, err := br.ReadBytes('\n')
		if err != nil && err != io.EOF {
			return err
		}

		bsJSON, err := jason.NewObjectFromBytes(bs)
		if err != nil && err != io.EOF {
			return err
		}

		currEvent = &Event{bsJSON}

		evCh <- currEvent

	}

}

