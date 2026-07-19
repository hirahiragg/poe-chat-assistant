package ui

import (
	"fmt"
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/gesture"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/hirahiragg/poe-chat-assistant/internal/chat"
)

type msgState struct {
	translatedMsg  string
	translatedOut  string
	replyText      string
	translatingMsg bool
	translatingOut bool
}

func msgKey(msg *chat.Message) string {
	return fmt.Sprintf("%d|%d|%s|%s", msg.Timestamp.UnixNano(), msg.Channel, msg.Player, msg.Body)
}

type DetailPane struct {
	translateMsgBtn  widget.Clickable
	translateOutBtn  widget.Clickable
	copyBtn          widget.Clickable
	bgClick          gesture.Click
	replyEditor      widget.Editor
	selBody          widget.Selectable
	selTranslated    widget.Selectable
	selTranslatedOut widget.Selectable
	cache            map[string]*msgState
	currentKey       string
}

func NewDetailPane() *DetailPane {
	d := &DetailPane{
		cache: make(map[string]*msgState),
	}
	d.replyEditor.SingleLine = true
	d.replyEditor.Submit = true
	return d
}

func (d *DetailPane) TranslateMsgClicked(gtx layout.Context) bool {
	return d.translateMsgBtn.Clicked(gtx)
}

func (d *DetailPane) TranslateOutClicked(gtx layout.Context) bool {
	return d.translateOutBtn.Clicked(gtx)
}

func (d *DetailPane) CopyClicked(gtx layout.Context) bool {
	return d.copyBtn.Clicked(gtx)
}

func (d *DetailPane) ReplyText() string {
	return d.replyEditor.Text()
}

func (d *DetailPane) SetTranslatedMsg(s string) {
	st := d.current()
	st.translatedMsg = s
	st.translatingMsg = false
}

func (d *DetailPane) SetTranslatedOut(s string) {
	st := d.current()
	st.translatedOut = s
	st.translatingOut = false
}

func (d *DetailPane) SetTranslatingMsg(b bool) {
	st := d.current()
	st.translatingMsg = b
	if b {
		st.translatedMsg = ""
	}
}

func (d *DetailPane) SetTranslatingOut(b bool) {
	d.current().translatingOut = b
}

func (d *DetailPane) TranslatedOut() string {
	return d.current().translatedOut
}

func (d *DetailPane) current() *msgState {
	st, ok := d.cache[d.currentKey]
	if !ok {
		st = &msgState{}
		d.cache[d.currentKey] = st
	}
	return st
}

func (d *DetailPane) SwitchMessage(msg *chat.Message) {
	if d.currentKey != "" {
		if st, ok := d.cache[d.currentKey]; ok {
			st.replyText = d.replyEditor.Text()
		}
	}

	key := msgKey(msg)
	d.currentKey = key
	st := d.current()

	d.replyEditor.SetText(st.replyText)
	d.selBody = widget.Selectable{}
	d.selTranslated = widget.Selectable{}
	d.selTranslatedOut = widget.Selectable{}
}

func (d *DetailPane) Layout(gtx layout.Context, th *material.Theme, msg *chat.Message) layout.Dimensions {
	for {
		ev, ok := d.bgClick.Update(gtx.Source)
		if !ok {
			break
		}
		if ev.Kind == gesture.KindPress {
			d.selBody.SetCaret(0, 0)
			d.selTranslated.SetCaret(0, 0)
			d.selTranslatedOut.SetCaret(0, 0)
			gtx.Execute(key.FocusCmd{})
		}
	}

	size := gtx.Constraints.Max
	fillRect(gtx, colorSurface, size)

	_ = layout.Inset{
		Top: unit.Dp(12), Bottom: unit.Dp(12),
		Left: unit.Dp(16), Right: unit.Dp(16),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return d.layoutHeader(gtx, th, msg)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return d.layoutOriginal(gtx, th, msg)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return d.layoutMsgTranslateRow(gtx, th)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return d.layoutDivider(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return d.layoutReplySection(gtx, th)
			}),
		)
	})

	area := clip.Rect{Max: size}.Push(gtx.Ops)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	d.bgClick.Add(gtx.Ops)
	pass.Pop()
	area.Pop()

	return layout.Dimensions{Size: size}
}

func (d *DetailPane) layoutHeader(gtx layout.Context, th *material.Theme, msg *chat.Message) layout.Dimensions {
	return layout.Flex{Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th, unit.Sp(12), msg.Channel.Symbol())
					lbl.Color = channelColor(msg.Channel)
					lbl.Font.Weight = font.Bold
					return layout.Inset{Right: unit.Dp(8)}.Layout(gtx, lbl.Layout)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th, unit.Sp(15), msg.Player)
					lbl.Color = colorText
					lbl.Font.Weight = font.Bold
					return lbl.Layout(gtx)
				}),
			)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(th, unit.Sp(12), msg.Timestamp.Format("2006/01/02 15:04:05"))
			lbl.Color = colorTextDim
			return lbl.Layout(gtx)
		}),
	)
}

func (d *DetailPane) layoutOriginal(gtx layout.Context, th *material.Theme, msg *chat.Message) layout.Dimensions {
	d.selBody.SetText(msg.Body)
	lbl := material.Label(th, unit.Sp(14), msg.Body)
	lbl.Color = colorText
	lbl.State = &d.selBody
	return lbl.Layout(gtx)
}

func (d *DetailPane) layoutMsgTranslateRow(gtx layout.Context, th *material.Theme) layout.Dimensions {
	st := d.current()
	if st.translatedMsg != "" {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				d.selTranslated.SetText(st.translatedMsg)
				lbl := material.Label(th, unit.Sp(14), st.translatedMsg)
				lbl.Color = colorTranslated
				lbl.State = &d.selTranslated
				return lbl.Layout(gtx)
			}),
			layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				label := "Re-translate"
				if st.translatingMsg {
					label = "..."
				}
				return layout.Flex{Spacing: layout.SpaceStart}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layoutButton(gtx, th, &d.translateMsgBtn, label, colorBtnBg)
					}),
				)
			}),
		)
	}
	label := "Translate"
	if st.translatingMsg {
		label = "..."
	}
	return layout.Flex{Spacing: layout.SpaceStart}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutButton(gtx, th, &d.translateMsgBtn, label, colorBtnBg)
		}),
	)
}

func (d *DetailPane) layoutDivider(gtx layout.Context) layout.Dimensions {
	size := image.Point{X: gtx.Constraints.Max.X, Y: gtx.Dp(unit.Dp(1))}
	defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
	paint.ColorOp{Color: colorBorder}.Add(gtx.Ops)
	paint.PaintOp{}.Add(gtx.Ops)
	return layout.Dimensions{Size: size}
}

func (d *DetailPane) layoutReplySection(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(th, unit.Sp(12), "Reply")
			lbl.Color = colorTextDim
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return d.layoutEditorBox(gtx, th)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return d.layoutButtons(gtx, th)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			st := d.current()
			if st.translatedOut == "" {
				return layout.Dimensions{}
			}
			return layout.Inset{Top: unit.Dp(10)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(th, unit.Sp(12), "English")
						lbl.Color = colorTextDim
						return lbl.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						d.selTranslatedOut.SetText(st.translatedOut)
						lbl := material.Label(th, unit.Sp(14), st.translatedOut)
						lbl.Color = colorAccent
						lbl.Font.Weight = font.Medium
						lbl.State = &d.selTranslatedOut
						return lbl.Layout(gtx)
					}),
					layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return layout.Flex{Spacing: layout.SpaceStart}.Layout(gtx,
							layout.Rigid(func(gtx layout.Context) layout.Dimensions {
								return layoutButton(gtx, th, &d.copyBtn, "Copy", colorBtnBg)
							}),
						)
					}),
				)
			})
		}),
	)
}

func (d *DetailPane) layoutEditorBox(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Background{}.Layout(gtx,
		func(gtx layout.Context) layout.Dimensions {
			fillRect(gtx, colorCard, image.Point{X: gtx.Constraints.Max.X, Y: gtx.Constraints.Min.Y})
			return layout.Dimensions{Size: image.Point{X: gtx.Constraints.Max.X, Y: gtx.Constraints.Min.Y}}
		},
		func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top: unit.Dp(10), Bottom: unit.Dp(6),
				Left: unit.Dp(10), Right: unit.Dp(10),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				ed := material.Editor(th, &d.replyEditor, "Type in Japanese...")
				ed.Color = colorText
				ed.HintColor = colorTextDim
				return ed.Layout(gtx)
			})
		},
	)
}

func (d *DetailPane) layoutButtons(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{Spacing: layout.SpaceStart}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			label := "Translate"
			if d.current().translatingOut {
				label = "..."
			}
			return layoutButton(gtx, th, &d.translateOutBtn, label, colorBtnBg)
		}),
	)
}

func layoutButton(gtx layout.Context, th *material.Theme, btn *widget.Clickable, label string, bg color.NRGBA) layout.Dimensions {
	b := material.Button(th, btn, label)
	b.Background = bg
	b.Color = colorBtnText
	b.CornerRadius = unit.Dp(3)
	b.TextSize = unit.Sp(11)
	b.Font.Weight = font.Medium
	b.Inset = layout.Inset{
		Top: unit.Dp(6), Bottom: unit.Dp(2),
		Left: unit.Dp(10), Right: unit.Dp(10),
	}
	return b.Layout(gtx)
}
