/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package mempubsub

import (
	"context"
	"sync"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/trustbloc/edge-core/pkg/log"

	"github.com/trustbloc/orb/pkg/activitypub/service/lifecycle"
	service "github.com/trustbloc/orb/pkg/activitypub/service/spi"
)

var logger = log.New("activitypub_service")

const (
	defaultTimeout     = 10 * time.Second
	defaultConcurrency = 20
	defaultBufferSize  = 20
)

// Config holds the configuration for the publisher/subscriber.
type Config struct {
	// Timeout is the time that we should wait for an Ack or a Nack.
	Timeout time.Duration

	// Concurrency specifies the maximum number of concurrent requests.
	Concurrency int

	// BufferSize is the size of the Go channel buffer for a subscription.
	BufferSize int
}

// DefaultConfig returns the default configuration.
func DefaultConfig() Config {
	return Config{
		Timeout:     defaultTimeout,
		Concurrency: defaultConcurrency,
		BufferSize:  defaultBufferSize,
	}
}

// PubSub implements a publisher/subscriber using Go channels. This implementation
// works only on a single node, i.e. handlers are not distributed. In order to distribute
// the load across a cluster, a persistent message queue (such as RabbitMQ or Kafka) should
// instead be used.
type PubSub struct {
	*lifecycle.Lifecycle
	Config

	serviceName     string
	msgChansByTopic map[string][]chan *message.Message
	mutex           sync.RWMutex
	publishChan     chan *entry
	ackChan         chan *message.Message
	doneChan        chan struct{}
}

type entry struct {
	topic    string
	messages []*message.Message
}

// New returns a new publisher/subscriber.
func New(name string, cfg Config) *PubSub {
	m := &PubSub{
		Config:          cfg,
		serviceName:     name,
		msgChansByTopic: make(map[string][]chan *message.Message),
		publishChan:     make(chan *entry, cfg.BufferSize),
		ackChan:         make(chan *message.Message, cfg.Concurrency),
		doneChan:        make(chan struct{}),
	}

	m.Lifecycle = lifecycle.New("httpsubscriber-"+name, lifecycle.WithStop(m.stop))

	go m.processMessages()
	go m.processAcks()

	// Start the service immediately.
	m.Start()

	return m
}

// Close closes all resources.
func (p *PubSub) Close() error {
	p.Stop()

	return nil
}

func (p *PubSub) stop() {
	logger.Infof("[%s] Stopping publisher/subscriber...", p.serviceName)

	p.doneChan <- struct{}{}

	logger.Debugf("[%s] ... waiting for publisher to stop...", p.serviceName)

	<-p.doneChan

	logger.Debugf("[%s] ... closing subscriber channels...", p.serviceName)

	p.mutex.Lock()
	defer p.mutex.Unlock()

	for _, msgChans := range p.msgChansByTopic {
		for _, msgChan := range msgChans {
			close(msgChan)
		}
	}

	p.msgChansByTopic = nil

	close(p.ackChan)

	logger.Infof("[%s] ... publisher/subscriber stopped.", p.serviceName)
}

// Subscribe subscribes to a topic and returns the Go channel over which messages
// are sent. The returned channel will be closed when Close() is called on this struct.
func (p *PubSub) Subscribe(_ context.Context, topic string) (<-chan *message.Message, error) {
	if p.State() != service.StateStarted {
		return nil, service.ErrNotStarted
	}

	logger.Debugf("[%s] Subscribing to topic [%s]", p.serviceName, topic)

	p.mutex.Lock()
	defer p.mutex.Unlock()

	msgChan := make(chan *message.Message, p.BufferSize)

	p.msgChansByTopic[topic] = append(p.msgChansByTopic[topic], msgChan)

	return msgChan, nil
}

// Publish publishes the given messages to the given topic. This function returns
// immediately after sending the messages to the Go channel(s), although it will
// block if the concurrency limit (defined by Config.Concurrency) has been reached.
func (p *PubSub) Publish(topic string, messages ...*message.Message) error {
	if p.State() != service.StateStarted {
		return service.ErrNotStarted
	}

	p.publishChan <- &entry{
		topic:    topic,
		messages: messages,
	}

	return nil
}

func (p *PubSub) processMessages() {
	for {
		select {
		case entry := <-p.publishChan:
			p.publish(entry)

		case <-p.doneChan:
			p.doneChan <- struct{}{}

			logger.Debugf("[%s] ... publisher has stopped", p.serviceName)

			return
		}
	}
}

func (p *PubSub) processAcks() {
	for msg := range p.ackChan {
		go p.check(msg)
	}
}

func (p *PubSub) publish(entry *entry) {
	p.mutex.RLock()
	msgChans := p.msgChansByTopic[entry.topic]
	p.mutex.RUnlock()

	for _, msgChan := range msgChans {
		for _, m := range entry.messages {
			// Copy the message so that the Ack/Nack is specific to a subscriber
			msg := m.Copy()

			logger.Debugf("[%s] Publishing message [%s]", p.serviceName, msg.UUID)

			msgChan <- msg
			p.ackChan <- msg
		}
	}
}

func (p *PubSub) check(msg *message.Message) {
	logger.Debugf("[%s] Checking for Ack/Nack on message [%s]", p.serviceName, msg.UUID)

	select {
	case <-msg.Acked():
		logger.Infof("[%s] Message was successfully acknowledged [%s]", p.serviceName, msg.UUID)

	case <-msg.Nacked():
		logger.Infof("[%s] Message was not successfully acknowledged. Posting to undeliverable queue [%s]",
			p.serviceName, msg.UUID)

		p.postToUndeliverable(msg)

	case <-time.After(p.Timeout):
		logger.Warnf("[%s] Timed out after %s waiting for Ack/Nack. Posting to undeliverable queue [%s]",
			p.serviceName, p.Timeout, msg.UUID)

		p.postToUndeliverable(msg)
	}
}

func (p *PubSub) postToUndeliverable(msg *message.Message) {
	p.mutex.RLock()
	msgChans := p.msgChansByTopic[service.UndeliverableTopic]
	p.mutex.RUnlock()

	// When sending to the undeliverable queue, we don't want to block since this may result in a deadlock.
	// So if the undeliverable channel buffer is full, the send will fail and the message will be dropped.

	for _, msgChan := range msgChans {
		select {
		case msgChan <- msg:
			logger.Infof("[%s] Message was added to the undeliverable queue [%s]", p.serviceName, msg.UUID)

		default:
			logger.Warnf("[%s] Message could not be added to the undeliverable queue and will be dropped [%s]",
				p.serviceName, msg.UUID)
		}
	}
}
