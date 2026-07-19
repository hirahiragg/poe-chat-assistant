package ui

import (
	"image"
	"io"
	"strings"

	"gioui.org/app"
	"gioui.org/font"
	"gioui.org/io/clipboard"
	"gioui.org/io/key"
	"gioui.org/io/system"
	"gioui.org/layout"
	"gioui.org/op"
	"gioui.org/op/clip"
	"gioui.org/op/paint"
	"gioui.org/unit"
	"gioui.org/widget"
	"gioui.org/widget/material"

	"github.com/hirahiragg/poe-chat-assistant/internal/chat"
	"github.com/hirahiragg/poe-chat-assistant/internal/config"
)

type TranslateFunc func(msg chat.Message, replyText string, done func(inbound, outbound string))

type App struct {
	window       *app.Window
	theme        *material.Theme
	chatList     *ChatList
	detail       *DetailPane
	store        *chat.Store
	translate    TranslateFunc
	settingsBtn  widget.Clickable
	loadMoreBtn  widget.Clickable
	settings     *SettingsPane
	showSettings bool
	onConfigSave func(*config.Config)
	onLoadMore   func()
	loading      bool
}

func NewApp(store *chat.Store, translateFn TranslateFunc, onConfigSave func(*config.Config), onLoadMore func()) *App {
	w := new(app.Window)
	w.Option(
		app.Title("PoE Chat Assistant"),
		app.Size(unit.Dp(680), unit.Dp(420)),
		app.MinSize(unit.Dp(500), unit.Dp(300)),
	)
	return &App{
		window:       w,
		theme:        newTheme(),
		chatList:     NewChatList(),
		detail:       NewDetailPane(),
		store:        store,
		translate:    translateFn,
		onConfigSave: onConfigSave,
		onLoadMore:   onLoadMore,
	}
}

func (a *App) Window() *app.Window { return a.window }

func (a *App) SetTranslateFunc(fn TranslateFunc) {
	a.translate = fn
}

func (a *App) Run() error {
	var ops op.Ops

	for {
		switch e := a.window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			gtx := app.NewContext(&ops, e)
			a.handleKeys(gtx)
			if a.showSettings {
				a.settings.HandleActions(gtx)
				a.settings.Layout(gtx, a.theme)
			} else {
				a.handleActions(gtx)
				a.layout(gtx)
			}
			e.Frame(gtx.Ops)
		}
	}
}

func (a *App) OpenSettings(cfg *config.Config) {
	a.settings = NewSettingsPane(cfg, func(c *config.Config) {
		if a.onConfigSave != nil {
			a.onConfigSave(c)
		}
	}, func() {
		a.showSettings = false
	})
	a.showSettings = true
}

func (a *App) handleKeys(gtx layout.Context) {
	for {
		ev, ok := gtx.Event(key.Filter{Name: key.NameUpArrow}, key.Filter{Name: key.NameDownArrow}, key.Filter{Name: key.NameEscape})
		if !ok {
			break
		}
		if e, ok := ev.(key.Event); ok && e.State == key.Press {
			switch e.Name {
			case key.NameEscape:
				if a.showSettings {
					a.showSettings = false
				} else {
					a.window.Perform(system.ActionClose)
				}
			case key.NameUpArrow:
				if !a.showSettings {
					a.chatList.SelectPrev()
					a.onSelectionChanged()
				}
			case key.NameDownArrow:
				if !a.showSettings {
					messages := a.store.List()
					a.chatList.SelectNext(len(messages))
					a.onSelectionChanged()
				}
			}
		}
	}
}

func (a *App) handleActions(gtx layout.Context) {
	if a.settingsBtn.Clicked(gtx) {
		cfg := config.Load()
		a.OpenSettings(cfg)
		return
	}

	if a.loadMoreBtn.Clicked(gtx) && !a.loading && a.onLoadMore != nil {
		a.loading = true
		go func() {
			a.onLoadMore()
			a.loading = false
			a.window.Invalidate()
		}()
	}

	messages := a.chatList.FilterMessages(a.store.List())
	sel := a.chatList.Selected()

	if sel < 0 || sel >= len(messages) {
		return
	}

	if a.detail.TranslateMsgClicked(gtx) {
		a.detail.SetTranslatingMsg(true)
		msg := messages[sel]
		go func() {
			a.translate(msg, "", func(inbound, _ string) {
				a.detail.SetTranslatedMsg(inbound)
				a.window.Invalidate()
			})
		}()
	}

	if a.detail.TranslateOutClicked(gtx) {
		replyText := a.detail.ReplyText()
		if replyText != "" {
			a.detail.SetTranslatingOut(true)
			msg := messages[sel]
			go func() {
				a.translate(msg, replyText, func(_, outbound string) {
					a.detail.SetTranslatedOut(outbound)
					a.window.Invalidate()
				})
			}()
		}
	}

	if a.detail.CopyClicked(gtx) {
		if txt := a.detail.TranslatedOut(); txt != "" {
			gtx.Execute(clipboard.WriteCmd{
				Type: "application/text",
				Data: io.NopCloser(strings.NewReader(txt)),
			})
		}
	}
}

func (a *App) onSelectionChanged() {
	a.detail.ClearReply()
	a.detail.ClearTranslation()
}

func (a *App) layout(gtx layout.Context) layout.Dimensions {
	fillRect(gtx, colorBg, gtx.Constraints.Max)

	messages := a.chatList.FilterMessages(a.store.List())

	if a.chatList.Selected() == -1 && len(messages) > 0 {
		a.chatList.selected = 0
	}

	return layout.Flex{Axis: layout.Vertical}.Layout(gtx,
		layout.Rigid(func(gtx layout.Context) layout.Dimensions {
			return a.layoutTopBar(gtx)
		}),
		layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
			return layout.Flex{}.Layout(gtx,
				layout.Flexed(0.4, func(gtx layout.Context) layout.Dimensions {
					return a.layoutLeftPane(gtx, messages)
				}),
				layout.Rigid(func(gtx layout.Context) layout.Dimensions {
					size := image.Point{X: gtx.Dp(unit.Dp(1)), Y: gtx.Constraints.Max.Y}
					defer clip.Rect{Max: size}.Push(gtx.Ops).Pop()
					paint.ColorOp{Color: colorBorder}.Add(gtx.Ops)
					paint.PaintOp{}.Add(gtx.Ops)
					return layout.Dimensions{Size: size}
				}),
				layout.Flexed(0.6, func(gtx layout.Context) layout.Dimensions {
					return a.layoutRightPane(gtx, messages)
				}),
			)
		}),
	)
}

func (a *App) layoutTopBar(gtx layout.Context) layout.Dimensions {
	return layout.Background{}.Layout(gtx,
		func(gtx layout.Context) layout.Dimensions {
			size := image.Point{X: gtx.Constraints.Max.X, Y: gtx.Constraints.Min.Y}
			fillRect(gtx, colorSurface, size)
			return layout.Dimensions{Size: size}
		},
		func(gtx layout.Context) layout.Dimensions {
			return layout.Inset{
				Top: unit.Dp(6), Bottom: unit.Dp(6),
				Left: unit.Dp(12), Right: unit.Dp(12),
			}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
				return layout.Flex{Alignment: layout.Middle, Spacing: layout.SpaceBetween}.Layout(gtx,
					layout.Flexed(1, func(gtx layout.Context) layout.Dimensions {
						lbl := material.Label(a.theme, unit.Sp(13), "PoE Chat Assistant")
						lbl.Color = colorTextDim
						lbl.Font.Weight = font.Medium
						return lbl.Layout(gtx)
					}),
					layout.Rigid(func(gtx layout.Context) layout.Dimensions {
						return material.Clickable(gtx, &a.settingsBtn, func(gtx layout.Context) layout.Dimensions {
							return layout.Inset{
								Top: unit.Dp(2), Bottom: unit.Dp(2),
								Left: unit.Dp(8), Right: unit.Dp(8),
							}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
								lbl := material.Label(a.theme, unit.Sp(14), "Settings")
								lbl.Color = colorAccent
								return lbl.Layout(gtx)
							})
						})
					}),
				)
			})
		},
	)
}

func (a *App) layoutLeftPane(gtx layout.Context, messages []chat.Message) layout.Dimensions {
	dims, changed := a.chatList.Layout(gtx, a.theme, messages, func(gtx layout.Context) layout.Dimensions {
		if a.onLoadMore == nil {
			return layout.Dimensions{}
		}
		return layout.Inset{
			Top: unit.Dp(4), Bottom: unit.Dp(8),
			Left: unit.Dp(12), Right: unit.Dp(12),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			label := "Load More"
			if a.loading {
				label = "Loading..."
			}
			return layoutButton(gtx, a.theme, &a.loadMoreBtn, label, colorBtnBg)
		})
	})
	if changed {
		a.onSelectionChanged()
	}
	return dims
}

func (a *App) layoutRightPane(gtx layout.Context, messages []chat.Message) layout.Dimensions {
	sel := a.chatList.Selected()
	if sel < 0 || sel >= len(messages) {
		return layout.Center.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			lbl := material.Label(a.theme, unit.Sp(13), "Select a chat")
			lbl.Color = colorTextDim
			return lbl.Layout(gtx)
		})
	}
	return a.detail.Layout(gtx, a.theme, &messages[sel])
}
