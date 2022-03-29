package esqv1

import (
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strconv"
	"strings"

	"github.com/impact-eintr/esq/gnode"
)

const (
	RESP_MESSAGE = 101
	RESP_ERROR   = 102
	RESP_RESULT  = 103
)

var (
	ErrTopicEmpty   = errors.New("topic is empty")
	ErrTopicChannel = errors.New("channel is empty")
)

var (
	pushBase    = "http://%s/push"
	declereBase = "http://%s/declareQueue?topic=%s&bindKey=%s"
	configBase  = "http://%s/config?topic=%s&isAutoAck=%d&mode=%d&msgTTR=%d&msgRetry=%d"
	popBase     = "http://%s/pop?topic=%s&bindKey=%s"
	ackBase     = "http://%s/ack?msgId=%s&topic=%s&bindKey=%s"
)

type MsgPkg struct {
	Body     string `json:"body"`
	Topic    string `json:"topic"`
	Delay    int    `json:"delay"`
	RouteKey string `json:"route_key"`
}

type MMsgPkg struct {
	Body  string
	Delay int
}

type Client struct {
	addr   string
	weight int
}

// 初始化客户端,建立和注册中心节点连接
func NewClient(addr string, weight int) *Client {
	if len(addr) == 0 {
		log.Fatalln("address is empty")
	}

	resp, err := http.Get("http://" + addr + "/ping")
	if err != nil {
		log.Fatalln(err)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil || string(b) != "OK" {
		log.Println(string(b), err)
		return nil
	}

	return &Client{
		addr:   addr,
		weight: weight,
	}
}

func (c *Client) Push_(msg, topic, routeKey string, delay int) error {
	conn, err := net.Dial("tcp4", "172.18.0.4:9502")
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()

	pkgMsg := make([]byte, 0)
	lbuf := make([]byte, 4) // size
	binary.BigEndian.PutUint32(lbuf, uint32(len([]byte(msg))))
	header := fmt.Sprintf("pub %s %s %d\n", topic, routeKey, delay)
	pkgMsg = append(pkgMsg, []byte(header)...)
	pkgMsg = append(pkgMsg, lbuf...)
	pkgMsg = append(pkgMsg, []byte(msg)...)
	_, err = conn.Write(pkgMsg)
	return err
}

func (c *Client) Push(msg, topic, routeKey string, delay int) error {
	var r http.Request
	r.ParseForm()
	data := fmt.Sprintf(`{"body":"%s","topic":"%s","delay":%d,"route_key":"%s"}`, msg, topic, delay, routeKey)
	r.Form.Add("data", data)
	bodystr := strings.TrimSpace(r.Form.Encode())

	request, err := http.NewRequest("POST", fmt.Sprintf(pushBase, c.addr), strings.NewReader(bodystr))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Connection", "Keep-Alive")

	var resp *http.Response
	resp, err = http.DefaultClient.Do(request)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	//b, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	return err
	//}
	//log.Println(string(b))
	return nil
}

// 2(respType) + 4(len) + len
func (c *Client) Pop_(topic, bindKey string) (*gnode.RespMsgData, error) {
	conn, err := net.Dial("tcp4", "172.18.0.4:9502")
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()
	// 先告诉esq需要消费
	header := []byte(fmt.Sprintf("pop %s %s\n", topic, bindKey))
	conn.Write(header)

	header = header[:0]
	conn.Read(header[:2])
	msgType := binary.BigEndian.Uint16(header[:2])
	if msgType != RESP_MESSAGE {
		return nil, fmt.Errorf("invalid msg type:%d\n", msgType)
	}

	header = header[:0]
	conn.Read(header[:4])
	size := binary.BigEndian.Uint32(header[:4])
	buf := make([]byte, size)
	conn.Read(buf)

	msg := &gnode.RespMsgData{}
	if err := json.Unmarshal(buf, msg); err != nil {
		return nil, err
	}
	return msg, nil
}

func (c *Client) Pop(topic, bindKey string) (*gnode.RespMsgData, error) {
	cli := &http.Client{}
	resp, err := cli.Get(fmt.Sprintf(popBase, c.addr, topic, bindKey))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	m := make(map[string]interface{})
	json.Unmarshal(b, &m)

	if m["data"] != nil {
		msg := &gnode.RespMsgData{}
		b, err = json.Marshal(m["data"])
		err = json.Unmarshal(b, &msg)
		if err != nil {
			return nil, err
		}
		//log.Println(msg.Body)
		return msg, nil
	}
	return nil, fmt.Errorf("no message")
}

func (c *Client) Declare(topic, bindKey string) error {
	cli := &http.Client{}
	resp, err := cli.Get(fmt.Sprintf(declereBase, c.addr, topic, bindKey))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	//b, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	return err
	//}
	//log.Println(string(b))
	return nil
}

func (c *Client) Config(topic string, isAutoAck, mode, msgTTR, msgRetry int) error {
	cli := &http.Client{}
	resp, err := cli.Get(fmt.Sprintf(configBase, c.addr, topic, isAutoAck, mode, msgTTR, msgRetry))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	//b, err := ioutil.ReadAll(resp.Body)
	//if err != nil {
	//	return err
	//}
	//log.Println(string(b))
	return nil
}

func (c *Client) Ack(topic, bindKey, id string) error {
	cli := &http.Client{}
	resp, err := cli.Get(fmt.Sprintf(ackBase, c.addr, id, topic, bindKey))
	if err != nil {
		log.Fatalln(err)
		return err
	}
	defer resp.Body.Close()
	return nil
}

func InitClients(endpoints string) ([]*Client, error) {
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("endpoints is empty.")
	}

	var r http.Request
	r.ParseForm()
	r.Form.Add("prefix", "/esq/node")
	bodystr := strings.TrimSpace(r.Form.Encode())

	request, err := http.NewRequest("POST", fmt.Sprintf("http://"+endpoints+"/lease/kv/%s", "nodeinfo"),
		strings.NewReader(bodystr))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Connection", "Keep-Alive")
	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	infobyte, _ := ioutil.ReadAll(resp.Body) // 读出来的是一个数组 所以至少有 [] 长度为2
	if len(infobyte) < 3 {
		return nil, fmt.Errorf("暂时没有数据")
	}

	var clients []*Client
	infos := make([]map[string]string, 0)
	err = json.Unmarshal(infobyte, &infos)
	if err != nil {
		return nil, err
	}

	for _, info := range infos {
		httpAddr := info["http_addr"]
		weight, _ := strconv.Atoi(info["weight"])
		c := NewClient(httpAddr, weight)
		clients = append(clients, c)
	}

	return clients, nil
}

// 权重模式
func getClientByMaxWeightMode(endpoints string) *Client {
	clients, err := InitClients(endpoints)
	if err != nil {
		log.Fatalln(err)
	}

	max := 0
	for _, c := range clients {
		if max < c.weight {
			max = c.weight
		}
	}

	for _, c := range clients {
		if c.weight == max {
			return c
		}
	}

	return nil
}

// 从集群中选择权重最大的队列 建立连接
// endpoint是raftd集群中某个节点的地址
func ChooseQueueInCluster(endpoint string) *Client {
	return getClientByMaxWeightMode(endpoint)
}

// 与队列建立连接 endpoint 是 esq节点的地址
func ChooseQueue(endpoint string) *Client {
	return &Client{
		addr: endpoint,
	}
}
