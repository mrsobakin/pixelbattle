package mpmc_test

import (
	"github.com/mrsobakin/pixelbattle/internal/mpmc"
	"sync"
	"testing"
	"time"
)

type testConfig struct {
	timeout   time.Duration
	nProducer int
	nConsumer int
	nMsg      int
	buffer    uint64
}

func genericTestBase(cfg testConfig, t *testing.T) {
	tx := mpmc.NewMPMC[int](cfg.buffer)

	rx := make([]*mpmc.Consumer[int], cfg.nConsumer)
	slots := make([][]bool, cfg.nConsumer)
	for i := range cfg.nConsumer {
		rx[i] = tx.Subscribe()
		slots[i] = make([]bool, cfg.nProducer*cfg.nMsg)
	}

	wg := sync.WaitGroup{}
	lags := false

	wg.Add(cfg.nConsumer)

	for consumer := range cfg.nConsumer {
		go func() {
			defer wg.Done()
			for range cfg.nProducer * cfg.nMsg {
				msg, err := rx[consumer].Receive()
				if err != nil {
					lags = true
					return
				}
				slots[consumer][*msg] = true
			}
		}()
	}

	for producer := range cfg.nProducer {
		go func() {
			for i := range cfg.nMsg {
				tx.Send(producer*cfg.nMsg + i)
			}
		}()
	}

	sync := make(chan bool)

	go func() {
		wg.Wait()
		sync <- false
	}()
	go func() {
		time.Sleep(cfg.timeout)
		sync <- true
	}()

	timeout := <-sync

	if lags {
		t.Fatal("MPMC lags too much")
		return
	}

	if timeout {
		t.Fatal("MPMC timed out")
		return
	}

	for consumer := range cfg.nConsumer {
		for i := range cfg.nProducer * cfg.nMsg {
			if !slots[consumer][i] {
				t.Fatal("Skipped over a message")
				return
			}
		}
	}
}

func testMPSC(t *testing.T) {
	genericTestBase(testConfig{
		timeout:   20 * time.Second,
		nProducer: 1000,
		nConsumer: 1,
		nMsg:      10000,
		buffer:    10000000,
	}, t)
}

func testSPSC(t *testing.T) {
	genericTestBase(testConfig{
		timeout:   20 * time.Second,
		nProducer: 1,
		nConsumer: 1,
		nMsg:      10000000,
		buffer:    10000000,
	}, t)
}

func testSPMC(t *testing.T) {
	genericTestBase(testConfig{
		timeout:   20 * time.Second,
		nProducer: 1,
		nConsumer: 1000,
		nMsg:      10000,
		buffer:    10000000,
	}, t)
}

func testMPMC(t *testing.T) {
	genericTestBase(testConfig{
		timeout:   20 * time.Second,
		nProducer: 100,
		nConsumer: 100,
		nMsg:      1000,
		buffer:    10000000,
	}, t)
}

func Test(t *testing.T) {
	t.Run("spsc", testSPSC)
	t.Run("spmc", testSPMC)
	t.Run("mpsc", testMPSC)
	t.Run("mpmc", testMPMC)
}
