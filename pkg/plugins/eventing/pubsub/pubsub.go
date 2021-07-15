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

package pubsub_service

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/nitric-dev/membrane/pkg/ifaces/pubsub"

	"cloud.google.com/go/pubsub"
	"github.com/nitric-dev/membrane/pkg/sdk"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
)

type PubsubEventService struct {
	sdk.UnimplementedEventingPlugin
	client ifaces_pubsub.PubsubClient
}

func (s *PubsubEventService) ListTopics() ([]string, error) {
	iter := s.client.Topics(context.TODO())

	var topics []string
	for {

		topic, err := iter.Next()
		if err == iterator.Done {
			break
		}

		if err != nil {
			return nil, fmt.Errorf("Error retrieving topics %v", err)
		}

		topics = append(topics, topic.ID())
	}

	return topics, nil
}

func (s *PubsubEventService) Publish(topic string, event *sdk.NitricEvent) error {
	ctx := context.TODO()

	eventBytes, err := json.Marshal(event)

	if err != nil {
		return fmt.Errorf("Payload marshalling error: %v", err)
	}

	pubsubTopic := s.client.Topic(topic)

	msg := ifaces_pubsub.AdaptPubsubMessage(&pubsub.Message{
		Attributes: map[string]string{
			"x-nitric-topic": topic,
		},
		Data: eventBytes,
	})

	if _, err := pubsubTopic.Publish(ctx, msg).Get(ctx); err != nil {
		return fmt.Errorf("Payload marshalling error: %v", err)
	}

	return nil
}

func New() (sdk.EventService, error) {
	ctx := context.Background()

	credentials, credentialsError := google.FindDefaultCredentials(ctx, pubsub.ScopeCloudPlatform)
	if credentialsError != nil {
		return nil, fmt.Errorf("GCP credentials error: %v", credentialsError)
	}

	client, clientError := pubsub.NewClient(ctx, credentials.ProjectID)
	if clientError != nil {
		return nil, fmt.Errorf("pubsub client error: %v", clientError)
	}

	return &PubsubEventService{
		client: ifaces_pubsub.AdaptPubsubClient(client),
	}, nil
}

func NewWithClient(client ifaces_pubsub.PubsubClient) (sdk.EventService, error) {
	return &PubsubEventService{
		client: client,
	}, nil
}