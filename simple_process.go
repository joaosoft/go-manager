package manager

import "sync"

// SimpleProcess ...
type SimpleProcess struct {
	function func() error
	started  bool
}

// NewSimpleProcess...
func NewSimpleProcess(function func() error) IProcess {
	return &SimpleProcess{
		function: function,
	}
}

// Start ...
func (process *SimpleProcess) Start(wg *sync.WaitGroup) error {
	if wg == nil {
		wg = &sync.WaitGroup{}
		wg.Add(1)
	}

	defer wg.Done()

	if process.started {
		return nil
	}

	if err := process.function(); err != nil {
		return err
	}

	process.started = true

	return nil
}

// Stop ...
func (process *SimpleProcess) Stop(wg *sync.WaitGroup) error {
	if wg == nil {
		wg = &sync.WaitGroup{}
		wg.Add(1)
	}

	defer wg.Done()

	if !process.started {
		return nil
	}

	process.started = false

	return nil
}

// Started ...
func (process *SimpleProcess) Started() bool {
	return process.started
}
