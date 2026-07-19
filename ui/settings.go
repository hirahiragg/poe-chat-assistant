package ui

import (
	"image"
	"image/color"

	"gioui.org/font"
	"gioui.org/layout"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/hirahiragg/poe-chat-assistant/internal/config"
)

type SettingsPane struct {
	logPathEditor    widget.Editor
	deeplKeyEditor   widget.Editor
	geminiKeyEditor  widget.Editor
	sourceLangEditor widget.Editor
	targetLangEditor widget.Editor

	translatorBtns [3]widget.Clickable
	translatorIdx  int

	saveBtn  widget.Clickable
	closeBtn widget.Clickable

	saved    bool
	onSave   func(*config.Config)
	onClose  func()
}

var translators = [3]string{"google", "deepl", "gemini"}

func NewSettingsPane(cfg *config.Config, onSave func(*config.Config), onClose func()) *SettingsPane {
	s := &SettingsPane{
		onSave:  onSave,
		onClose: onClose,
	}
	s.logPathEditor.SingleLine = true
	s.logPathEditor.SetText(cfg.LogPath)
	s.deeplKeyEditor.SingleLine = true
	s.deeplKeyEditor.SetText(cfg.DeepLKey)
	s.geminiKeyEditor.SingleLine = true
	s.geminiKeyEditor.SetText(cfg.GeminiKey)
	s.sourceLangEditor.SingleLine = true
	s.sourceLangEditor.SetText(cfg.SourceLang)
	s.targetLangEditor.SingleLine = true
	s.targetLangEditor.SetText(cfg.TargetLang)

	for i, t := range translators {
		if t == cfg.Translator {
			s.translatorIdx = i
			break
		}
	}
	return s
}

func (s *SettingsPane) HandleActions(gtx layout.Context) {
	for i := range s.translatorBtns {
		if s.translatorBtns[i].Clicked(gtx) {
			s.translatorIdx = i
			s.saved = false
		}
	}

	if s.saveBtn.Clicked(gtx) {
		cfg := &config.Config{
			LogPath:    s.logPathEditor.Text(),
			Translator: translators[s.translatorIdx],
			DeepLKey:   s.deeplKeyEditor.Text(),
			GeminiKey:  s.geminiKeyEditor.Text(),
			SourceLang: s.sourceLangEditor.Text(),
			TargetLang: s.targetLangEditor.Text(),
		}
		s.onSave(cfg)
		s.saved = true
	}

	if s.closeBtn.Clicked(gtx) {
		s.onClose()
	}
}

func (s *SettingsPane) Layout(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Background{}.Layout(gtx,
		func(gtx layout.Context) layout.Dimensions {
			fillRect(gtx, colorBg, gtx.Constraints.Max)
			return layout.Dimensions{Size: gtx.Constraints.Max}
		},
		func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top: unit.Dp(16), Bottom: unit.Dp(16),
				Left: unit.Dp(24), Right: unit.Dp(24),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return s.layoutHeader(gtx, th)
					}),
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
					layout.Rigid(layout.Spacer{Height: unit.Dp(24)}.Layout),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return s.layoutFooter(gtx, th)
					}),
				)
			})
		},
	)
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
					bg := colorCard
					if i == s.translatorIdx {
						bg = colorAccent
					}
					return layoutButton(gtx, th, &s.translatorBtns[i], label, bg)
				}))
			}
			return layout.Flex{}.Layout(gtx, children...)
		}),
	)
}

func (s *SettingsPane) layoutLangRow(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(th, unit.Sp(12), "Language")
			lbl.Color = colorTextDim
			return lbl.Layout(gtx)
		}),
		layout.Rigid(layout.Spacer{Height: unit.Dp(4)}.Layout),
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Max.X = gtx.Dp(unit.Dp(60))
					return s.layoutEditorBox(gtx, th, &s.sourceLangEditor, "en")
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					return layout.Inset{Left: unit.Dp(8), Right: unit.Dp(8)}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(th, unit.Sp(14), "→")
						lbl.Color = colorTextDim
						return lbl.Layout(gtx)
					})
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					gtx.Constraints.Max.X = gtx.Dp(unit.Dp(60))
					return s.layoutEditorBox(gtx, th, &s.targetLangEditor, "ja")
				}),
			)
		}),
	)
}

func (s *SettingsPane) layoutFooter(gtx layout.Context, th *material.Theme) layout.Dimensions {
	return layout.Flex{Alignment: layout.Middle}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return layoutButton(gtx, th, &s.saveBtn, "Save", colorAccent)
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
