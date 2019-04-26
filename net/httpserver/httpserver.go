package httpserver

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/champly/lib4go/security/md5"
	"github.com/olivere/elastic"
)

type ccData struct {
	SessionID string `json:"session_id"`
	URL       string `json:"url"`
	Method    string `json:"method"`
	Data      string `json:"data"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}

type Context struct {
	Request  *http.Request
	Response http.ResponseWriter
}

type HttpServer struct {
	mux       *http.ServeMux
	closeChan chan int
	data      chan ccData
	esClient  *elastic.Client
}

func New() *HttpServer {
	client, _ := elastic.NewClient(
		elastic.SetURL("http://localhost:9200"),
		elastic.SetErrorLog(log.New(os.Stderr, "ELASTIC ", log.Ldate|log.Ltime|log.LstdFlags)),
	)

	h := &HttpServer{
		mux:       http.NewServeMux(),
		closeChan: make(chan int),
		data:      make(chan ccData, 10),
		esClient:  client,
	}
	go h.sendToEs()
	return h
}

func (h *HttpServer) hookHandler(handler func(ctx *Context)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		h.before(r)
		ctx := &Context{r, w}
		handler(ctx)
		h.after(r)
	}
}

func (h *HttpServer) before(r *http.Request) {
	v := r.Header.Get("__call_chain_v__")
	if v == "" {
		v = "1"
	} else {
		v += ".1"
	}

	sid := r.Header.Get("__call_chain__")
	if sid == "" {
		sid = getGUID()
		r.Header.Set("__call_chain__", sid)
	}

	r.Header.Set("__call_chain_v__", v)

	h.data <- ccData{
		SessionID: sid,
		URL:       r.URL.String(),
		Method:    r.Method,
		Data:      fmt.Sprintf(`{"header":"%+v"}`, r.Header),
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Version:   v,
	}
}

func (h *HttpServer) after(r *http.Request) {
	fmt.Println(r.Header)
	sid := r.Header.Get("__call_chain__")
	v := r.Header.Get("__call_chain_v__")
	if v != "" {
		n := strings.LastIndex(v, ".")
		vv, err := strconv.Atoi(v[n+1:])
		if err != nil {
			fmt.Println("version is not rule:", v)
			return
		}
		v = fmt.Sprintf("%s%d", v[:n+1], vv+1)
	} else {
		v = "not found version"
	}

	h.data <- ccData{
		SessionID: sid,
		URL:       r.URL.String(),
		Method:    r.Method,
		Data:      fmt.Sprintf(`{"header":"%+v"}`, r.Header),
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Version:   v,
	}
}

func (h *HttpServer) HandleFunc(pattern string, handler func(ctx *Context)) {
	h.mux.HandleFunc(pattern, h.hookHandler(handler))
}

func (h *HttpServer) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, h.mux)
}

func (h *HttpServer) sendToEs() {
END:
	for {
	CONTINUE:
		select {
		case <-h.closeChan:
			fmt.Println("http server stopd")
			break END
		case d, ok := <-h.data:
			if !ok {
				fmt.Println("http server close data")
				break END
			}
			ok, err := h.esClient.IndexExists("call_chain").Do(context.Background())
			if err != nil {
				fmt.Println("judge index exists err:", err)
				break CONTINUE
			}
			if !ok {
				// res, err := h.esClient.CreateIndex("call_chain").Do(context.Background())
				h.esClient.CreateIndex("call_chain").Do(context.Background())
				// if err != nil {
				// if !strings.Contains("already exists", err.Error()) {
				// fmt.Println("create index err:", err)
				// break CONTINUE
				// }
				// }
				// if !res.Acknowledged {
				// fmt.Println("create index fail:", res)
				// break CONTINUE
				// }
			}

			b, _ := json.Marshal(d)
			res, err := h.esClient.Index().
				Index("call_chain").
				BodyString(string(b)).
				Refresh("true").
				Do(context.Background())

			if err != nil {
				fmt.Println("insert data err:", err)
				break CONTINUE
			}
			if res.Result != "created" {
				fmt.Println("insert data fail:", res)
			}
		}
	}
}

func getGUID() string {
	b := make([]byte, 48)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""

	}
	return md5.Encrypt(base64.URLEncoding.EncodeToString(b))
}
