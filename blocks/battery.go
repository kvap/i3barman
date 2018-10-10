package blocks

import (
	"fmt"
	"time"
	"strings"

	"github.com/kvap/i3barman/battery"
)

type Battery struct {
	name string
	long bool
	clicks chan string
	defaults BlockData
	manager *battery.Manager
	out chan<- BlockUpdate
}

func (b *Battery) Click(instance string) {
	b.clicks <- instance
}

func (b *Battery) Name() string {
	return b.name
}

func (b *Battery) update() {
	status, err := b.manager.GetStatus()
	if err != nil {
		return
	}

	d := b.defaults
	d.Name = b.name

	timeTo := "∞"
	if status.Charging {
		timeTo = status.TimeToFull.Truncate(time.Minute).String()
	} else if status.Discharging {
		timeTo = status.TimeToEmpty.Truncate(time.Minute).String()
	}
	timeTo = strings.TrimSuffix(timeTo, "0s")

	d.ShortText = fmt.Sprintf("%0.f%%", status.Percentage)
	if b.long {
		d.FullText = fmt.Sprintf("bat:%0.2f%% %s", status.Percentage, timeTo)
		if status.Charging {
			d.FullText += fmt.Sprintf(" ↑%0.1fW", status.Rate)
		} else if status.Discharging {
			d.FullText += fmt.Sprintf(" ↓%0.1fW", status.Rate)
		}
		d.Urgent = true
	} else {
		d.FullText = fmt.Sprintf("%0.f%% %s", status.Percentage, timeTo)
		d.Urgent = status.Percentage < 30
	}

	b.out <- BlockUpdate{
		Name: b.name,
		Data: []BlockData{d},
	}
}

func (b *Battery) loop() {
	ticker := time.NewTicker(1 * time.Second)

	b.update()
	for {
		select {
		case <-b.clicks:
			b.long = !b.long
			b.update()
		case <-ticker.C:
			b.update()
		}
	}
}

func NewBattery(defaults BlockData, name string, out chan<- BlockUpdate) *Battery {
	b := &Battery{
		name: name,
		clicks: make(chan string),
		defaults: defaults,
		out: out,
	}

	var err error
	b.manager, err = battery.NewManager()
	if err != nil {
		return nil
	}

	go b.loop()
	return b
}
