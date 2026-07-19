package ui

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/gesture"
	"gioui.org/io/key"
	"gioui.org/io/pointer"
	"gioui.org/layout"
	"gioui.org/op/clip"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/hirahiragg/poe-chat-assistant/internal/config"
)

type SettingsPane struct {
	list             widget.List
	logPathEditor    widget.Editor
	deeplKeyEditor   widget.Editor
	geminiKeyEditor  widget.Editor
	targetLangEditor widget.Editor
	hotkeyEditor     widget.Editor

	translatorBtns [3]widget.Clickable
	translatorIdx  int

	targetLangBtns [4]widget.Clickable

	bgClick  gesture.Click
	saveBtn  widget.Clickable
	closeBtn widget.Clickable

	saved    bool
	onSave   func(*config.Config)
	onClose  func()
}

var translators = [3]string{"google", "deepl", "gemini"}
var languages = [4]string{"en", "ja", "ko", "zh"}

func NewSettingsPane(cfg *config.Config, onSave func(*config.Config), onClose func()) *SettingsPane {
	s := &SettingsPane{
		onSave:  onSave,
		onClose: onClose,
	}
	s.list.Axis = layout.Vertical
	s.logPathEditor.SingleLine = true
	s.logPathEditor.SetText(cfg.LogPath)
	s.deeplKeyEditor.SingleLine = true
	s.deeplKeyEditor.SetText(cfg.DeepLKey)
	s.geminiKeyEditor.SingleLine = true
	s.geminiKeyEditor.SetText(cfg.GeminiKey)
	s.targetLangEditor.SingleLine = true
	s.targetLangEditor.SetText(cfg.TargetLang)
	s.hotkeyEditor.SingleLine = true
	s.hotkeyEditor.SetText(cfg.Hotkey)

	for i, t := range translators {
		if t == cfg.Translator {
			s.translatorIdx = i
			break
		}
	}
	return s
}

func (s *SettingsPane) HandleActions(gtx layout.Context) {
	for {
		ev, ok := s.bgClick.Update(gtx.Source)
		if !ok {
			break
		}
		if ev.Kind == gesture.KindPress {
			gtx.Execute(key.FocusCmd{})
		}
	}
	for i := range s.translatorBtns {
		if s.translatorBtns[i].Clicked(gtx) {
			s.translatorIdx = i
			s.saved = false
			gtx.Execute(key.FocusCmd{})
		}
	}
	for i := range s.targetLangBtns {
		if s.targetLangBtns[i].Clicked(gtx) {
			s.targetLangEditor.SetText(languages[i])
			s.saved = false
			gtx.Execute(key.FocusCmd{})
		}
	}

	if s.saveBtn.Clicked(gtx) {
		cfg := &config.Config{
			LogPath:    s.logPathEditor.Text(),
			Translator: translators[s.translatorIdx],
			DeepLKey:   s.deeplKeyEditor.Text(),
			GeminiKey:  s.geminiKeyEditor.Text(),
			TargetLang: s.targetLangEditor.Text(),
			Hotkey:     s.hotkeyEditor.Text(),
		}
		s.onSave(cfg)
		s.saved = true
	}

	if s.closeBtn.Clicked(gtx) {
		s.onClose()
	}
}

func (s *SettingsPane) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	size := gtx.Constraints.Max
	fillRect(gtx, colorBg, size)

	_ = layout.Inset{
		Top: unit.Dp(16), Bottom: unit.Dp(16),
		Left: unit.Dp(24), Right: unit.Dp(24),
	}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
		return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
			layout.Rigid(func(gtx layout.Context) layout.Dimensions {
				return s.layoutHeader(gtx, th)
			}),
			layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
				return material.List(th, &s.list).Layout(gtx, 1, func(gtx layout.Context, _ int) layout.Dimensions {
					return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
						layout.Rigid(layout.Spacer{Height: unit.Dp(20)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return s.layoutField(gtx, th, "Client.txt Path", &s.logPathEditor, "C:\\path\\to\\Client.txt")
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return s.layoutTranslatorSelect(gtx, th)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return s.layoutField(gtx, th, "DeepL API Key", &s.deeplKeyEditor, "")
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(12)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return s.layoutField(gtx, th, "Gemini API Key", &s.geminiKeyEditor, "")
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return s.layoutLangRow(gtx, th)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(16)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return s.layoutHotkeyField(gtx, th)
						}),
						layout.Rigid(layout.Spacer{Height: unit.Dp(24)}.Layout),
						layout.Rigid(func(gtx layout.Context) layout.Dimensions {
							return s.layoutFooter(gtx, th)
						}),
					)
				})
			}),
		)
	})

	// PassOp overlay on top: detects clicks everywhere but passes them
	// through so editors and buttons below still work normally.
	// When an editor is clicked, its FocusCmd (issued during Layout)
	// overrides our unfocus FocusCmd (issued in HandleActions).
	area := clip.Rect{Max: size}.Push(gtx.Ops)
	pass := pointer.PassOp{}.Push(gtx.Ops)
	s.bgClick.Add(gtx.Ops)
	pass.Pop()
	area.Pop()

	return layout.Dimensions{Size: size}
}

func (s *SettingsPane) layoutHeader(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(th, unit.Sp(18), "Settings")
			lbl.Color = colorText
			lbl.Font.Weight = font.Bold
			return lbl.Layout(gtx)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutButton(gtx, th, &s.closeBtn, "Close", colorBorder)
		}),
	)
}

func (s *SettingsPane) layoutField(gtx layout.Context, th *material.Theme, label string, editor *widget.Editor, hint string) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(th, unit.Sp(12), label)
			lbl.Color = colorTextDim
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutEditorBox(gtx, th, editor, hint)
		}),
	)
}

func (s *SettingsPane) layoutEditorBox(gtx layout.Context, th *material.Theme, editor *widget.Editor, hint string) layout.Dimensions {
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
				ed := material.Editor(th, editor, hint)
				ed.Color = colorText
				ed.HintColor = colorTextDim
				return ed.Layout(gtx)
			})
		},
	)
}

func (s *SettingsPane) layoutTranslatorSelect(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(th, unit.Sp(12), "Translator")
			lbl.Color = colorTextDim
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			labels := [3]string{"Google", "DeepL", "Gemini"}
			children := make([]layout.FlexChild, 0, len(labels)*2-1)
			for i, label := range labels {
				i, label := i, label
				if i > 0 {
					children = append(children, layout.Rigid(layout.Spacer{Width: unit.Dp(6)}.Layout))
				}
				children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					bg := colorBorder
					if i == s.translatorIdx {
						bg = colorBtnBg
					}
					return layoutButton(gtx, th, &s.translatorBtns[i], label, bg)
				}))
			}
			return layout.Flex{}.Layout(gtx, children...)
		}),
	)
}

func (s *SettingsPane) layoutLangBtns(gtx layout.Context, th *material.Theme, btns []widget.Clickable, current string) layout.Dimensions {
	children := make([]layout.FlexChild, 0, len(languages)*2-1)
	for i := range languages {
		i := i
		if i > 0 {
			children = append(children, layout.Rigid(layout.Spacer{Width: unit.Dp(6)}.Layout))
		}
		children = append(children, layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			bg := colorBorder
			if languages[i] == current {
				bg = colorBtnBg
			}
			return layoutButton(gtx, th, &btns[i], languages[i], bg)
		}))
	}
	return layout.Flex{}.Layout(gtx, children...)
}

func (s *SettingsPane) layoutLangSection(gtx layout.Context, th *material.Theme, label string, btns []widget.Clickable, editor *widget.Editor) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(th, unit.Sp(12), label)
			lbl.Color = colorTextDim
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return s.layoutLangBtns(gtx, th, btns, editor.Text())
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(6)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Max.X = gtx.Dp(unit.Dp(120))
			return s.layoutEditorBox(gtx, th, editor, "e.g. fr, de, pt")
		}),
	)
}

func (s *SettingsPane) layoutHotkeyField(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(th, unit.Sp(12), "Toggle Hotkey")
			lbl.Color = colorTextDim
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			gtx.Constraints.Max.X = gtx.Dp(unit.Dp(200))
			return s.layoutEditorBox(gtx, th, &s.hotkeyEditor, "e.g. ctrl+shift+space")
		}),
	)
}

func (s *SettingsPane) layoutLangRow(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return s.layoutLangSection(gtx, th, "Target Language", s.targetLangBtns[:], &s.targetLangEditor)
}

func (s *SettingsPane) layoutFooter(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutButton(gtx, th, &s.saveBtn, "Save", colorBtnBg)
		}),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			if !s.saved {
				return layout.Dimensions{}
			}
			return layout.Inset{Left: unit.Dp(12)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				lbl := material.Label(th, unit.Sp(12), "Saved!")
				lbl.Color = color.NRGBA{R: 0xa0, G: 0xd0, B: 0xa0, A: 0xff}
				return lbl.Layout(gtx)
			})
		}),
	)
}
