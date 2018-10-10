package main

import (
	"fmt"
	"encoding/json"
	"io"
	"os"

	"github.com/kvap/i3barman/blocks"
)

func SkipDelim(decoder *json.Decoder, expect rune) {
	token, err := decoder.Token()
	if err != nil {
		panic(err)
	}
	delim, ok := token.(json.Delim)
	if !ok || rune(delim) != expect {
		panic(fmt.Errorf("expected '%c'", expect))
	}
}

type ClickEvent struct {
	Name string `json:"name"`
	Instance string`json:"instance"`
	Button int `json:"button"`
	X int `json:"x"`
	Y int `json:"y"`
	RelativeX int `json:"relative_x"`
	RelativeY int `json:"relative_y"`
	Width int `json:"width"`
	Height int `json:"height"`
}


type Header struct {
	Version int `json:"version"`
	ClickEvents bool `json:"click_events"`
}

func ReadClickEvents(reader io.Reader, out chan<- ClickEvent) {
	decoder := json.NewDecoder(reader)
	SkipDelim(decoder, '[')
	for decoder.More() {
		var e ClickEvent
		if err := decoder.Decode(&e); err != nil {
			panic(err)
		}
		out <- e
	}
	SkipDelim(decoder, ']')
}

func WriteUpdates(writer io.Writer, updates <-chan []blocks.BlockData) {
	encoder := json.NewEncoder(writer)
	if _, err := io.WriteString(writer, "["); err != nil {
		panic(err)
	}

	needComma := false
	for u := range updates {
		if needComma {
			if _, err := io.WriteString(writer, ","); err != nil {
				panic(err)
			}
		} else {
			needComma = true
		}
		if err := encoder.Encode(u); err != nil {
			panic(err)
		}
	}

	if _, err := io.WriteString(writer, "]"); err != nil {
		panic(err)
	}
}

func WriteHeader(writer io.Writer) {
	encoded, err := json.Marshal(Header{Version: 1, ClickEvents: true})
	if err != nil {
		panic(err)
	}

	if _, err := writer.Write(encoded); err != nil {
		panic(err)
	}
}

type BlockSet struct {
	blocks []blocks.Block
	blockmap map[string]blocks.Block
	datamap map[string][]blocks.BlockData
}

func (bs *BlockSet) Click(event ClickEvent) {
	block, found := bs.blockmap[event.Name]
	if !found {
		return
	}

	block.Click(event.Instance)
}

func (bs *BlockSet) Update(u blocks.BlockUpdate) {
	bs.datamap[u.Name] = u.Data
}

func (bs *BlockSet) Render() []blocks.BlockData {
	var data []blocks.BlockData
	for _, b := range bs.blocks {
		data = append(data, bs.datamap[b.Name()]...)
	}
	return data
}

func (bs *BlockSet) Add(block blocks.Block) {
	bs.blocks = append(bs.blocks, block)
	bs.blockmap[block.Name()] = block
}

func NewBlockSet() *BlockSet {
	var bs BlockSet
	bs.blockmap = make(map[string]blocks.Block)
	bs.datamap = make(map[string][]blocks.BlockData)
	return &bs
}

func main() {
	reader := os.Stdin
	writer := os.Stdout

	WriteHeader(writer)

	events := make(chan ClickEvent)
	go ReadClickEvents(reader, events)

	wholeUpdates := make(chan []blocks.BlockData)
	go WriteUpdates(writer, wholeUpdates)

	bs := NewBlockSet()
	blockUpdates := make(chan blocks.BlockUpdate)

	defaults := blocks.BlockData{
		Separator: true,
		SeparatorBlockWidth: 9,
	}

	bs.Add(blocks.NewNetwork(defaults, "network", blockUpdates))
	bs.Add(blocks.NewBattery(defaults, "battery", blockUpdates))
	bs.Add(blocks.NewClock(defaults, "clock", blockUpdates))

	for {
		select {
		case e := <-events:
			bs.Click(e)
		case u := <-blockUpdates:
			bs.Update(u)
			wholeUpdates <- bs.Render()
		}
	}
}
