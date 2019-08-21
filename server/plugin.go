package main

import (
	"bytes"
	"fmt"
	"sync"

	"github.com/mattermost/mattermost-server/plugin"
	"github.com/pkg/errors"
)

// Plugin implements the interface expected by the Mattermost server to communicate between the server and plugin processes.
type Plugin struct {
	plugin.MattermostPlugin
}

func (p *Plugin) OnActivate() error {
	err := p.API.KVSet("test1", []byte("value1"))
	if err != nil {
		return errors.Wrap(err, "failed to set key test1")
	}

	value1, err := p.API.KVGet("test1")
	if err != nil {
		return errors.Wrap(err, "failed to get key test1")
	}
	if bytes.Compare(value1, []byte("value1")) != 0 {
		return fmt.Errorf("unexpected value for key test1: %v", string(value1))
	}

	ok, err := p.API.KVCompareAndSet("test1", []byte("value1"), []byte("value2"))
	if err != nil {
		return errors.Wrap(err, "failed to compare and set key test1")
	}
	if !ok {
		return errors.New("should have compared and set key test1")
	}

	// KVCompareAndDelete will only successfully delete a value if the provided old value
	// matches whats in the database. Providing "value1" should fail, since the value is now
	// value2.
	ok, err = p.API.KVCompareAndDelete("test1", []byte("value1"))
	if err != nil {
		return errors.Wrap(err, "failed to compare and set key test1")
	}
	if ok {
		return errors.New("should not have deleted key test1 with old value value1")
	}

	// Verify the value is still in the database, since KVCompareAndDelete should have been
	// successful.
	value1, err = p.API.KVGet("test1")
	if err != nil {
		return errors.Wrap(err, "failed to get key test1")
	}
	if bytes.Compare(value1, []byte("value2")) != 0 {
		return fmt.Errorf("unexpected value for key test1: %v", string(value1))
	}

	// Actually delete the value, since we're providing the correct old value.
	ok, err = p.API.KVCompareAndDelete("test1", []byte("value2"))
	if err != nil {
		return errors.Wrap(err, "failed to compare and set key test1")
	}
	if !ok {
		return errors.New("should have deleted key test1 with old value value2")
	}

	// Verify that the value is actually deleted.
	value1, err = p.API.KVGet("test1")
	if err != nil {
		return errors.Wrap(err, "failed to get key test1")
	}
	if value1 != nil {
		return fmt.Errorf("deleted value was still present: %v", value1)
	}

	var wg sync.WaitGroup
	success := make(chan bool, 100)

	for i := 0; i < 100; i++ {
		threadId := i
		wg.Add(1)
		go func() {
			defer wg.Done()

			ok, err := p.API.KVCompareAndSet("test1", nil, []byte(fmt.Sprintf("value%d", threadId)))
			if err != nil {
				p.API.LogError("failed to compare and set key test1", "err", err)
			}

			success <- ok
		}()
	}

	wg.Wait()
	close(success)

	countSuccess := 0
	for result := range success {
		if result {
			countSuccess++
		}
	}

	if countSuccess != 1 {
		return fmt.Errorf("only one thread should have succeeded, got %d", countSuccess)
	}

	value1, err = p.API.KVGet("test1")
	if err != nil {
		return errors.Wrap(err, "failed to get key test1")
	}

	var wg2 sync.WaitGroup
	success2 := make(chan bool, 100)

	for i := 0; i < 100; i++ {
		wg2.Add(1)
		go func() {
			defer wg2.Done()

			ok, err := p.API.KVCompareAndDelete("test1", value1)
			if err != nil {
				p.API.LogError("failed to compare and delete key test1", "err", err)
			}

			success2 <- ok
		}()
	}

	wg2.Wait()
	close(success2)

	countSuccess2 := 0
	for result := range success2 {
		if result {
			countSuccess2++
		}
	}

	if countSuccess2 != 1 {
		return fmt.Errorf("only one thread should have succeeded to delete, got %d", countSuccess2)
	}

	p.API.LogInfo("Successfully validated kv store")

	return nil
}
