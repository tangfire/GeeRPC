package example

import (
	"GeeRPC/geerpc"
	"GeeRPC/registry"
	"GeeRPC/xclient"
	"context"
	"log"
	"net"
	"net/http"
	"sync"
	"testing"
	"time"
)

type YGGP struct{}

type YGGPArgs struct {
	Str string
}

// 无重复字符的最长子串
func (yggp YGGP) LengthOfLongestSubstring(args YGGPArgs, reply *int) error {
	s := args.Str
	var slow, fast, length int
	mp := make(map[byte]int)

	slow, fast, length = 0, 0, len(s)
	for fast < length {
		mp[s[fast]]++
		for mp[s[fast]] > 1 && slow < fast {
			mp[s[slow]]--
			slow++
		}
		fast++
		*reply = max(*reply, fast-slow)
	}

	return nil
}

func startYGGPServer(registryAddr string, wg *sync.WaitGroup) {
	var yggp YGGP
	l, _ := net.Listen("tcp", ":0")
	server := geerpc.NewServer()
	_ = server.Register(&yggp)
	registry.Heartbeat(registryAddr, "tcp@"+l.Addr().String(), 0)
	wg.Done()
	server.Accept(l)
}

func yggp(xc *xclient.XClient, ctx context.Context, typ, serviceMethod string, args *YGGPArgs) {
	var reply int
	var err error
	switch typ {
	case "call":
		err = xc.Call(ctx, serviceMethod, args, &reply)
	case "broadcast":
		err = xc.Broadcast(ctx, serviceMethod, args, &reply)
	}
	if err != nil {
		log.Printf("%s %s error: %v", typ, serviceMethod, err)
	} else {
		log.Printf("%s %s success: LengthOfLongestSubstring(%s) is %d", typ, serviceMethod, args.Str, reply)
	}
}

func yggpcall(registry string) {
	TestStr := []string{"abcabcbb", "abcabcbbabcabcbb", "bbbbb", "pwwkew", "pwwkewpwwkew"}
	d := xclient.NewGeeRegistryDiscovery(registry, 0)
	xc := xclient.NewXClient(d, xclient.RandomSelect, nil)
	defer func() { _ = xc.Close() }()
	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			yggp(xc, context.Background(), "call", "YGGP.LengthOfLongestSubstring", &YGGPArgs{
				Str: TestStr[i],
			})
		}(i)
	}
	wg.Wait()
}
func startRegistry(wg *sync.WaitGroup) {
	l, _ := net.Listen("tcp", ":9999")
	registry.HandleHTTP()
	wg.Done()
	_ = http.Serve(l, nil)
}

func TestYGGP(t *testing.T) {
	log.SetFlags(0)
	registryAddr := "http://localhost:9999/_geerpc_/registry"
	var wg sync.WaitGroup
	wg.Add(1)
	go startRegistry(&wg)
	wg.Wait()

	time.Sleep(time.Second)
	wg.Add(1)
	//go startServer(registryAddr, &wg)
	//go startServer(registryAddr, &wg)
	go startYGGPServer(registryAddr, &wg)
	wg.Wait()

	time.Sleep(time.Second)
	yggpcall(registryAddr)
	//call(registryAddr)
	//broadcast(registryAddr)

}
