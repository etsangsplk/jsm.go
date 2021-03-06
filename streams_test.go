// Copyright 2020 The NATS Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package jsm_test

import (
	"testing"
	"time"

	"github.com/nats-io/jsm.go"
	"github.com/nats-io/jsm.go/api"
)

func TestNewStreamFromDefault(t *testing.T) {
	srv, nc := startJSServer(t)
	defer srv.Shutdown()
	defer nc.Flush()

	stream, err := jsm.NewStreamFromDefault("q1", jsm.DefaultWorkQueue, jsm.StreamConnection(jsm.WithConnection(nc)), jsm.MemoryStorage(), jsm.MaxAge(time.Hour))
	checkErr(t, err, "create failed")

	stream.Reset()
	if stream.Name() != "q1" {
		t.Fatalf("incorrect name %q", stream.Name())
	}

	if stream.Storage() != api.MemoryStorage {
		t.Fatalf("incorrect storage, expected memory got %s", stream.Storage())
	}

	if stream.MaxAge() != time.Hour {
		t.Fatalf("incorrect max age, expected 1 hour got %v", stream.MaxAge())
	}
}

func TestLoadOrNewStreamFromDefault(t *testing.T) {
	srv, nc := startJSServer(t)
	defer srv.Shutdown()
	defer nc.Flush()

	names, err := jsm.StreamNames()
	checkErr(t, err, "ls failed")
	if len(names) != 0 {
		t.Fatalf("expected no streams")
	}

	stream, err := jsm.LoadOrNewStreamFromDefault("q1", jsm.DefaultWorkQueue, jsm.MemoryStorage(), jsm.MaxAge(time.Hour))
	checkErr(t, err, "create failed")

	_, err = jsm.LoadOrNewStreamFromDefault("q1", stream.Configuration())
	checkErr(t, err, "load failed")

	names, err = jsm.StreamNames()
	checkErr(t, err, "ls failed")

	if len(names) != 1 || names[0] != "q1" {
		t.Fatalf("expected [q1] got %v", names)
	}
}

func TestNewStream(t *testing.T) {
	srv, nc := startJSServer(t)
	defer srv.Shutdown()
	defer nc.Flush()

	stream, err := jsm.NewStream("q1", jsm.FileStorage())
	checkErr(t, err, "create failed")

	if stream.Name() != "q1" {
		t.Fatalf("expected q1 got %s", stream.Name())
	}

	if stream.Storage() != api.FileStorage {
		t.Fatalf("expected file storage got %s", stream.Storage())
	}

	if stream.Retention() != api.LimitsPolicy {
		t.Fatalf("expected stream retention got %s", stream.Retention())
	}
}

func TestLoadOrNewStream(t *testing.T) {
	srv, nc := startJSServer(t)
	defer srv.Shutdown()
	defer nc.Flush()

	names, err := jsm.StreamNames()
	checkErr(t, err, "ls failed")
	if len(names) != 0 {
		t.Fatalf("expected no streams")
	}

	stream, err := jsm.LoadOrNewStream("q1", jsm.FileStorage())
	checkErr(t, err, "create failed")

	if stream.Name() != "q1" {
		t.Fatalf("expected q1 got %s", stream.Name())
	}

	if stream.Storage() != api.FileStorage {
		t.Fatalf("expected file storage got %s", stream.Storage())
	}

	if stream.Retention() != api.LimitsPolicy {
		t.Fatalf("expected stream retention got %s", stream.Retention())
	}

	stream, err = jsm.LoadOrNewStream("q1", jsm.FileStorage())
	checkErr(t, err, "create failed")

	if stream.Name() != "q1" {
		t.Fatalf("expected q1 got %s", stream.Name())
	}
}

func TestLoadStream(t *testing.T) {
	srv, nc := startJSServer(t)
	defer srv.Shutdown()
	defer nc.Flush()

	names, err := jsm.StreamNames()
	checkErr(t, err, "ls failed")
	if len(names) != 0 {
		t.Fatalf("expected no streams")
	}

	_, err = jsm.LoadStream("q1")
	if err == nil {
		t.Fatal("expected error")
	}

	_, err = jsm.NewStream("q1", jsm.FileStorage())
	checkErr(t, err, "create failed")

	stream, err := jsm.LoadStream("q1")
	checkErr(t, err, "load failed")

	if stream.Name() != "q1" {
		t.Fatalf("expected q1 got %s", stream.Name())
	}

	if stream.Storage() != api.FileStorage {
		t.Fatalf("expected file storage got %s", stream.Storage())
	}
}

func TestStream_Reset(t *testing.T) {
	srv, nc := startJSServer(t)
	defer srv.Shutdown()
	defer nc.Flush()

	orig, err := jsm.NewStream("q1", jsm.FileStorage())
	checkErr(t, err, "create failed")

	// make original inconsistent
	err = orig.Delete()
	checkErr(t, err, "delete failed")

	// create a new one in its place
	_, err = jsm.NewStream("q1", jsm.MemoryStorage())
	checkErr(t, err, "create failed")

	// reload
	err = orig.Reset()
	checkErr(t, err, "reset failed")

	// verify its good
	if orig.Storage() != api.MemoryStorage {
		t.Fatalf("expected memory storage got %s", orig.Storage())
	}
}

func TestStream_LoadConsumer(t *testing.T) {
	srv, nc := startJSServer(t)
	defer srv.Shutdown()
	defer nc.Flush()

	stream, err := jsm.NewStream("q1", jsm.FileStorage())
	checkErr(t, err, "create failed")

	_, err = stream.NewConsumer(jsm.DurableName("c1"))
	checkErr(t, err, "create failed")

	consumer, err := stream.LoadConsumer("c1")
	checkErr(t, err, "load failed")

	if consumer.Name() != "c1" {
		t.Fatalf("expected consumer c1 got %s", consumer.Name())
	}
}

func TestStream_NewConsumerFromDefault(t *testing.T) {
	srv, nc := startJSServer(t)
	defer srv.Shutdown()
	defer nc.Flush()

	stream, err := jsm.NewStream("q1", jsm.FileStorage())
	checkErr(t, err, "create failed")

	consumer, err := stream.NewConsumerFromDefault(jsm.SampledDefaultConsumer, jsm.DurableName("q1"))
	checkErr(t, err, "create failed")
	consumer.Reset()
	if !consumer.IsSampled() {
		t.Fatalf("expected a sampled consumer")
	}
}

func TestStream_LoadOrNewConsumerFromDefault(t *testing.T) {
	srv, nc := startJSServer(t)
	defer srv.Shutdown()
	defer nc.Flush()

	stream, err := jsm.NewStream("q1", jsm.FileStorage())
	checkErr(t, err, "create failed")

	_, err = stream.LoadOrNewConsumerFromDefault("q1", jsm.SampledDefaultConsumer, jsm.DurableName("q1"))
	checkErr(t, err, "create failed")

	consumer, err := stream.LoadOrNewConsumerFromDefault("q1", jsm.SampledDefaultConsumer, jsm.DurableName("q1"))
	checkErr(t, err, "load failed")
	if !consumer.IsSampled() {
		t.Fatalf("expected a sampled consumer")
	}
}

func TestStream_UpdateConfiguration(t *testing.T) {
	srv, nc := startJSServer(t)
	defer srv.Shutdown()
	defer nc.Flush()

	stream, err := jsm.NewStream("q1", jsm.FileStorage(), jsm.Subjects("ORDERS.new"))
	checkErr(t, err, "create failed")

	if stream.Configuration().Subjects[0] != "ORDERS.new" {
		t.Fatalf("expected [ORDERS.new], got %v", stream.Configuration().Subjects)
	}

	cfg := stream.Configuration()
	cfg.Subjects = []string{"ORDERS.*"}

	err = stream.UpdateConfiguration(cfg)
	checkErr(t, err, "update failed")

	if len(stream.Configuration().Subjects) != 1 {
		t.Fatalf("expected [ORDERS.*], got %v", stream.Configuration().Subjects)
	}

	if stream.Configuration().Subjects[0] != "ORDERS.*" {
		t.Fatalf("expected [ORDERS.*], got %v", stream.Configuration().Subjects)
	}

	err = stream.UpdateConfiguration(stream.Configuration(), jsm.Subjects("ARCHIVE.*"))
	checkErr(t, err, "update failed")

	if len(stream.Configuration().Subjects) != 1 {
		t.Fatalf("expected [ARCHIVE.*], got %v", stream.Configuration().Subjects)
	}

	if stream.Configuration().Subjects[0] != "ARCHIVE.*" {
		t.Fatalf("expected [ARCHIVE.*], got %v", stream.Configuration().Subjects)
	}
}

func TestStream_ConsumerNames(t *testing.T) {
	srv, nc := startJSServer(t)
	defer srv.Shutdown()
	defer nc.Flush()

	stream, err := jsm.NewStream("q1", jsm.FileStorage())
	checkErr(t, err, "create failed")

	_, err = stream.NewConsumer(jsm.DurableName("c1"))
	checkErr(t, err, "create failed")

	_, err = stream.NewConsumer(jsm.DurableName("c2"))
	checkErr(t, err, "create failed")

	names, err := stream.ConsumerNames()
	checkErr(t, err, "names failed")

	if len(names) != 2 || names[0] != "c1" || names[1] != "c2" {
		t.Fatalf("expected [c1, c2] got %v", names)
	}
}

func TestStream_Information(t *testing.T) {
	srv, nc := startJSServer(t)
	defer srv.Shutdown()
	defer nc.Flush()

	stream, err := jsm.NewStream("q1", jsm.FileStorage())
	checkErr(t, err, "create failed")

	info, err := stream.Information()
	checkErr(t, err, "info failed")

	if info.Config.Name != "q1" {
		t.Fatalf("expected q1 got %s", info.Config.Name)
	}
}

func TestStream_Statistics(t *testing.T) {
	srv, nc := startJSServer(t)
	defer srv.Shutdown()
	defer nc.Flush()

	stream, err := jsm.NewStream("q1", jsm.FileStorage(), jsm.Subjects("test"))
	checkErr(t, err, "create failed")

	stats, err := stream.State()
	checkErr(t, err, "stats failed")
	if stats.Msgs != 0 {
		t.Fatalf("expected 0 messages got %d", stats.Msgs)
	}

	nc.Publish(stream.Subjects()[0], []byte("message 1"))

	stats, err = stream.State()
	checkErr(t, err, "stats failed")
	if stats.Msgs != 1 {
		t.Fatalf("expected 1 messages got %d", stats.Msgs)
	}
}

func TestStream_Delete(t *testing.T) {
	srv, nc := startJSServer(t)
	defer srv.Shutdown()
	defer nc.Flush()

	jsm.Trace()
	defer jsm.NoTrace()

	stream, err := jsm.NewStream("q1", jsm.FileStorage(), jsm.Subjects("test"))
	checkErr(t, err, "create failed")

	err = stream.Delete()
	checkErr(t, err, "delete failed")

	names, err := jsm.StreamNames()
	checkErr(t, err, "names failed")

	if len(names) != 0 {
		t.Fatalf("got streams where none should be: %v", names)
	}
}

func TestStream_Purge(t *testing.T) {
	srv, nc := startJSServer(t)
	defer srv.Shutdown()
	defer nc.Flush()

	stream, err := jsm.NewStream("q1", jsm.FileStorage(), jsm.Subjects("test"))
	checkErr(t, err, "create failed")

	nc.Publish(stream.Subjects()[0], []byte("message 1"))

	stats, err := stream.State()
	checkErr(t, err, "stats failed")
	if stats.Msgs != 1 {
		t.Fatalf("expected 1 messages got %d", stats.Msgs)
	}

	err = stream.Purge()
	checkErr(t, err, "purge failed")

	stats, err = stream.State()
	checkErr(t, err, "stats failed")
	if stats.Msgs != 0 {
		t.Fatalf("expected 0 messages got %d", stats.Msgs)
	}
}

func TestStream_ReadMessage(t *testing.T) {
	srv, nc := startJSServer(t)
	defer srv.Shutdown()
	defer nc.Flush()

	stream, err := jsm.NewStream("q1", jsm.FileStorage(), jsm.Subjects("test"))
	checkErr(t, err, "create failed")

	nc.Publish(stream.Subjects()[0], []byte("message 1"))

	msg, err := stream.ReadMessage(1)
	checkErr(t, err, "load failed")
	if string(msg.Data) != "message 1" {
		t.Fatalf("expected message 1 got %q", msg.Data)
	}

	_, err = stream.ReadMessage(2)
	if err == nil {
		t.Fatalf("ReadMessage didnt fail on unknown message")
	}
}

func TestStream_DeleteMessage(t *testing.T) {
	srv, nc := startJSServer(t)
	defer srv.Shutdown()
	defer nc.Flush()

	stream, err := jsm.NewStream("q1", jsm.FileStorage(), jsm.Subjects("test"))
	checkErr(t, err, "create failed")

	nc.Publish(stream.Subjects()[0], []byte("message 1"))

	_, err = stream.ReadMessage(1)
	checkErr(t, err, "load failed")

	err = stream.DeleteMessage(1)
	checkErr(t, err, "delete failed")

	msg, err := stream.ReadMessage(1)
	if err == nil {
		t.Fatalf("deleted messages was loaded %q", msg.Data)
	}
}

func testStreamConfig() *jsm.StreamConfig {
	return &jsm.StreamConfig{
		StreamConfig: api.StreamConfig{
			Subjects:     []string{},
			MaxAge:       -1,
			MaxBytes:     -1,
			MaxMsgSize:   -1,
			MaxMsgs:      -1,
			MaxConsumers: -1,
			Replicas:     -1,
		},
	}
}
func TestFileStorage(t *testing.T) {
	cfg := testStreamConfig()
	err := jsm.FileStorage()(cfg)
	checkErr(t, err, "failed")
	if cfg.Storage != api.FileStorage {
		t.Fatalf("expected FileStorage")
	}
}

func TestInterestRetention(t *testing.T) {
	cfg := testStreamConfig()
	err := jsm.InterestRetention()(cfg)
	checkErr(t, err, "failed")
	if cfg.Retention != api.InterestPolicy {
		t.Fatalf("expected InterestPolicy")
	}
}

func TestMaxAge(t *testing.T) {
	cfg := testStreamConfig()
	err := jsm.MaxAge(time.Hour)(cfg)
	checkErr(t, err, "failed")
	if cfg.MaxAge != time.Hour {
		t.Fatalf("expected 1 hour")
	}
}

func TestMaxBytes(t *testing.T) {
	cfg := testStreamConfig()
	err := jsm.MaxBytes(1024)(cfg)
	checkErr(t, err, "failed")
	if cfg.MaxBytes != 1024 {
		t.Fatalf("expected 1024")
	}
}

func TestMaxMessageSize(t *testing.T) {
	cfg := testStreamConfig()
	err := jsm.MaxMessageSize(1024)(cfg)
	checkErr(t, err, "failed")
	if cfg.MaxMsgSize != 1024 {
		t.Fatalf("expected 1024")
	}
}

func TestMaxMessages(t *testing.T) {
	cfg := testStreamConfig()
	err := jsm.MaxMessages(1024)(cfg)
	checkErr(t, err, "failed")
	if cfg.MaxMsgs != 1024 {
		t.Fatalf("expected 1024")
	}
}

func TestMaxObservables(t *testing.T) {
	cfg := testStreamConfig()
	err := jsm.MaxConsumers(1024)(cfg)
	checkErr(t, err, "failed")
	if cfg.MaxConsumers != 1024 {
		t.Fatalf("expected 1024")
	}
}

func TestMemoryStorage(t *testing.T) {
	cfg := testStreamConfig()
	err := jsm.MemoryStorage()(cfg)
	checkErr(t, err, "memory storage failed")
	if cfg.Storage != api.MemoryStorage {
		t.Fatalf("expected MemoryStorage")
	}
}

func TestDiscardNew(t *testing.T) {
	cfg := testStreamConfig()
	err := jsm.DiscardNew()(cfg)
	checkErr(t, err, "discard new failed")
	if cfg.Discard != api.DiscardNew {
		t.Fatal("expected DiscardNew")
	}
}

func TestDiscardOld(t *testing.T) {
	cfg := testStreamConfig()
	err := jsm.DiscardOld()(cfg)
	checkErr(t, err, "discard old failed")
	if cfg.Discard != api.DiscardOld {
		t.Fatal("expected DiscardOld")
	}
}

func TestNoAck(t *testing.T) {
	cfg := testStreamConfig()
	err := jsm.NoAck()(cfg)
	checkErr(t, err, "failed")
	if !cfg.NoAck {
		t.Fatalf("expected NoAck")
	}
}

func TestReplicas(t *testing.T) {
	cfg := testStreamConfig()
	err := jsm.Replicas(1024)(cfg)
	checkErr(t, err, "failed")
	if cfg.Replicas != 1024 {
		t.Fatalf("expected 1024")
	}
}

func TestLimitsRetention(t *testing.T) {
	cfg := testStreamConfig()
	err := jsm.LimitsRetention()(cfg)
	checkErr(t, err, "failed")
	if cfg.Retention != api.LimitsPolicy {
		t.Fatalf("expected LimitsPolicy")
	}
}

func TestSubjects(t *testing.T) {
	cfg := testStreamConfig()
	err := jsm.Subjects("one", "two")(cfg)
	checkErr(t, err, "failed")
	if len(cfg.Subjects) != 2 || cfg.Subjects[0] != "one" || cfg.Subjects[1] != "two" {
		t.Fatalf("expected [one, two] got %v", cfg.Subjects)
	}
}

func TestWorkQueueRetention(t *testing.T) {
	cfg := testStreamConfig()
	err := jsm.WorkQueueRetention()(cfg)
	checkErr(t, err, "failed")
	if cfg.Retention != api.WorkQueuePolicy {
		t.Fatalf("expected WorkQueuePolicy")
	}
}
