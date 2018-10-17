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

const (
	iconWireless = '\uf1eb' // 
	iconShield = '\uf3ed' // 
	iconLocked = '\uf023' // 
	iconUnlocked = '\uf3c1' // 
	iconMobile = '\uf10b' // 
	iconWired = '\uf6ff' // 
)

func pangoFA(icon rune) string {
	return fmt.Sprintf(`<span face="Font Awesome 5 Free">%c</span>`, icon)
}

func (n *Network) render(conns []network.Connection) []BlockData {
	var data []BlockData

	for _, c := range conns {
		d := n.defaults

		switch c.Type {
		case "vpn":
			if c.Good {
				d.FullText = pangoFA(iconLocked)
			} else {
				d.FullText = pangoFA(iconUnlocked)
			}
			if n.long {
				d.FullText += c.Name
			}
		case "802-3-ethernet":
			d.FullText = pangoFA(iconWired)
		case "802-11-wireless":
			d.FullText = pangoFA(iconWireless) + c.Name
			if n.long {
				d.FullText += fmt.Sprintf("@%dMbps", c.Rate / 1000000)
			}
		case "gsm":
			d.FullText = pangoFA(iconMobile)
		default:
			continue
		}

		d.Name = n.name
		d.Urgent = !c.Good
		if !n.long {
			d.Separator = false
			d.SeparatorBlockWidth = 0
		}
		d.Markup = "pango"

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
				update(conns)
			case conns := <-updates:
				update(conns)
			}
		}
	}()
	return n
}
