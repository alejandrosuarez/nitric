// Copyright 2021 Nitric Pty Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package worker

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/nitrictech/nitric/pkg/triggers"

	"github.com/google/uuid"
	pb "github.com/nitrictech/nitric/interfaces/nitric/v1"
	"github.com/valyala/fasthttp"
)

// RouteWorker - Worker representation for an http api route handler
type RouteWorker struct {
	methods []string
	path    string
	// gRPC Stream for this worker
	stream pb.FaasService_TriggerStreamServer
	// Response channels for this worker
	responseQueueLock sync.Mutex
	responseQueue     map[string]chan *pb.TriggerResponse
}

// newTicket - Generates a request/response ID and response channel
// for the requesting thread to wait on
func (s *RouteWorker) newTicket() (string, chan *pb.TriggerResponse) {
	s.responseQueueLock.Lock()
	defer s.responseQueueLock.Unlock()

	ID := uuid.New().String()
	responseChan := make(chan *pb.TriggerResponse)

	s.responseQueue[ID] = responseChan

	return ID, responseChan
}

// resolveTicket - Retrieves a response channel from the queue for
// the given ID and removes the entry from the map
func (s *RouteWorker) resolveTicket(ID string) (chan *pb.TriggerResponse, error) {
	s.responseQueueLock.Lock()
	defer func() {
		delete(s.responseQueue, ID)
		s.responseQueueLock.Unlock()
	}()

	if s.responseQueue[ID] == nil {
		return nil, fmt.Errorf("attempted to resolve ticket that does not exist!")
	}

	return s.responseQueue[ID], nil
}

// slashSplitter - used to split strings, with the same output regardless of leading or trailing slashes
// e.g - strings.FieldsFunc("/one/two/three/", f) == strings.FieldsFunc("/one/two/three", f) == strings.FieldsFunc("one/two/three", f) == ["one" "two" "three"]
func slashSplitter(c rune) bool {
	return c == '/'
}

func (s *RouteWorker) HandlesHttpRequest(trigger *triggers.HttpRequest) bool {

	for _, m := range s.methods {
		if strings.ToLower(m) != strings.ToLower(trigger.Method) {
			return false
		}
	}
	// Add path and method matching

	requestPathSegments := strings.FieldsFunc(trigger.Path, slashSplitter)
	pathSegments := strings.FieldsFunc(s.path, slashSplitter)

	// TODO: Filter for trailing/leading slashes
	if len(requestPathSegments) != len(pathSegments) {
		return false
	}

	for i, p := range pathSegments {
		if !strings.HasPrefix(p, ":") && p != requestPathSegments[i] {
			return false
		}
	}

	return true
}

func (s *RouteWorker) HandlesEvent(trigger *triggers.Event) bool {
	return false
}

func (s *RouteWorker) HandleHttpRequest(trigger *triggers.HttpRequest) (*triggers.HttpResponse, error) {
	// Generate an ID here
	ID, returnChan := s.newTicket()

	var mimeType string = ""
	if trigger.Header != nil && len(trigger.Header["Content-Type"]) > 0 {
		mimeType = trigger.Header["Content-Type"][0]
	}

	if mimeType == "" {
		mimeType = http.DetectContentType(trigger.Body)
	}

	headers := make(map[string]*pb.HeaderValue)
	headersOld := make(map[string]string)
	for k, v := range trigger.Header {
		if v != nil {
			headers[k] = &pb.HeaderValue{
				Value: v,
			}
			if len(v) > 0 {
				headersOld[k] = v[0]
			}
		}
	}

	query := make(map[string]*pb.QueryValue)
	queryOld := make(map[string]string)
	for k, v := range trigger.Query {
		if v != nil {
			query[k] = &pb.QueryValue{
				Value: v,
			}
			if len(v) > 0 {
				queryOld[k] = v[0]
			}
		}
	}

	triggerRequest := &pb.TriggerRequest{
		Data:     trigger.Body,
		MimeType: mimeType,
		Context: &pb.TriggerRequest_Http{
			Http: &pb.HttpTriggerContext{
				Path:           trigger.Path,
				Method:         trigger.Method,
				QueryParams:    query,
				QueryParamsOld: queryOld,
				Headers:        headers,
				HeadersOld:     headersOld,
			},
		},
	}

	// construct the message
	message := &pb.ServerMessage{
		Id: ID,
		Content: &pb.ServerMessage_TriggerRequest{
			TriggerRequest: triggerRequest,
		},
	}

	// send the message
	err := s.stream.Send(message)

	if err != nil {
		// There was an error enqueuing the message
		return nil, err
	}

	// wait for the response
	triggerResponse := <-returnChan

	httpResponse := triggerResponse.GetHttp()

	if httpResponse == nil {
		return nil, fmt.Errorf("fatal: Error handling event, incorrect response received from function")
	}

	fasthttpHeader := &fasthttp.ResponseHeader{}

	for key, val := range httpResponse.GetHeaders() {
		headerList := val.Value
		if key == "Set-Cookie" || key == "Cookie" {
			for _, v := range headerList {
				fasthttpHeader.Add(key, v)
			}
		} else if len(headerList) > 0 {
			fasthttpHeader.Set(key, headerList[0])
		}
	}

	response := &triggers.HttpResponse{
		Body: triggerResponse.Data,
		// No need to worry about integer truncation
		// as this should be a HTTP status code...
		StatusCode: int(httpResponse.Status),
		Header:     fasthttpHeader,
	}

	return response, nil
}

func (s *RouteWorker) HandleEvent(trigger *triggers.Event) error {
	return fmt.Errorf("route workers cannot handle events")
}

// listen
func (s *RouteWorker) Listen(errchan chan error) {
	// Listen for responses
	for {
		msg, err := s.stream.Recv()

		if err != nil {
			if err == io.EOF {
				// return will close stream from server side
				log.Println("exit")
			}
			if err != nil {
				log.Printf("received error %v", err)
			}

			errchan <- err
			break
		}

		if msg.GetInitRequest() != nil {
			fmt.Println("Received init request from worker")
			// FIXME: This appears to not work with the PHP runtime?
			//s.stream.Send(&pb.ServerMessage{
			//	Content: &pb.ServerMessage_InitResponse{
			//		InitResponse: &pb.InitResponse{},
			//	},
			//})
			continue
		}

		// Load the response channel and delete its map key reference
		if val, err := s.resolveTicket(msg.GetId()); err == nil {
			// For now assume this is a trigger response...
			response := msg.GetTriggerResponse()
			// Write the response the the waiting recipient
			val <- response
		} else {
			fmt.Println("Fatal: FaaS Worker in bad state closing stream: ", msg.GetId())
			errchan <- fmt.Errorf("Fatal: FaaS Worker in bad state closing stream! %v", msg.GetId())
			break
		}
	}
}

type RouteWorkerOptions struct {
	Path    string
	Methods []string
}

// Package private method
// Only a pool may create a new faas worker
func NewRouteWorker(stream pb.FaasService_TriggerStreamServer, opts *RouteWorkerOptions) *RouteWorker {
	return &RouteWorker{
		path:              opts.Path,
		methods:           opts.Methods,
		stream:            stream,
		responseQueueLock: sync.Mutex{},
		responseQueue:     make(map[string]chan *pb.TriggerResponse),
	}
}