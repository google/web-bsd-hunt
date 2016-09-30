// Copyright 2016 The Web BSD Hunt Authors.
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
//
////////////////////////////////////////////////////////////////////////////////
//
// TODO: High-level file comment.
package apputils

import(
	"fmt"
	"io"
	"errors"
	"strings"
	"strconv"
	"net/url"
	"net/http"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine"
	"google.golang.org/appengine/log"
	"cloud.google.com/go/pubsub"
)

var ErrKeepaliveTimeout = errors.New("Keepalive timeout")

type KeepAliveMessage struct {
	Hostname	string
	Instance	string
	Seq		uint64
	URL		*url.URL
}

func (m *KeepAliveMessage) String() string {
	return fmt.Sprintf("KeepAliveMessage{Hostname: %s, Instance: %s, Seq: %d, URL: %s}", m.Hostname, m.Instance, m.Seq, m.URL.String())
}

type KeepAlive struct {
	ctx		context.Context
	client		*pubsub.Client
	subscription	*pubsub.Subscription
}

func NewKeepAlive(r *http.Request, topic string, timeout time.Duration) (*KeepAlive, error) {
	ctx, _ := context.WithTimeout(appengine.NewContext(r), timeout)

	appid := appengine.AppID(ctx)

	client, err := pubsub.NewClient(ctx, appid)
	if err != nil {
		return nil, err
	}

	var sub *pubsub.Subscription
	sub, err = client.CreateSubscription(ctx, "keepalive", client.Topic(topic), 0, nil)
	if err != nil {
		if !strings.Contains(err.Error(), "Resource already exists in the project") {
			return nil, err
		}
		sub = client.Subscription("keepalive")
	}

	keepalive := &KeepAlive {
		ctx:		ctx,
		client:		client,
		subscription:	sub,
	}

	return keepalive, nil
}

func ParseKeepAliveMessage(m *pubsub.Message) (*KeepAliveMessage, error) {
	var err error

	km := &KeepAliveMessage{}

	km.Hostname = m.Attributes["hostname"]
	km.Instance = m.Attributes["instance"]
	km.Seq, err = strconv.ParseUint(m.Attributes["seq"], 0, 64)
	if err != nil {
		return nil, err
	}

	ustr := string(m.Data)

	km.URL, err = url.Parse(ustr)
	if err != nil {
		return nil, err
	}

	return km, nil
}

//
// Fetch and process (waiting if necessary) messages for up to the timeout given  when the keepalive was created
//
func (keepalive *KeepAlive) Pull(msgHandler func(msg *KeepAliveMessage) error) error {
	log.Errorf(keepalive.ctx, "keepalive.Pull: start")

	it, err := keepalive.subscription.Pull(keepalive.ctx, pubsub.MaxPrefetch(1), pubsub.MaxExtension(10*time.Second))
	if err != nil {
		if err == context.DeadlineExceeded || strings.Contains(err.Error(), "deadline exceeded") {
			log.Errorf(keepalive.ctx, "keepalive.Pull: timed out")
			return nil
		}

		log.Errorf(keepalive.ctx, "keepalive.Pull: %v", err)
		return err
	}
	defer it.Stop()

	log.Errorf(keepalive.ctx, "keepalive.Pull: iterate")

	n := 0
	for n = 0; ; n++ {
		m, err := it.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			if err == context.DeadlineExceeded || strings.Contains(err.Error(), "Deadline exceeded") {
				// XXX(tad): strings.Contains is necessary because it's not returning the original error, it's returning an extended one
				log.Errorf(keepalive.ctx, "keepalive.Pull[%d]: %v", n, err)
				break
			}

			log.Errorf(keepalive.ctx, "keepalive.Pull[%d]: %v", n, err)
			return err
		}

		m.Done(true)

		km, err := ParseKeepAliveMessage(m)
		if err != nil {
			log.Errorf(keepalive.ctx, "keepalive.Pull[%d]: %v", n, err)
			return err
		}

		err = msgHandler(km)
		if err != nil {
			return err
		}
	}

	log.Errorf(keepalive.ctx, "keepalive.Pull: complete after processing %d messages\n", n)

	return nil
}

