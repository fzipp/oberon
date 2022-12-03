package canvas

import (
	_ "embed"
	"fmt"
	"html/template"
	"image"
	"image/color"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var (
	//go:embed web/canvas-websocket.js
	javaScriptCode []byte

	//go:embed web/index.html.tmpl
	indexHTMLCode     string
	indexHTMLTemplate = template.Must(template.New("index.html.tmpl").Parse(indexHTMLCode))
)

func ListenAndServe(addr string, run func(*Context), size image.Rectangle) error {
	config := config{
		title:               "Project Oberon RISC Emulator",
		width:               size.Dx(),
		height:              size.Dy(),
		backgroundColor:     color.Black,
		eventMask:           maskMouseMove | maskMouseDown | maskMouseUp | maskKeyDown | maskKeyUp | maskTouchMove | maskClipboardChange,
		cursorDisabled:      true,
		contextMenuDisabled: true,
	}
	return http.ListenAndServe(addr, newServeMux(run, config))
}

func newServeMux(run func(*Context), config config) *http.ServeMux {
	mux := http.NewServeMux()
	mux.Handle("/", &htmlHandler{
		config: config,
	})
	mux.HandleFunc("/canvas-websocket.js", javaScriptHandler)
	mux.Handle("/draw", &drawHandler{
		config: config,
		draw:   run,
	})
	return mux
}

type htmlHandler struct {
	config config
}

func (h *htmlHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	model := map[string]any{
		"DrawURL":             template.URL("draw"),
		"Width":               h.config.width,
		"Height":              h.config.height,
		"Title":               h.config.title,
		"BackgroundColor":     template.CSS(rgbaString(h.config.backgroundColor)),
		"EventMask":           h.config.eventMask,
		"CursorDisabled":      h.config.cursorDisabled,
		"ContextMenuDisabled": h.config.contextMenuDisabled,
		"FullPage":            h.config.fullPage,
		"ReconnectInterval":   int64(h.config.reconnectInterval / time.Millisecond),
	}
	err := indexHTMLTemplate.Execute(w, model)
	if err != nil {
		log.Println(err)
		return
	}
}

func rgbaString(c color.Color) string {
	clr := color.RGBAModel.Convert(c).(color.RGBA)
	return fmt.Sprintf("rgba(%d, %d, %d, %g)", clr.R, clr.G, clr.B, float64(clr.A)/255)
}

func javaScriptHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Add("Content-Type", "text/javascript")
	_, err := w.Write(javaScriptCode)
	if err != nil {
		log.Println(err)
		return
	}
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type drawHandler struct {
	config config
	draw   func(*Context)
}

func (h *drawHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	events := make(chan Event)
	defer close(events)
	draws := make(chan []byte)
	defer close(draws)

	wg := sync.WaitGroup{}
	wg.Add(2)
	go readMessages(conn, events, &wg)
	go writeMessages(conn, draws, &wg)

	ctx := newContext(draws, events, h.config)
	go func() {
		defer wg.Done()
		h.draw(ctx)
	}()

	wg.Wait()
	wg.Add(1)
	events <- CloseEvent{}
	wg.Wait()
}

func writeMessages(conn *websocket.Conn, messages <-chan []byte, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		message := <-messages
		err := conn.WriteMessage(websocket.BinaryMessage, message)
		if err != nil {
			break
		}
	}
}

func readMessages(conn *websocket.Conn, events chan<- Event, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			break
		}
		if messageType != websocket.BinaryMessage {
			continue
		}
		event, err := decodeEvent(p)
		if err != nil {
			continue
		}
		events <- event
	}
}

type config struct {
	title               string
	width               int
	height              int
	backgroundColor     color.Color
	eventMask           eventMask
	cursorDisabled      bool
	contextMenuDisabled bool
	fullPage            bool
	reconnectInterval   time.Duration
}
