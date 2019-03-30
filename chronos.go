package chronos

import (
	"sync"
	"time"
)

// Chronos c
type Chronos struct {
	entries []*Entry
	stop    chan struct{}
	running bool
}

// Entry e
type Entry struct {
	Period time.Duration
	Next   time.Time
	Prev   time.Time
	Job    interface{}
}

func newEntry(t time.Duration) *Entry {
	return &Entry{
		Period: t,
		Prev:   time.Unix(0, 0),
	}
}

func (c *Chronos) schedule(e *Entry) {
	e.Prev = time.Now()
	e.Next = e.Prev.Add(e.Period)
}

func (c *Chronos) runPending() {
	go func() {
		for _, entry := range c.entries {
			if time.Now().After(entry.Next) {
				go c.runJob(entry)
			}
		}
	}()
}

func (c *Chronos) runJob(e *Entry) {
	defer func() {
		if r := recover(); r != nil {
			c.schedule(e)
		}
	}()

	c.schedule(e)
	j := e.Job.(func())
	j()
}

// run r
func (c *Chronos) run(wg *sync.WaitGroup) {
	ticker := time.NewTicker(100 * time.Millisecond)

	for {
		select {
		case <-ticker.C:
			c.runPending()
			continue
		case <-c.stop:
			ticker.Stop()
			return
		}
	}
}

// New n
func New() *Chronos {
	return &Chronos{
		entries: nil,
		stop:    make(chan struct{}),
		running: false,
	}
}

// Every e
func (c *Chronos) Every(t time.Duration) *Entry {
	entry := newEntry(t)
	c.schedule(entry)

	if !c.running {
		c.entries = append(c.entries, entry)
	}
	return entry
}

// Do d
func (e *Entry) Do(job interface{}) {
	e.Job = job
}

// Start s
func (c *Chronos) Start() {
	var wg sync.WaitGroup

	wg.Add(len(c.entries))

	if c.running {
		return
	}

	c.running = true
	go c.run(&wg)
	wg.Wait()
}

// Stop s
func (c *Chronos) Stop() {
	if !c.running {
		return
	}

	c.stop <- struct{}{}
	c.running = false
}

/* ------------  Test -------------------*/

// func write() {
// 	fmt.Println("I Work!")
// }

// func main() {
// 	c := New()
// 	c.Every(5 * time.Second).Do(write)
// 	c.Start()
// }
