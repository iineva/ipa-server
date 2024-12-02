// Open client side file from server side over websocket
package websocketfile

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var ErrResponse = errors.New("ErrResponse")
var _ io.Reader = (*websocketFile)(nil)
var _ io.ReaderAt = (*websocketFile)(nil)

type websocketFile struct {
	sync.RWMutex
	conn   *websocket.Conn
	rand   *rand.Rand
	offset int64
	size   int64
}

type CommandType int32

const (
	CommandTypeReadAt CommandType = 1
	CommandTypeSize   CommandType = 2
	CommandTypeName   CommandType = 3
	CommandTypeDone   CommandType = 4
)

type Command struct {
	Command   CommandType            `json:"command"`   // command type
	Param     map[string]interface{} `json:"param"`     // command param
	RequestId string                 `json:"requestId"` // requestId for check response
}

type WebsocketFile interface {
	io.ReaderAt
	io.Reader
	Size() (int64, error)
	Name() (string, error)
	Done(p map[string]interface{}) error
}

func NewWebsocketFile(w http.ResponseWriter, r *http.Request) (WebsocketFile, error) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return nil, err
	}
	return &websocketFile{conn: conn, rand: rand.New(rand.NewSource(time.Now().UnixNano())), size: -1}, nil
}

func (w *websocketFile) setOffset(off int64) {
	w.Lock()
	defer w.Unlock()
	w.offset = off
}

func (w *websocketFile) getOffset() int64 {
	w.RLock()
	defer w.RUnlock()
	return w.offset
}

func (w *websocketFile) ReadAt(p []byte, off int64) (n int, err error) {

	// EOF
	size, err := w.Size()
	if err != nil {
		return 0, err
	}
	if off >= size {
		return 0, io.EOF
	}

	resp, err := w.request(CommandTypeReadAt, map[string]interface{}{
		"offset": off,
		"length": len(p),
	})
	if err != nil {
		return 0, err
	}

	data, ok := resp.Param["data"]
	if !ok {
		return 0, nil
	}

	d, ok := data.(string)
	if !ok {
		return 0, nil
	}

	if d == "" {
		return 0, io.EOF
	}

	// log.Printf("d: %v", d)
	buf, err := base64.StdEncoding.DecodeString(d)
	if err != nil {
		return 0, err
	}

	n = copy(p, buf)
	w.setOffset(off + int64(n))

	return n, nil
}

func (w *websocketFile) Read(p []byte) (n int, err error) {
	return w.ReadAt(p, w.getOffset())
}

func (w *websocketFile) Size() (n int64, err error) {

	// read from cache
	if w.size != -1 {
		return w.size, nil
	}

	resp, err := w.request(CommandTypeSize, nil)
	if err != nil {
		return 0, err
	}

	data, ok := resp.Param["size"]
	if !ok {
		return 0, nil
	}
	size, ok := data.(float64)
	if !ok {
		return 0, nil
	}

	// cache size
	w.size = int64(size)

	return w.size, nil
}

func (w *websocketFile) Name() (n string, err error) {
	resp, err := w.request(CommandTypeName, nil)
	if err != nil {
		return "", err
	}

	data, ok := resp.Param["name"]
	if !ok {
		return "", nil
	}
	name, ok := data.(string)
	if !ok {
		return "", nil
	}

	return name, nil
}

func (w *websocketFile) Done(p map[string]interface{}) error {
	err := w.send(CommandTypeDone, p)
	if err != nil {
		return err
	}
	return nil
}

func (w *websocketFile) send(typ CommandType, p map[string]interface{}) error {
	requestId := fmt.Sprintf("%d", rand.Uint64())
	err := w.conn.WriteJSON(&Command{
		Command:   typ,
		RequestId: requestId,
		Param:     p,
	})
	if err != nil {
		return err
	}
	return nil
}

func (w *websocketFile) request(typ CommandType, p map[string]interface{}) (*Command, error) {
	requestId := fmt.Sprintf("%d", rand.Uint64())
	err := w.conn.WriteJSON(&Command{
		Command:   typ,
		RequestId: requestId,
		Param:     p,
	})
	if err != nil {
		return nil, err
	}

	resp := &Command{}
	err = w.conn.ReadJSON(resp)
	if err != nil {
		return nil, err
	}
	if resp.Command != typ || resp.RequestId != requestId {
		return nil, ErrResponse
	}

	return resp, nil
}
