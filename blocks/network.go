package blocks

import (
	"fmt"

	"github.com/kvap/i3barman/network"
)

type Network struct {
	name string
	long bool
	clicks chan string
	defaults BlockData
	manager *network.Manager
}

func (n *Network) Click(instance string) {
	n.clicks <- instance
}

func (n *Network) Name() string {
	return n.name
}

func (n *Network) render(conns []network.Connection) []BlockData {
	var data []BlockData

	for _, c := range conns {
		d := n.defaults
		d.Name = n.name
		switch c.Type {
		case "vpn":
			d.FullText = fmt.Sprintf("vpn:%s", c.Name)
			d.ShortText = "vpn"
			d.Urgent = !c.Good
		case "802-11-wireless":
			d.FullText = fmt.Sprintf("wifi:%s@%dMbps", c.Name, c.Rate / 1000000)
			d.ShortText = fmt.Sprintf("wifi:%s", c.Name)
			d.Urgent = !c.Good
		default:
			continue
		}
		data = append(data, d)
	}

	if len(data) == 0 {
		d := n.defaults
		d.Name = n.name
		d.FullText = "no network"
		d.Urgent = true
		data = append(data, d)
	}

	return data
}

func NewNetwork(defaults BlockData, name string, out chan<- BlockUpdate) *Network {
	n := &Network{
		name: name,
		clicks: make(chan string),
		defaults: defaults,
	}

	var err error
	n.manager, err = network.NewManager()
	if err != nil {
		return nil
	}

	go func() {
		update := func(conns []network.Connection) {
			out <- BlockUpdate{
				Name: n.name,
				Data: n.render(conns),
			}
		}

		updates, err := n.manager.WatchActiveConnections()
		if err != nil {
			return
		}

		conns, err := n.manager.GetActiveConnections()
		if err != nil {
			return
		}

		update(conns)
		for {
			select {
			case <-n.clicks:
				n.long = !n.long
				//update()
			case conns := <-updates:
				update(conns)
			}
		}
	}()
	return n
}
