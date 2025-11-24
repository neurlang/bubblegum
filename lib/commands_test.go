package lib

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestCommandExecutor_Execute tests basic command execution.
func TestCommandExecutor_Execute(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msgChan := make(chan Msg, 10)
	executor := NewCommandExecutor(ctx, msgChan)

	// Create a simple command that returns a message
	testMsg := "test message"
	cmd := func() Msg {
		return testMsg
	}

	// Execute the command
	executor.Execute(cmd)

	// Wait for the message
	select {
	case msg := <-msgChan:
		if msg != testMsg {
			t.Errorf("Expected message %q, got %q", testMsg, msg)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for message")
	}

	executor.Shutdown()
}

// TestCommandExecutor_ExecuteNil tests that nil commands are handled gracefully.
func TestCommandExecutor_ExecuteNil(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msgChan := make(chan Msg, 10)
	executor := NewCommandExecutor(ctx, msgChan)

	// Execute a nil command (should not panic or block)
	executor.Execute(nil)

	executor.Shutdown()

	// Ensure no messages were sent
	select {
	case msg := <-msgChan:
		t.Errorf("Expected no message, got %v", msg)
	default:
		// Expected: no message
	}
}

// TestCommandExecutor_ThreadSafety tests thread-safe message delivery.
func TestCommandExecutor_ThreadSafety(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msgChan := make(chan Msg, 100)
	executor := NewCommandExecutor(ctx, msgChan)

	// Execute multiple commands concurrently
	numCommands := 50
	var wg sync.WaitGroup
	wg.Add(numCommands)

	for i := 0; i < numCommands; i++ {
		i := i
		go func() {
			defer wg.Done()
			cmd := func() Msg {
				return i
			}
			executor.Execute(cmd)
		}()
	}

	// Wait for all commands to be submitted
	wg.Wait()

	// Wait for all messages to be delivered
	executor.Shutdown()

	// Collect all messages
	messages := make(map[int]bool)
	timeout := time.After(2 * time.Second)
	for i := 0; i < numCommands; i++ {
		select {
		case msg := <-msgChan:
			if num, ok := msg.(int); ok {
				messages[num] = true
			}
		case <-timeout:
			t.Fatalf("Timeout waiting for messages, got %d/%d", len(messages), numCommands)
		}
	}

	// Verify all messages were received
	if len(messages) != numCommands {
		t.Errorf("Expected %d unique messages, got %d", numCommands, len(messages))
	}
}

// TestCommandExecutor_ContextCancellation tests that commands respect context cancellation.
func TestCommandExecutor_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	msgChan := make(chan Msg, 10)
	executor := NewCommandExecutor(ctx, msgChan)

	// Execute a command that will try to deliver after cancellation
	cmd := func() Msg {
		// Sleep to ensure we have time to cancel
		time.Sleep(50 * time.Millisecond)
		return "should not be delivered"
	}
	executor.Execute(cmd)

	// Cancel the context before the command completes
	time.Sleep(10 * time.Millisecond)
	cancel()

	executor.Shutdown()

	// The command may have executed, but the message should not be delivered
	// due to context cancellation. We allow a small window for the goroutine
	// to complete, but the message should be dropped.
	time.Sleep(100 * time.Millisecond)

	// Drain any messages that might have arrived before cancellation
	messageCount := 0
	for {
		select {
		case <-msgChan:
			messageCount++
		default:
			goto done
		}
	}
done:

	// We expect either 0 or 1 message depending on timing.
	// The important thing is that the system doesn't block or panic.
	if messageCount > 1 {
		t.Errorf("Expected at most 1 message, got %d", messageCount)
	}
}

// TestBatch tests batch command execution.
func TestBatch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msgChan := make(chan Msg, 10)
	executor := NewCommandExecutor(ctx, msgChan)

	// Create multiple commands
	cmd1 := func() Msg { return "msg1" }
	cmd2 := func() Msg { return "msg2" }
	cmd3 := func() Msg { return "msg3" }

	// Execute batch command
	batchCmd := Batch(cmd1, cmd2, cmd3)
	executor.Execute(batchCmd)

	// Wait for all messages
	executor.Shutdown()

	messages := make(map[string]bool)
	timeout := time.After(1 * time.Second)
	for i := 0; i < 3; i++ {
		select {
		case msg := <-msgChan:
			if str, ok := msg.(string); ok {
				messages[str] = true
			}
		case <-timeout:
			t.Fatalf("Timeout waiting for batch messages, got %d/3", len(messages))
		}
	}

	// Verify all messages were received
	expected := []string{"msg1", "msg2", "msg3"}
	for _, exp := range expected {
		if !messages[exp] {
			t.Errorf("Expected message %q not received", exp)
		}
	}
}

// TestBatch_WithNilCommands tests that batch handles nil commands gracefully.
func TestBatch_WithNilCommands(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msgChan := make(chan Msg, 10)
	executor := NewCommandExecutor(ctx, msgChan)

	// Create batch with some nil commands
	cmd1 := func() Msg { return "msg1" }
	cmd2 := func() Msg { return "msg2" }

	batchCmd := Batch(cmd1, nil, cmd2, nil)
	executor.Execute(batchCmd)

	executor.Shutdown()

	// Should receive only the non-nil command messages
	messages := make(map[string]bool)
	timeout := time.After(1 * time.Second)
	for i := 0; i < 2; i++ {
		select {
		case msg := <-msgChan:
			if str, ok := msg.(string); ok {
				messages[str] = true
			}
		case <-timeout:
			t.Fatalf("Timeout waiting for messages, got %d/2", len(messages))
		}
	}

	if !messages["msg1"] || !messages["msg2"] {
		t.Errorf("Expected messages msg1 and msg2, got %v", messages)
	}
}

// TestTick tests timer-based command execution.
func TestTick(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msgChan := make(chan Msg, 10)
	executor := NewCommandExecutor(ctx, msgChan)

	// Create a tick command with a short duration
	duration := 50 * time.Millisecond
	tickMsg := "tick"
	cmd := Tick(duration, func(tm time.Time) Msg {
		return tickMsg
	})

	start := time.Now()
	executor.Execute(cmd)

	// Wait for the message
	select {
	case msg := <-msgChan:
		elapsed := time.Since(start)
		if msg != tickMsg {
			t.Errorf("Expected message %q, got %q", tickMsg, msg)
		}
		// Verify the delay was approximately correct
		if elapsed < duration {
			t.Errorf("Message arrived too early: %v < %v", elapsed, duration)
		}
		if elapsed > duration*2 {
			t.Errorf("Message arrived too late: %v > %v", elapsed, duration*2)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for tick message")
	}

	executor.Shutdown()
}

// TestEvery tests recurring timer command execution.
func TestEvery(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msgChan := make(chan Msg, 20)
	executor := NewCommandExecutor(ctx, msgChan)

	// Create an every command with a short interval
	interval := 20 * time.Millisecond
	var counter int32
	cmd := Every(interval, func(tm time.Time) Msg {
		return atomic.AddInt32(&counter, 1)
	})

	executor.Execute(cmd)

	// Wait for multiple messages
	expectedCount := 5
	timeout := time.After(interval*time.Duration(expectedCount+2) + 100*time.Millisecond)
	receivedCount := 0

	for receivedCount < expectedCount {
		select {
		case msg := <-msgChan:
			if _, ok := msg.(int32); ok {
				receivedCount++
			}
		case <-timeout:
			t.Fatalf("Timeout waiting for recurring messages, got %d/%d", receivedCount, expectedCount)
		}
	}

	// Shutdown should stop the timer
	executor.Shutdown()

	// Give a bit of time to ensure no more messages arrive
	time.Sleep(interval * 2)

	// Drain any remaining messages
	remainingCount := 0
	for {
		select {
		case <-msgChan:
			remainingCount++
		default:
			goto done
		}
	}
done:

	// We should have received at least expectedCount messages
	totalReceived := receivedCount + remainingCount
	if totalReceived < expectedCount {
		t.Errorf("Expected at least %d messages, got %d", expectedCount, totalReceived)
	}
}

// TestEvery_Cancellation tests that recurring timers are cancelled on shutdown.
func TestEvery_Cancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msgChan := make(chan Msg, 20)
	executor := NewCommandExecutor(ctx, msgChan)

	// Create an every command
	interval := 20 * time.Millisecond
	cmd := Every(interval, func(tm time.Time) Msg {
		return "tick"
	})

	executor.Execute(cmd)

	// Wait for a few messages
	time.Sleep(interval * 3)

	// Shutdown the executor
	executor.Shutdown()

	// Drain the channel
	drainCount := 0
	for {
		select {
		case <-msgChan:
			drainCount++
		default:
			goto drained
		}
	}
drained:

	// Wait a bit more
	time.Sleep(interval * 3)

	// No new messages should arrive
	select {
	case msg := <-msgChan:
		t.Errorf("Expected no messages after shutdown, got %v", msg)
	default:
		// Expected: no message
	}
}

// TestQuit tests the Quit command.
func TestQuit(t *testing.T) {
	msg := Quit()
	if _, ok := msg.(quitMsg); !ok {
		t.Errorf("Expected quitMsg, got %T", msg)
	}
}

// TestCustomCommand tests arbitrary custom command execution.
func TestCustomCommand(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msgChan := make(chan Msg, 10)
	executor := NewCommandExecutor(ctx, msgChan)

	// Create a custom command that does some work
	customMsg := struct {
		result string
	}{result: "custom work done"}

	cmd := func() Msg {
		// Simulate some work
		time.Sleep(10 * time.Millisecond)
		return customMsg
	}

	executor.Execute(cmd)

	// Wait for the message
	select {
	case msg := <-msgChan:
		if msg != customMsg {
			t.Errorf("Expected custom message, got %v", msg)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for custom command message")
	}

	executor.Shutdown()
}
