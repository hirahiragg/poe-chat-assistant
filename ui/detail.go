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

type DetailPane struct {
	replyEditor      widget.Editor
	translateMsgBtn  widget.Clickable
	translateOutBtn  widget.Clickable
	copyBtn          widget.Clickable
	translatedMsg    string
	translatedOut    string
	translatingMsg   bool
	translatingOut   bool
}

func NewDetailPane() *DetailPane {
	d := &DetailPane{}
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
	d.translatedMsg = s
	d.translatingMsg = false
}

func (d *DetailPane) SetTranslatedOut(s string) {
	d.translatedOut = s
	d.translatingOut = false
}

func (d *DetailPane) SetTranslatingMsg(b bool) {
	d.translatingMsg = b
}

func (d *DetailPane) SetTranslatingOut(b bool) {
	d.translatingOut = b
}

func (d *DetailPane) TranslatedOut() string {
	return d.translatedOut
}

func (d *DetailPane) ClearReply() {
	d.replyEditor.SetText("")
	d.translatedOut = ""
	d.translatingOut = false
}

func (d *DetailPane) ClearTranslation() {
	d.translatedMsg = ""
	d.translatingMsg = false
}

func (d *DetailPane) Layout(gtx layout.Context, th *material.Theme, msg *chat.Message) layout.Dimensions {
	return layout.Background{}.Layout(gtx,
		func(gtx layout.Context) layout.Dimensions {
			fillRect(gtx, colorSurface, gtx.Constraints.Max)
			return layout.Dimensions{Size: gtx.Constraints.Max}
		},
		func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
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
		},
	)
}

func (d *DetailPane) layoutHeader(gtx layout.Context, th *material.Theme, msg *chat.Message) layout.Dimensions {
	return layout.Flex{Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					lbl := material.Label(th, unit.Sp(12), msg.Channel.String())
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
	lbl := material.Label(th, unit.Sp(14), msg.Body)
	lbl.Color = colorText
	return lbl.Layout(gtx)
}

func (d *DetailPane) layoutMsgTranslateRow(gtx layout.Context, th *material.Theme) layout.Dimensions {
	if d.translatedMsg != "" {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(th, unit.Sp(14), d.translatedMsg)
				lbl.Color = colorTranslated
				return lbl.Layout(gtx)
			}),
		)
	}
	label := "Translate"
	if d.translatingMsg {
		label = "..."
	}
	return layout.Flex{}.Layout(gtx,
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
			if d.translatedOut == "" {
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
						lbl := material.Label(th, unit.Sp(14), d.translatedOut)
						lbl.Color = colorAccent
						lbl.Font.Weight = font.Medium
						return lbl.Layout(gtx)
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
				Top: unit.Dp(8), Bottom: unit.Dp(8),
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
			if d.translatingOut {
				label = "..."
			}
			return layoutButton(gtx, th, &d.translateOutBtn, label, colorBtnBg)
		}),
		layout.Rigid(layout.Spacer{Width: unit.Dp(8)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if d.translatedOut == "" {
				return layout.Dimensions{}
			}
			return layoutButton(gtx, th, &d.copyBtn, "Copy", colorAccent)
		}),
	)
}

func layoutButton(gtx layout.Context, th *material.Theme, btn *widget.Clickable, label string, bg color.NRGBA) layout.Dimensions {
	b := material.Button(th, btn, label)
	b.Background = bg
	b.Color = colorBtnText
	b.CornerRadius = unit.Dp(4)
	b.TextSize = unit.Sp(13)
	b.Font.Weight = font.Medium
	b.Inset = layout.Inset{
		Top: unit.Dp(9), Bottom: unit.Dp(3),
		Left: unit.Dp(16), Right: unit.Dp(16),
	}
	return b.Layout(gtx)
}
