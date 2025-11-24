package lib

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// TestCommandExecutor_Integration tests the complete command execution flow.
func TestCommandExecutor_Integration(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msgChan := make(chan Msg, 50)
	executor := NewCommandExecutor(ctx, msgChan)

	// Track received messages
	var simpleCount, batchCount, tickCount, everyCount, customCount int32

	// 1. Simple command
	simpleCmd := func() Msg {
		atomic.AddInt32(&simpleCount, 1)
		return "simple"
	}
	executor.Execute(simpleCmd)

	// 2. Batch command
	batchCmd := Batch(
		func() Msg {
			atomic.AddInt32(&batchCount, 1)
			return "batch1"
		},
		func() Msg {
			atomic.AddInt32(&batchCount, 1)
			return "batch2"
		},
		func() Msg {
			atomic.AddInt32(&batchCount, 1)
			return "batch3"
		},
	)
	executor.Execute(batchCmd)

	// 3. Tick command
	tickCmd := Tick(20*time.Millisecond, func(t time.Time) Msg {
		atomic.AddInt32(&tickCount, 1)
		return "tick"
	})
	executor.Execute(tickCmd)

	// 4. Every command (recurring)
	everyCmd := Every(15*time.Millisecond, func(t time.Time) Msg {
		atomic.AddInt32(&everyCount, 1)
		return "every"
	})
	executor.Execute(everyCmd)

	// 5. Custom command with work
	customCmd := func() Msg {
		time.Sleep(10 * time.Millisecond)
		atomic.AddInt32(&customCount, 1)
		return struct{ data string }{"custom"}
	}
	executor.Execute(customCmd)

	// Wait for messages to be processed
	time.Sleep(100 * time.Millisecond)

	// Shutdown the executor
	executor.Shutdown()

	// Verify all commands executed
	if atomic.LoadInt32(&simpleCount) != 1 {
		t.Errorf("Expected 1 simple command execution, got %d", simpleCount)
	}

	if atomic.LoadInt32(&batchCount) != 3 {
		t.Errorf("Expected 3 batch command executions, got %d", batchCount)
	}

	if atomic.LoadInt32(&tickCount) != 1 {
		t.Errorf("Expected 1 tick command execution, got %d", tickCount)
	}

	everyCountVal := atomic.LoadInt32(&everyCount)
	if everyCountVal < 3 {
		t.Errorf("Expected at least 3 every command executions, got %d", everyCountVal)
	}

	if atomic.LoadInt32(&customCount) != 1 {
		t.Errorf("Expected 1 custom command execution, got %d", customCount)
	}

	// Verify messages were delivered
	messageCount := 0
	timeout := time.After(100 * time.Millisecond)
	for {
		select {
		case <-msgChan:
			messageCount++
		case <-timeout:
			goto done
		}
	}
done:

	// We should have received messages from all commands
	expectedMin := 1 + 3 + 1 + 3 + 1 // simple + batch + tick + every(min) + custom
	if messageCount < expectedMin {
		t.Errorf("Expected at least %d messages, got %d", expectedMin, messageCount)
	}
}

// TestCommandExecutor_QuitFlow tests the quit command flow.
func TestCommandExecutor_QuitFlow(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msgChan := make(chan Msg, 10)
	executor := NewCommandExecutor(ctx, msgChan)

	// Execute a command that returns Quit
	quitCmd := func() Msg {
		return Quit()
	}
	executor.Execute(quitCmd)

	// Wait for the message
	select {
	case msg := <-msgChan:
		if _, ok := msg.(quitMsg); !ok {
			t.Errorf("Expected quitMsg, got %T", msg)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for quit message")
	}

	executor.Shutdown()
}

// TestCommandExecutor_NestedBatch tests nested batch commands.
func TestCommandExecutor_NestedBatch(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msgChan := make(chan Msg, 20)
	executor := NewCommandExecutor(ctx, msgChan)

	// Create nested batch commands
	innerBatch := Batch(
		func() Msg { return "inner1" },
		func() Msg { return "inner2" },
	)

	outerBatch := Batch(
		func() Msg { return "outer1" },
		innerBatch,
		func() Msg { return "outer2" },
	)

	executor.Execute(outerBatch)
	executor.Shutdown()

	// Collect all messages
	messages := make(map[string]bool)
	timeout := time.After(1 * time.Second)
	for i := 0; i < 4; i++ {
		select {
		case msg := <-msgChan:
			if str, ok := msg.(string); ok {
				messages[str] = true
			}
		case <-timeout:
			t.Fatalf("Timeout waiting for messages, got %d/4", len(messages))
		}
	}

	// Verify all messages were received
	expected := []string{"outer1", "outer2", "inner1", "inner2"}
	for _, exp := range expected {
		if !messages[exp] {
			t.Errorf("Expected message %q not received", exp)
		}
	}
}

// TestCommandExecutor_ErrorHandling tests that commands can return error messages.
func TestCommandExecutor_ErrorHandling(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	msgChan := make(chan Msg, 10)
	executor := NewCommandExecutor(ctx, msgChan)

	// Create a command that returns an error message
	type errorMsg struct {
		err string
	}

	errorCmd := func() Msg {
		return errorMsg{err: "something went wrong"}
	}

	executor.Execute(errorCmd)

	// Wait for the error message
	select {
	case msg := <-msgChan:
		if errMsg, ok := msg.(errorMsg); !ok {
			t.Errorf("Expected errorMsg, got %T", msg)
		} else if errMsg.err != "something went wrong" {
			t.Errorf("Expected error message 'something went wrong', got %q", errMsg.err)
		}
	case <-time.After(1 * time.Second):
		t.Fatal("Timeout waiting for error message")
	}

	executor.Shutdown()
}
