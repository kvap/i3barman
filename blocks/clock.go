package blocks

import (
	"time"
)

type Clock struct {
	name string
	long bool
	clicks chan struct{}
}

func (c *Clock) Click(instance string) {
	c.clicks <- struct{}{}
}

func (c *Clock) Name() string {
	return c.name
}

func NewClock(defaults BlockData, name string, out chan<- BlockUpdate) *Clock {
	c := &Clock{
		name: name,
		clicks: make(chan struct{}),
	}

	go func() {
		ticker := time.NewTicker(1 * time.Second)

		update := func() {
			data := defaults
			data.Name = c.name
			data.Urgent = c.long
			t := time.Now()
			data.ShortText = t.Format("15:04")
			if c.long {
				data.FullText = t.Format("2006-01-02 Mon 15:04:05 MST")
			} else {
				data.FullText = data.ShortText
			}
			out <- BlockUpdate{
				Name: c.name,
				Data: []BlockData{data},
			}
		}

		update()
		for {
			select {
			case <-c.clicks:
				c.long = !c.long
				update()
			case <-ticker.C:
				update()
			}
		}
	}()

	return c
}
