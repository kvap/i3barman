package blocks

type BlockData struct {
	// The full_text will be displayed by i3bar on the status line. This is
	// the only required key.
	FullText string `json:"full_text"`

	// Where appropriate, the short_text (string) entry should also be
	// provided. It will be used in case the status line needs to be
	// shortened because it uses more space than your screen provides. For
	// example, when displaying an IPv6 address, the prefix is usually (!)
	// more relevant than the suffix, because the latter stays constant
	// when using autoconf, while the prefix changes. When displaying the
	// date, the time is more important than the date (it is more likely
	// that you know which day it is than what time it is).
	ShortText string `json:"short_text,omitempty"`

	// To make the current state of the information easy to spot, colors
	// can be used. For example, the wireless block could be displayed in
	// red (using the color (string) entry) if the card is not associated
	// with any network and in green or yellow (depending on the signal
	// strength) when it is associated. Colors are specified in hex (like
	// in HTML), starting with a leading hash sign. For example, #ff0000
	// means red.
	Color string `json:"color,omitempty"`

	// Overrides the background color for this particular block.
	Background string `json:"background,omitempty"`

	// Overrides the border color for this particular block.
	Border string `json:"border,omitempty"`

	// The minimum width (in pixels) of the block. If the content of the
	// full_text key take less space than the specified min_width, the
	// block will be padded to the left and/or the right side, according to
	// the align key. This is useful when you want to prevent the whole
	// status line to shift when value take more or less space between each
	// iteration. The value can also be a string. In this case, the width
	// of the text given by min_width determines the minimum width of the
	// block. This is useful when you want to set a sensible minimum width
	// regardless of which font you are using, and at what particular size.
	MinWidth string `json:"min_width,omitempty"`

	// Align text on the center, right or left (default) of the block, when
	// the minimum width of the latter, specified by the min_width key, is
	// not reached.
	Align string `json:"align,omitempty"`

	// Every block should have a unique name (string) entry so that it can
	// be easily identified in scripts which process the output. i3bar
	// completely ignores the name and instance fields. Make sure to also
	// specify an instance (string) entry where appropriate. For example,
	// the user can have multiple disk space blocks for multiple mount
	// points.
	Name string `json:"name,omitempty"`
	Instance string `json:"instance,omitempty"`

	// A boolean which specifies whether the current value is urgent.
	// Examples are battery charge values below 1 percent or no more
	// available disk space (for non-root users). The presentation of
	// urgency is up to i3bar.
	Urgent bool `json:"urgent,omitempty"`

	// A boolean which specifies whether a separator line should be drawn
	// after this block. The default is true, meaning the separator line
	// will be drawn. Note that if you disable the separator line, there
	// will still be a gap after the block, unless you also use
	// separator_block_width.
	Separator bool `json:"separator"`

	// The amount of pixels to leave blank after the block. In the middle
	// of this gap, a separator line will be drawn unless separator is
	// disabled. Normally, you want to set this to an odd value (the
	// default is 9 pixels), since the separator line is drawn in the
	// middle.
	SeparatorBlockWidth int `json:"separator_block_width"`

	// A string that indicates how the text of the block should be parsed.
	// Set to "pango" to use Pango markup. Set to "none" to not use any
	// markup (default). Pango markup only works if you use a pango font.
	Markup string `json:"markup,omitempty"`
}

type Block interface {
	Click(instance string)
	Name() string
}

type BlockUpdate struct {
	Name string
	Data []BlockData
}
