package queue

import (
	"sync"

	"github.com/pkg/errors"
)

func New(workers, max int) *Queue {
	q := &Queue{
		waiting: list{m: map[*item]struct{}{}},
		queue:   make(chan *item, max),
	}
	for i := 0; i < workers; i++ {
		go q.worker()
	}
	return q
}

type Queue struct {
	waiting list
	queue   chan *item
}

func (q *Queue) worker() {
	for i := range q.queue {

		// remove from the waiting list and broadcast to the others waiting that their position has
		// changed
		q.waiting.dequeue(i)

		// we can close the start channel to signal to the consumer that it should start processing.
		close(i.start)

		// wait for the consumer to close the end channel
		<-i.end
	}
}

// Slot requests a slot in the queue. The log function is called when the queue position changes. Execution
// of the work should not start until the start channel has been closed. The end channel should be closed
// when work is finished.
func (q *Queue) Slot(log func(int)) (start, end chan struct{}, err error) {
	start = make(chan struct{})
	end = make(chan struct{})

	i := &item{log: log, start: start, end: end}

	select {
	// Send the item to the workers. If no worker is available, it will join the queue. The channel
	// has n buffered spaces. If the buffer is full, we select the default and return an error.
	case q.queue <- i:
		// continue
	default:
		return nil, nil, TooManyItemsQueued
	}

	// add the item to the waiting list, and send it's position to the client
	if err := q.waiting.enqueue(i); err != nil {
		return nil, nil, err
	}

	return start, end, nil
}

type item struct {
	log        func(int)
	start, end chan struct{}
	position   int
}

type list struct {
	sync.RWMutex
	m map[*item]struct{}
}

var TooManyItemsQueued = errors.New("Sorry, too many items queued - try later.")

func (l list) enqueue(i *item) error {
	l.Lock()
	defer l.Unlock()
	l.m[i] = struct{}{}
	i.position = len(l.m)
	i.log(i.position)
	return nil
}

func (l list) dequeue(i *item) {
	l.Lock()
	defer l.Unlock()
	delete(l.m, i)
	for v := range l.m {
		// decrement the position of all the others waiting and fire a log for the new position
		v.position--
		v.log(v.position)
	}
}
