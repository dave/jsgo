package queue

import "sync"

func New(max int) *Queue {
	q := &Queue{
		waiting: list{m: map[*item]struct{}{}},
		queue:   make(chan *item),
	}
	for i := 0; i < max; i++ {
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
		// wait for the consumer to close the end channel
		<-i.end
	}
}

// Slot requests a slot in the queue. The log function is called when the queue position changes. Execution
// of the work should not start until the start channel has been closed. The end channel should be closed
// when work is finished.
func (q *Queue) Slot(log func(int)) (start, end chan struct{}) {
	start = make(chan struct{})
	end = make(chan struct{})

	i := &item{log: log, end: end}
	q.waiting.enqueue(i)

	go func() {
		q.queue <- i
		// as soon as the item is accepted by one of the workers, we can close the start channel to signal
		// to the consumer that it should start processing.
		close(start)
	}()

	return start, end
}

type item struct {
	log      func(int)
	end      chan struct{}
	position int
}

func (i *item) send(position int) {
	// log might be relatively long-running and we don't really care if it arrives out-of order, so we
	// run it in a goroutine. We don't want to lock the mutexes on the list while this sends data over
	// the network.
	go i.log(position)
}

type list struct {
	sync.RWMutex
	m map[*item]struct{}
}

func (l list) enqueue(i *item) {
	l.Lock()
	defer l.Unlock()
	l.m[i] = struct{}{}
	i.position = len(l.m)
	i.send(i.position)
}

func (l list) dequeue(i *item) {
	l.Lock()
	defer l.Unlock()
	delete(l.m, i)
	i.send(0)
	for v := range l.m {
		// decrement the position of all the others waiting and fire a log for the new position
		v.position--
		v.send(v.position)
	}
}
