package http

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

type CallChainCb struct {
	closeChan chan int
	data      chan ccData
	esClient  *elastic.Client
}

func NewCallChainCb() *CallChainCb {
	client, _ := elastic.NewClient(
		elastic.SetURL("http://localhost:9200"),
		elastic.SetErrorLog(log.New(os.Stderr, "ELASTIC ", log.Ldate|log.Ltime|log.LstdFlags)),
	)
	c := &CallChainCb{
		closeChan: make(chan int),
		data:      make(chan ccData, 10),
		esClient:  client,
	}
	go c.sendToEs()
	return c
}

func (c *CallChainCb) Before(req *http.Request) {
	v := req.Header.Get("__call_chain_v__")
	if v == "" {
		return
	}

	sid := req.Header.Get("__call_chain__")
	c.data <- ccData{
		SessionID: sid,
		URL:       req.URL.String(),
		Method:    req.Method,
		Data:      fmt.Sprintf(`{"header":"%+v"}`, req.Header),
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Version:   v,
	}
}

func (c *CallChainCb) After(req *http.Request, resp *http.Response, err error) {
	v := req.Header.Get("__call_chain_v__")
	if v == "" {
		return
	}

	sid := req.Header.Get("__call_chain__")
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

	c.data <- ccData{
		SessionID: sid,
		URL:       req.URL.String(),
		Method:    req.Method,
		Data:      fmt.Sprintf(`{"header":"%+v"}`, req.Header),
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Version:   v,
	}
}

func (c *CallChainCb) sendToEs() {
END:
	for {
	CONTINUE:
		select {
		case <-c.closeChan:
			fmt.Println("http server stopd")
			break END
		case d, ok := <-c.data:
			if !ok {
				fmt.Println("http server close data")
				break END
			}
			ok, err := c.esClient.IndexExists("call_chain").Do(context.Background())
			if err != nil {
				fmt.Println("judge index exists err:", err)
				break CONTINUE
			}
			if !ok {
				res, err := c.esClient.CreateIndex("call_chain").Do(context.Background())
				if err != nil {
					fmt.Println("create index err:", err)
					break CONTINUE
				}
				if !res.Acknowledged {
					fmt.Println("create index fail:", res)
					break CONTINUE
				}
			}

			b, _ := json.Marshal(d)
			res, err := c.esClient.Index().
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
