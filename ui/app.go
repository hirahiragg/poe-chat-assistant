package ui

import (
	"image"
	"io"
	"strings"

	"sync/atomic"

	"gioui.org/app"
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

type LoadMoreResult int

const (
	LoadMoreFound    LoadMoreResult = iota // found messages
	LoadMoreNotFound                       // no messages in range, but more file to read
	LoadMoreEOF                            // reached beginning of file
)

type App struct {
	window       *app.Window
	theme        *material.Theme
	chatList     *ChatList
	detail       *DetailPane
	store        *chat.Store
	cfg          *config.Config
	translate    TranslateFunc
	settingsBtn  widget.Clickable
	loadMoreBtn  widget.Clickable
	settings     *SettingsPane
	showSettings bool
	onConfigSave func(*config.Config)
	onLoadMore   func() LoadMoreResult
	loading      bool
	loadResult   LoadMoreResult
	hidden       atomic.Bool
	toggleReq    atomic.Bool
}

func NewApp(store *chat.Store, cfg *config.Config, translateFn TranslateFunc, onConfigSave func(*config.Config), onLoadMore func() LoadMoreResult) *App {
	w := new(app.Window)
	w.Option(
		app.Size(unit.Dp(680), unit.Dp(420)),
		app.MinSize(unit.Dp(500), unit.Dp(300)),
	)
	return &App{
		window:       w,
		theme:        newTheme(),
		chatList:     NewChatList(&cfg.Filters),
		detail:       NewDetailPane(),
		store:        store,
		cfg:          cfg,
		translate:    translateFn,
		onConfigSave: onConfigSave,
		onLoadMore:   onLoadMore,
	}
}

func (a *App) Window() *app.Window { return a.window }

func (a *App) SetTranslateFunc(fn TranslateFunc) {
	a.translate = fn
}

func (a *App) RequestToggle() {
	a.toggleReq.Store(true)
	a.window.Invalidate()
}

func (a *App) Run() error {
	var ops op.Ops

	for {
		switch e := a.window.Event().(type) {
		case app.DestroyEvent:
			return e.Err
		case app.FrameEvent:
			if a.toggleReq.CompareAndSwap(true, false) {
				if a.hidden.Load() {
					a.hidden.Store(false)
					a.window.Option(app.Windowed.Option())
					a.window.Perform(system.ActionRaise)
				} else {
					a.hidden.Store(true)
					a.window.Perform(system.ActionMinimize)
				}
			}
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

	if a.loadMoreBtn.Clicked(gtx) && !a.loading && a.loadResult != LoadMoreEOF && a.onLoadMore != nil {
		a.loading = true
		go func() {
			a.loadResult = a.onLoadMore()
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
			msg := messages[sel]
			if msg.Channel == chat.ChannelWhisperIn || msg.Channel == chat.ChannelWhisperOut {
				txt = "@" + msg.Player + " " + txt
			}
			gtx.Execute(clipboard.WriteCmd{
				Type: "application/text",
				Data: io.NopCloser(strings.NewReader(txt)),
			})
		}
	}
}

func (a *App) saveFilters() {
	a.cfg.Filters = a.chatList.FiltersConfig()
	go a.cfg.Save()
}

func (a *App) onSelectionChanged() {
	messages := a.chatList.FilterMessages(a.store.List())
	sel := a.chatList.Selected()
	if sel >= 0 && sel < len(messages) {
		a.detail.SwitchMessage(&messages[sel])
	}
}

func (a *App) layout(gtx layout.Context) layout.Dimensions {
	fillRect(gtx, colorBg, gtx.Constraints.Max)

	messages := a.chatList.FilterMessages(a.store.List())

	if a.chatList.Selected() == -1 && len(messages) > 0 {
		a.chatList.selected = 0
	}

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
}

func (a *App) layoutLeftPane(gtx layout.Context, messages []chat.Message) layout.Dimensions {
	dims, ev := a.chatList.Layout(gtx, a.theme, messages, &a.settingsBtn, func(gtx layout.Context) layout.Dimensions {
		if a.onLoadMore == nil {
			return layout.Dimensions{}
		}
		return layout.Inset{
			Top: unit.Dp(4), Bottom: unit.Dp(8),
			Left: unit.Dp(12), Right: unit.Dp(12),
		}.Layout(gtx, func(gtx layout.Context) layout.Dimensions {
			var label string
			switch {
			case a.loading:
				label = "Loading..."
			case a.loadResult == LoadMoreEOF:
				label = "No more messages"
			case a.loadResult == LoadMoreNotFound:
				label = "Load More (no chats found, retry)"
			default:
				label = "Load More"
			}
			return layoutButton(gtx, a.theme, &a.loadMoreBtn, label, colorBtnBg)
		})
	})
	if ev.SelectionChanged {
		a.onSelectionChanged()
	}
	if ev.FilterChanged {
		a.saveFilters()
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
