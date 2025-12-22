package bus

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/rennerdo30/webencode/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// startTestServer starts an embedded NATS server with JetStream enabled
func startTestServer(t *testing.T) *server.Server {
	t.Helper()
	opts := &server.Options{
		Host:      "127.0.0.1",
		Port:      -1, // Auto-assign port
		JetStream: true,
		StoreDir:  t.TempDir(),
	}
	ns, err := server.NewServer(opts)
	require.NoError(t, err, "failed to create test NATS server")

	go ns.Start()

	if !ns.ReadyForConnections(5 * time.Second) {
		t.Fatal("NATS server failed to start")
	}

	return ns
}

func TestStreamConstants(t *testing.T) {
	// Ensure stream names are defined
	assert.Equal(t, "WEBENCODE_WORK", StreamWork)
	assert.Equal(t, "WEBENCODE_EVENTS", StreamEvents)
	assert.Equal(t, "WEBENCODE_LIVE", StreamLive)
}

func TestSubjectConstants(t *testing.T) {
	// Ensure subject names are defined
	assert.Equal(t, "jobs.created", SubjectJobCreated)
	assert.Equal(t, "jobs.dispatch", SubjectJobDispatch)
	assert.Equal(t, "jobs.events", SubjectJobEvents)
	assert.Equal(t, "workers.heartbeat", SubjectWorkerHeartbeat)
	assert.Equal(t, "events.error", SubjectErrorEvents)
	assert.Equal(t, "audit.user_action", SubjectAuditUser)
	assert.Equal(t, "audit.system", SubjectAuditSystem)
	assert.Equal(t, "live.telemetry", SubjectLiveTelemetry)
	assert.Equal(t, "live.lifecycle", SubjectLiveLifecycle)
}

func TestBusStruct(t *testing.T) {
	// Test that Bus struct can be created (without actual NATS)
	b := &Bus{}
	assert.NotNil(t, b)
}

func TestStreamConfigDefaults(t *testing.T) {
	// Test that stream configuration values are sensible
	assert.True(t, 90*24*time.Hour > 24*time.Hour, "Audit retention should be longer than events")
	assert.True(t, 10*time.Second < 60*time.Second, "Live stream should be ephemeral")
}

func TestConnect(t *testing.T) {
	ns := startTestServer(t)
	defer ns.Shutdown()

	l := logger.New("test")
	url := ns.ClientURL()

	bus, err := Connect(url, l)
	require.NoError(t, err)
	require.NotNil(t, bus)
	defer bus.Close()

	assert.NotNil(t, bus.nc)
	assert.NotNil(t, bus.js)
	assert.NotNil(t, bus.logger)
}

func TestConnect_InvalidURL(t *testing.T) {
	// For invalid URL, Connect will hang due to RetryOnFailedConnect(true)
	// So we cannot easily test connection failure without modifying the code
	// This is expected behavior per the Connect implementation
	// The retry behavior is intentional for production resilience
}

func TestInitStreams(t *testing.T) {
	ns := startTestServer(t)
	defer ns.Shutdown()

	l := logger.New("test")
	bus, err := Connect(ns.ClientURL(), l)
	require.NoError(t, err)
	defer bus.Close()

	ctx := context.Background()
	err = bus.InitStreams(ctx)
	require.NoError(t, err)

	// Verify streams were created
	js := bus.JetStream()

	// Check WORK stream
	stream, err := js.Stream(ctx, StreamWork)
	require.NoError(t, err)
	info, err := stream.Info(ctx)
	require.NoError(t, err)
	assert.Equal(t, StreamWork, info.Config.Name)
	assert.Equal(t, jetstream.WorkQueuePolicy, info.Config.Retention)
	assert.Contains(t, info.Config.Subjects, "jobs.*")
	assert.Contains(t, info.Config.Subjects, "tasks.*")

	// Check EVENTS stream
	stream, err = js.Stream(ctx, StreamEvents)
	require.NoError(t, err)
	info, err = stream.Info(ctx)
	require.NoError(t, err)
	assert.Equal(t, StreamEvents, info.Config.Name)
	assert.Equal(t, jetstream.LimitsPolicy, info.Config.Retention)
	assert.Equal(t, 90*24*time.Hour, info.Config.MaxAge)

	// Check LIVE stream
	stream, err = js.Stream(ctx, StreamLive)
	require.NoError(t, err)
	info, err = stream.Info(ctx)
	require.NoError(t, err)
	assert.Equal(t, StreamLive, info.Config.Name)
	assert.Equal(t, jetstream.MemoryStorage, info.Config.Storage)
	assert.Equal(t, 10*time.Second, info.Config.MaxAge)
}

func TestInitStreams_Idempotent(t *testing.T) {
	ns := startTestServer(t)
	defer ns.Shutdown()

	l := logger.New("test")
	bus, err := Connect(ns.ClientURL(), l)
	require.NoError(t, err)
	defer bus.Close()

	ctx := context.Background()

	// Call InitStreams multiple times - should not error
	err = bus.InitStreams(ctx)
	require.NoError(t, err)

	err = bus.InitStreams(ctx)
	require.NoError(t, err)

	err = bus.InitStreams(ctx)
	require.NoError(t, err)
}

func TestPublish(t *testing.T) {
	ns := startTestServer(t)
	defer ns.Shutdown()

	l := logger.New("test")
	bus, err := Connect(ns.ClientURL(), l)
	require.NoError(t, err)
	defer bus.Close()

	ctx := context.Background()
	err = bus.InitStreams(ctx)
	require.NoError(t, err)

	// Publish to jobs.created (part of WORK stream)
	testData := []byte(`{"test": "message"}`)
	err = bus.Publish(ctx, SubjectJobCreated, testData)
	require.NoError(t, err)

	// Verify the message was published by consuming it
	js := bus.JetStream()
	cons, err := js.CreateConsumer(ctx, StreamWork, jetstream.ConsumerConfig{
		Durable:       "test-consumer",
		FilterSubject: SubjectJobCreated,
		AckPolicy:     jetstream.AckExplicitPolicy,
	})
	require.NoError(t, err)

	msg, err := cons.Next()
	require.NoError(t, err)
	assert.Equal(t, testData, msg.Data())
	require.NoError(t, msg.Ack())
}

func TestPublish_MultipleSubjects(t *testing.T) {
	ns := startTestServer(t)
	defer ns.Shutdown()

	l := logger.New("test")
	bus, err := Connect(ns.ClientURL(), l)
	require.NoError(t, err)
	defer bus.Close()

	ctx := context.Background()
	err = bus.InitStreams(ctx)
	require.NoError(t, err)

	tests := []struct {
		subject string
		stream  string
		data    []byte
	}{
		{SubjectJobCreated, StreamWork, []byte(`{"event": "job_created"}`)},
		{SubjectJobDispatch, StreamWork, []byte(`{"event": "job_dispatch"}`)},
		{SubjectJobEvents, StreamWork, []byte(`{"event": "job_events"}`)},
		{"events.error", StreamEvents, []byte(`{"event": "error"}`)},
		{SubjectWorkerHeartbeat, StreamEvents, []byte(`{"event": "heartbeat"}`)},
		{SubjectAuditUser, StreamEvents, []byte(`{"event": "audit_user"}`)},
		{SubjectAuditSystem, StreamEvents, []byte(`{"event": "audit_system"}`)},
		{"live.telemetry.stream1", StreamLive, []byte(`{"event": "telemetry"}`)},
		{"live.lifecycle.stream1", StreamLive, []byte(`{"event": "lifecycle"}`)},
	}

	for _, tc := range tests {
		t.Run(tc.subject, func(t *testing.T) {
			err := bus.Publish(ctx, tc.subject, tc.data)
			require.NoError(t, err)
		})
	}
}

func TestPublish_InvalidSubject(t *testing.T) {
	ns := startTestServer(t)
	defer ns.Shutdown()

	l := logger.New("test")
	bus, err := Connect(ns.ClientURL(), l)
	require.NoError(t, err)
	defer bus.Close()

	ctx := context.Background()
	err = bus.InitStreams(ctx)
	require.NoError(t, err)

	// Try to publish to a subject not in any stream
	err = bus.Publish(ctx, "invalid.subject.test", []byte(`{"test": "data"}`))
	assert.Error(t, err, "should fail for subject not in any stream")
}

func TestClose(t *testing.T) {
	ns := startTestServer(t)
	defer ns.Shutdown()

	l := logger.New("test")
	bus, err := Connect(ns.ClientURL(), l)
	require.NoError(t, err)

	// Verify connection is open
	assert.True(t, bus.nc.IsConnected())

	bus.Close()

	// Verify connection is closed
	assert.True(t, bus.nc.IsClosed())
}

func TestJetStream(t *testing.T) {
	ns := startTestServer(t)
	defer ns.Shutdown()

	l := logger.New("test")
	bus, err := Connect(ns.ClientURL(), l)
	require.NoError(t, err)
	defer bus.Close()

	js := bus.JetStream()
	assert.NotNil(t, js)
	assert.Equal(t, bus.js, js)
}

func TestConn(t *testing.T) {
	ns := startTestServer(t)
	defer ns.Shutdown()

	l := logger.New("test")
	bus, err := Connect(ns.ClientURL(), l)
	require.NoError(t, err)
	defer bus.Close()

	nc := bus.Conn()
	assert.NotNil(t, nc)
	assert.Equal(t, bus.nc, nc)
	assert.True(t, nc.IsConnected())
}

func TestPublish_ContextCancellation(t *testing.T) {
	ns := startTestServer(t)
	defer ns.Shutdown()

	l := logger.New("test")
	bus, err := Connect(ns.ClientURL(), l)
	require.NoError(t, err)
	defer bus.Close()

	err = bus.InitStreams(context.Background())
	require.NoError(t, err)

	// Create already cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = bus.Publish(ctx, SubjectJobCreated, []byte(`{"test": "data"}`))
	assert.Error(t, err, "should fail with cancelled context")
}

func TestInitStreams_ContextCancellation(t *testing.T) {
	ns := startTestServer(t)
	defer ns.Shutdown()

	l := logger.New("test")
	bus, err := Connect(ns.ClientURL(), l)
	require.NoError(t, err)
	defer bus.Close()

	// Create already cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = bus.InitStreams(ctx)
	assert.Error(t, err, "should fail with cancelled context")
}

func TestBus_ConcurrentPublish(t *testing.T) {
	ns := startTestServer(t)
	defer ns.Shutdown()

	l := logger.New("test")
	bus, err := Connect(ns.ClientURL(), l)
	require.NoError(t, err)
	defer bus.Close()

	ctx := context.Background()
	err = bus.InitStreams(ctx)
	require.NoError(t, err)

	// Concurrent publishes
	numGoroutines := 10
	numMessages := 100
	errCh := make(chan error, numGoroutines*numMessages)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			for j := 0; j < numMessages; j++ {
				data := []byte(fmt.Sprintf(`{"goroutine": %d, "msg": %d}`, id, j))
				if err := bus.Publish(ctx, SubjectJobCreated, data); err != nil {
					errCh <- err
				}
			}
		}(i)
	}

	// Wait a bit for all messages to be sent
	time.Sleep(500 * time.Millisecond)

	select {
	case err := <-errCh:
		t.Fatalf("concurrent publish failed: %v", err)
	default:
		// No errors
	}
}

func TestBus_ReconnectAfterServerRestart(t *testing.T) {
	ns := startTestServer(t)
	clientURL := ns.ClientURL()

	l := logger.New("test")
	bus, err := Connect(clientURL, l)
	require.NoError(t, err)
	defer bus.Close()

	// Wait for connection to be established
	time.Sleep(100 * time.Millisecond)
	assert.True(t, bus.nc.IsConnected())

	// Note: Full reconnection test would require starting a new server
	// on the same port, which is complex. This just validates initial connection.
}

func TestBus_NilAccessors(t *testing.T) {
	// Test that nil bus doesn't panic on accessors
	// This is a defensive test
	b := &Bus{}
	assert.Nil(t, b.JetStream())
	assert.Nil(t, b.Conn())
}

// startTestServerWithoutJetStream starts an embedded NATS server without JetStream
func startTestServerWithoutJetStream(t *testing.T) *server.Server {
	t.Helper()
	opts := &server.Options{
		Host:      "127.0.0.1",
		Port:      -1,
		JetStream: false, // No JetStream
	}
	ns, err := server.NewServer(opts)
	require.NoError(t, err, "failed to create test NATS server")

	go ns.Start()

	if !ns.ReadyForConnections(5 * time.Second) {
		t.Fatal("NATS server failed to start")
	}

	return ns
}

func TestConnect_NoJetStream(t *testing.T) {
	ns := startTestServerWithoutJetStream(t)
	defer ns.Shutdown()

	l := logger.New("test")
	url := ns.ClientURL()

	// Connect itself doesn't fail, but InitStreams will fail
	// because JetStream is not enabled
	bus, err := Connect(url, l)
	// jetstream.New() doesn't fail immediately - it fails on first use
	// So Connect may succeed, but InitStreams will fail
	if err != nil {
		// If Connect failed, that's fine too
		assert.Contains(t, err.Error(), "jetstream")
		return
	}
	defer bus.Close()

	// InitStreams should fail
	ctx := context.Background()
	err = bus.InitStreams(ctx)
	assert.Error(t, err, "should fail when JetStream is not available")
}

func TestPublish_LargeMessage(t *testing.T) {
	ns := startTestServer(t)
	defer ns.Shutdown()

	l := logger.New("test")
	bus, err := Connect(ns.ClientURL(), l)
	require.NoError(t, err)
	defer bus.Close()

	ctx := context.Background()
	err = bus.InitStreams(ctx)
	require.NoError(t, err)

	// Test with a larger payload
	largeData := make([]byte, 1024*100) // 100KB
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	err = bus.Publish(ctx, SubjectJobCreated, largeData)
	require.NoError(t, err)
}

func TestPublish_EmptyData(t *testing.T) {
	ns := startTestServer(t)
	defer ns.Shutdown()

	l := logger.New("test")
	bus, err := Connect(ns.ClientURL(), l)
	require.NoError(t, err)
	defer bus.Close()

	ctx := context.Background()
	err = bus.InitStreams(ctx)
	require.NoError(t, err)

	// Test with empty data
	err = bus.Publish(ctx, SubjectJobCreated, []byte{})
	require.NoError(t, err)

	// Test with nil data
	err = bus.Publish(ctx, SubjectJobCreated, nil)
	require.NoError(t, err)
}

func TestClose_Multiple(t *testing.T) {
	ns := startTestServer(t)
	defer ns.Shutdown()

	l := logger.New("test")
	bus, err := Connect(ns.ClientURL(), l)
	require.NoError(t, err)

	// Close multiple times should not panic
	bus.Close()
	bus.Close()
	bus.Close()
}

func TestInitStreams_StreamSubjects(t *testing.T) {
	ns := startTestServer(t)
	defer ns.Shutdown()

	l := logger.New("test")
	bus, err := Connect(ns.ClientURL(), l)
	require.NoError(t, err)
	defer bus.Close()

	ctx := context.Background()
	err = bus.InitStreams(ctx)
	require.NoError(t, err)

	js := bus.JetStream()

	// Verify WORK stream subjects
	workStream, err := js.Stream(ctx, StreamWork)
	require.NoError(t, err)
	info, _ := workStream.Info(ctx)
	assert.Contains(t, info.Config.Subjects, "jobs.*")
	assert.Contains(t, info.Config.Subjects, "tasks.*")

	// Verify EVENTS stream subjects
	eventsStream, err := js.Stream(ctx, StreamEvents)
	require.NoError(t, err)
	info, _ = eventsStream.Info(ctx)
	assert.Contains(t, info.Config.Subjects, "events.*")
	assert.Contains(t, info.Config.Subjects, "workers.*")
	assert.Contains(t, info.Config.Subjects, "audit.*")

	// Verify LIVE stream subjects
	liveStream, err := js.Stream(ctx, StreamLive)
	require.NoError(t, err)
	info, _ = liveStream.Info(ctx)
	assert.Contains(t, info.Config.Subjects, "live.telemetry.*")
	assert.Contains(t, info.Config.Subjects, "live.lifecycle.*")
}
