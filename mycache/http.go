package mycache

import (
	"fmt"
	"io/ioutil"
	"log"
	"mycache/consistenthash"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

//HTTP服务端
const (
	defaultBasePath = "/_mycache/"
	defaultReplicas = 50
)

type HTTPPool struct {
	self        string
	basePath    string
	mu          sync.Mutex //这下面三个是为获取远程结点而加上去的
	peers       *consistenthash.Map
	httpGetters map[string]*httpGetter //这里的key是一个远程结点的URL,每一个远程节点对应一个 httpGetter，因为 httpGetter 与远程节点的地址 `baseURL` 有关。
}

func NewHTTPPool(self string) *HTTPPool {
	return &HTTPPool{
		self:     self,
		basePath: defaultBasePath,
	}
}

func (p *HTTPPool) Log(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", p.self, fmt.Sprintf(format, v...))
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if !strings.HasPrefix(req.URL.Path, p.basePath) {
		panic("HTTPPool serving unexpected path: " + req.URL.Path)
	}
	p.Log("%s %s", req.Method, req.URL.Path)

	// /<basepath>/<groupname>/<key> required 因为需要 groupname 和key 才能到mycache里面去查询缓存数据
	parts := strings.SplitN(req.URL.Path[len(p.basePath):], "/", 2)
	if len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		http.Error(w, "没这样的group"+groupName, http.StatusNotFound)
		return
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}

//HTTP客户端
type httpGetter struct {
	baseURL string
} //baseURL 表示将要访问的远程节点的地址

func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf(
		"%v%v/%v",
		h.baseURL,
		url.QueryEscape(group),
		url.QueryEscape(key),
	)
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Server returned: %v", res.Status)
	}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}
	return bytes, nil
}

var _ PeerGetter = (*httpGetter)(nil) //上面的Get函数是实现PeerGetter接口

func (p *HTTPPool) Set(peers ...string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	/*
		实例化
	*/
	p.peers = consistenthash.New(defaultReplicas, nil)
	/*
		添加传入节点
	*/
	p.peers.Add(peers...)
	/*
		为每个节点配置Getter
	*/
	p.httpGetters = make(map[string]*httpGetter, len(peers))

	/*
		peer如http://xxxx.com/
		basePath如 _Gache//
		合起来就是一个节点的访问路径
	*/
	for _, peer := range peers {
		p.httpGetters[peer] = &httpGetter{
			baseURL: peer + p.basePath,
		}

	}
}
func (p *HTTPPool) PickPeer(key string) (PeerGetter, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

var _ PeerPicker = (*HTTPPool)(nil) //上面是实现PeerPicker接口
