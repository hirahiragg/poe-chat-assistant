package ui

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/hirahiragg/poe-chat-assistant/internal/chat"
)

type channelFilter struct {
	label   string
	channel chat.Channel
	btn     widget.Clickable
	enabled bool
}

type ChatList struct {
	list     widget.List
	clicks   []widget.Clickable
	selected int
	filters  []channelFilter
}

func NewChatList() *ChatList {
	cl := &ChatList{selected: -1}
	cl.list.Axis = layout.Vertical
	cl.filters = []channelFilter{
		{label: "#", channel: chat.ChannelGlobal, enabled: true},
		{label: "$", channel: chat.ChannelTrade, enabled: true},
		{label: "%", channel: chat.ChannelParty, enabled: true},
		{label: "@", channel: chat.ChannelWhisperIn, enabled: true},
	}
	return cl
}

func (cl *ChatList) Selected() int { return cl.selected }

func (cl *ChatList) isVisible(ch chat.Channel) bool {
	for _, f := range cl.filters {
		if f.channel == ch {
			return f.enabled
		}
		if f.channel == chat.ChannelWhisperIn && ch == chat.ChannelWhisperOut {
			return f.enabled
		}
	}
	return true
}

func (cl *ChatList) FilterMessages(messages []chat.Message) []chat.Message {
	filtered := make([]chat.Message, 0, len(messages))
	for _, m := range messages {
		if cl.isVisible(m.Channel) {
			filtered = append(filtered, m)
		}
	}
	return filtered
}

func (cl *ChatList) SelectPrev() {
	if cl.selected > 0 {
		cl.selected--
	}
}

func (cl *ChatList) SelectNext(max int) {
	if cl.selected < max-1 {
		cl.selected++
	}
}

func (cl *ChatList) Layout(gtx layout.Context, th *material.Theme, messages []chat.Message, footer func(layout.Context) layout.Dimensions) (layout.Dimensions, bool) {
	for len(cl.clicks) < len(messages) {
		cl.clicks = append(cl.clicks, widget.Clickable{})
	}

	changed := false

	for i := range cl.filters {
		if cl.filters[i].btn.Clicked(gtx) {
			cl.filters[i].enabled = !cl.filters[i].enabled
			cl.selected = -1
			changed = true
		}
	}

	for i := range messages {
		if cl.clicks[i].Clicked(gtx) {
			if cl.selected != i {
				cl.selected = i
				changed = true
			}
		}
	}

	total := len(messages)
	if footer != nil {
		total++
	}

	dims := layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return cl.layoutFilterBar(gtx, th)
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			if len(messages) == 0 && footer == nil {
				return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th, unit.Sp(13), "Waiting for chats...")
					lbl.Color = colorTextDim
					return lbl.Layout(gtx)
				})
			}
			return material.List(th, &cl.list).Layout(gtx, total, func(gtx layout.Context, i int) layout.Dimensions {
				if i < len(messages) {
					return cl.layoutItem(gtx, th, &messages[i], i)
				}
				return footer(gtx)
			})
		}),
	)
	return dims, changed
}

func (cl *ChatList) layoutFilterBar(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Background{}.Layout(gtx,
		func(gtx layout.Context) layout.Dimensions {
			size := image.Point{X: gtx.Constraints.Max.X, Y: gtx.Constraints.Min.Y}
			fillRect(gtx, colorBg, size)
			return layout.Dimensions{Size: size}
		},
		func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top: unit.Dp(4), Bottom: unit.Dp(4),
				Left: unit.Dp(8), Right: unit.Dp(8),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				children := make([]layout.FlexChild, 0, len(cl.filters)*2)
				for i := range cl.filters {
					i := i
					if i > 0 {
						children = append(children, layout.Rigid(layout.Spacer{Width: unit.Dp(4)}.Layout))
					}
					children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return cl.layoutFilterBtn(gtx, th, &cl.filters[i])
					}))
				}
				return layout.Flex{Alignment: layout.Middle}.Layout(gtx, children...)
			})
		},
	)
}

func (cl *ChatList) layoutFilterBtn(gtx layout.Context, th *material.Theme, f *channelFilter) layout.Dimensions {
	return material.Clickable(gtx, &f.btn, func(gtx layout.Context) layout.Dimensions {
		return layout.Background{}.Layout(gtx,
			func(gtx layout.Context) layout.Dimensions {
				bg := colorCard
				if f.enabled {
					bg = channelColor(f.channel)
				}
				rr := gtx.Dp(unit.Dp(3))
				defer clip.RRect{
					Rect: image.Rectangle{Max: gtx.Constraints.Min},
					NE:   rr, NW: rr, SE:  rr, SW:  rr,
				}.Push(gtx.Ops).Pop()
				paint.ColorOp{Color: bg}.Add(gtx.Ops)
				paint.PaintOp{}.Add(gtx.Ops)
				return layout.Dimensions{Size: gtx.Constraints.Min}
			},
			func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{
					Top: unit.Dp(3), Bottom: unit.Dp(3),
					Left: unit.Dp(8), Right: unit.Dp(8),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					c := colorBtnText
					if !f.enabled {
						c = colorTextDim
					}
					lbl := material.Label(th, unit.Sp(11), f.label)
					lbl.Color = c
					lbl.Font.Weight = font.Bold
					return lbl.Layout(gtx)
				})
			},
		)
	})
}

func (cl *ChatList) layoutItem(gtx layout.Context, th *material.Theme, msg *chat.Message, index int) layout.Dimensions {
	bg := colorSurface
	if index == cl.selected {
		bg = colorSelected
	}

	return material.Clickable(gtx, &cl.clicks[index], func(gtx layout.Context) layout.Dimensions {
		return layout.Background{}.Layout(gtx,
			func(gtx layout.Context) layout.Dimensions {
				defer clip.Rect{Max: gtx.Constraints.Min}.Push(gtx.Ops).Pop()
				paint.ColorOp{Color: bg}.Add(gtx.Ops)
				paint.PaintOp{}.Add(gtx.Ops)
				return layout.Dimensions{Size: gtx.Constraints.Min}
			},
			func(gtx layout.Context) layout.Dimensions {
				return layout.Inset{
					Top: unit.Dp(8), Bottom: unit.Dp(8),
					Left: unit.Dp(12), Right: unit.Dp(12),
				}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical, Spacing: layout.SpaceStart}.Layout(gtx,
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return layout.Flex{Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
								layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
									return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											lbl := material.Label(th, unit.Sp(11), msg.Channel.Symbol())
											lbl.Color = channelColor(msg.Channel)
											lbl.Font.Weight = font.Bold
											return layout.Inset{Right: unit.Dp(6)}.Layout(gtx, lbl.Layout)
										}),
										layout.Rigid(func(gtx layout.Context) layout.Dimensions {
											lbl := material.Label(th, unit.Sp(13), msg.Player)
											lbl.Color = colorText
											lbl.Font.Weight = font.Bold
											return lbl.Layout(gtx)
										}),
									)
								}),
								layout.Rigid(func(gtx layout.Context) layout.Dimensions {
									lbl := material.Label(th, unit.Sp(11), msg.Timestamp.Format("01/02 15:04"))
									lbl.Color = colorTextDim
									return lbl.Layout(gtx)
								}),
							)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(3)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							lbl := material.Label(th, unit.Sp(12), truncate(msg.Body, 60))
							lbl.Color = colorTextDim
							return lbl.Layout(gtx)
						}),
					)
				})
			},
		)
	})
}

func channelColor(ch chat.Channel) color.NRGBA {
	switch ch {
	case chat.ChannelWhisperIn, chat.ChannelWhisperOut:
		return colorWhisper
	case chat.ChannelTrade:
		return colorTrade
	case chat.ChannelParty:
		return colorParty
	default:
		return colorGlobal
	}
}

func truncate(s string, maxRunes int) string {
	runes := []rune(s)
	if len(runes) <= maxRunes {
		return s
	}
	return string(runes[:maxRunes-1]) + "…"
}

func fillRect(gtx layout.Context, col color.NRGBA, size image.Point) {
	defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
	paint.ColorOp{Color: col}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
}
