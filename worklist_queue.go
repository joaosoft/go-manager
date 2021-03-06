package manager

import (
	"encoding/json"
	"fmt"
	"github.com/joaosoft/logger"
	"github.com/labstack/gommon/log"
	"sync"
)

// Mode ...
type Mode int

const (
	// First In First Out
	FIFO Mode = iota
	// Last In Last Out
	LIFO
)

type Node struct {
	id       string
	data     interface{}
	next     *Node
	previous *Node
}

// Queue ...
type Queue struct {
	size    int
	start   *Node
	end     *Node
	mode    Mode
	maxSize int
	mux     *sync.Mutex
	ids     map[string]*Node
	logger logger.ILogger
}

// NewQueue ...
func (manager *Manager) NewQueue(options ...QueueOption) IList {
	queue := &Queue{
		ids: make(map[string]*Node),
		mux: &sync.Mutex{},
		logger: manager.logger,
	}
	queue.Reconfigure(options...)

	return queue
}

// Add ...
func (queue *Queue) Add(id string, data interface{}) error {
	queue.mux.Lock()
	defer queue.mux.Unlock()

	if queue.maxSize > 0 && queue.size >= queue.maxSize {
		return fmt.Errorf("the queue is full with [ size: %d ]", queue.size)
	}

	nodeToAdd := &Node{id: id, data: data}
	if queue.size == 0 {
		queue.start = nodeToAdd
		queue.end = nodeToAdd
	} else {
		nodeToAdd.next = queue.start
		queue.start.previous = nodeToAdd
		queue.start = nodeToAdd
	}
	queue.ids[id] = nodeToAdd
	queue.size++
	return nil
}

// Remove ...
func (queue *Queue) Remove(ids ...string) interface{} {
	queue.mux.Lock()
	defer queue.mux.Unlock()

	if queue.size == 0 {
		return nil
	}
	var nodeToRemove *Node
	if len(ids) == 0 {
		switch queue.mode {
		case FIFO:
			nodeToRemove = queue.end
			if queue.size > 1 {
				queue.end = nodeToRemove.previous
				queue.end.next = nil
			} else {
				queue.start = nil
				queue.end = nil
			}
			delete(queue.ids, nodeToRemove.id)
			queue.size--
			return nodeToRemove.data

		case LIFO:
			nodeToRemove = queue.start
			if queue.size > 1 {
				queue.start = queue.start.next
			} else {
				queue.start.next = nil
			}
			delete(queue.ids, nodeToRemove.id)
			queue.size--
			return nodeToRemove.data

		default:
			return nil
		}
	} else {
		var nodesRemoved []interface{}
		for _, id := range ids {
			nodeToRemove = queue.ids[id]
			nodeToRemove.previous.next = nodeToRemove.next
			delete(queue.ids, nodeToRemove.id)
			nodesRemoved = append(nodesRemoved, nodeToRemove.data)
			queue.size--
		}
		return nodesRemoved
	}

	return nil
}

// Size ...
func (queue *Queue) Size() int {
	queue.mux.Lock()
	defer queue.mux.Unlock()
	return queue.size
}

// IsEmpty ...
func (queue *Queue) IsEmpty() bool {
	queue.mux.Lock()
	defer queue.mux.Unlock()
	return queue.size == 0
}

// Dump ...
func (queue *Queue) Dump() string {
	type queuePrint struct {
		Size    int              `json:"size"`
		Mode    Mode             `json:"mode"`
		MaxSize int              `json:"max_size"`
		Ids     map[string]*Node `json:"ids"`
	}

	print := queuePrint{
		Size:    queue.size,
		Mode:    queue.mode,
		MaxSize: queue.maxSize,
		Ids:     queue.ids,
	}

	if json, err := json.Marshal(print); err != nil {
		log.Error(err)
		return ""
	} else {
		return string(json)
	}
}
