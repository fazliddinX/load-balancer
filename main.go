//package main
//
//import (
//	"context"
//	"log"
//	"net/http"
//	"sync"
//	"time"
//)
//
//var indicate = true
//
//// Backend interface defines methods to interact with backend servers
//type Backend interface {
//	SetAlive(bool)
//	IsAlive() bool
//	GetURL() string
//	Serve(http.ResponseWriter, *http.Request)
//}
//
//// backend struct represents a backend server
//type backend struct {
//	url         string
//	alive       bool
//	mux         sync.RWMutex
//	connections int
//}
//
//// SetAlive sets the alive status of the backend server
//func (b *backend) SetAlive(alive bool) {
//	b.mux.Lock()
//	defer b.mux.Unlock()
//	b.alive = alive
//}
//
//// IsAlive returns the alive status of the backend server
//func (b *backend) IsAlive() bool {
//	b.mux.RLock()
//	defer b.mux.RUnlock()
//	return b.alive
//}
//
//// GetURL returns the URL of the backend server
//func (b *backend) GetURL() string {
//	return b.url
//}
//
//// GetActiveConnections returns the number of active connections to the backend
//func (b *backend) GetActiveConnections() int {
//	b.mux.RLock()
//	defer b.mux.RUnlock()
//	log.Println("address: ", b.url, "  count users -", b.connections)
//	return b.connections
//}
//
//// Serve relays the request to the backend using a reverse proxy
//func (b *backend) Serve(w http.ResponseWriter, r *http.Request) {
//	b.mux.Lock()
//	b.connections++
//	b.mux.Unlock()
//
//	urlDir := b.GetURL()
//	http.Redirect(w, r, urlDir, http.StatusTemporaryRedirect)
//}
//
//// ServerPool interface defines methods to manage backend servers
//type ServerPool interface {
//	GetBackends() []Backend
//	GetNextValidPeer() Backend
//	AddBackend(Backend)
//	GetServerPoolSize() int
//}
//
//// lcServerPool implements ServerPool for Least Connections strategy
//type lcServerPool struct {
//	backends []Backend
//	mux      sync.RWMutex
//}
//
//// GetBackends returns all backends in the pool
//func (s *lcServerPool) GetBackends() []Backend {
//	s.mux.RLock()
//	defer s.mux.RUnlock()
//	return s.backends
//}
//
//// AddBackend adds a backend server to the pool
//func (s *lcServerPool) AddBackend(b Backend) {
//	s.mux.Lock()
//	defer s.mux.Unlock()
//	s.backends = append(s.backends, b)
//}
//
//// GetServerPoolSize returns the size of the server pool
//func (s *lcServerPool) GetServerPoolSize() int {
//	s.mux.RLock()
//	defer s.mux.RUnlock()
//	return len(s.backends)
//}
//
//// GetNextValidPeer selects the backend server with the least active connections
//func (s *lcServerPool) GetNextValidPeer() Backend {
//	var res Backend
//
//	if indicate {
//		res = s.backends[0]
//		indicate = false
//		log.Println("[1] - ", s.backends[0].GetURL())
//	} else {
//		res = s.backends[0]
//		indicate = true
//		log.Println("[2] - ", s.backends[0].GetURL())
//	}
//
//	return res
//}
//
//// LoadBalancer interface defines methods to handle incoming requests
//type LoadBalancer interface {
//	Serve(http.ResponseWriter, *http.Request)
//}
//
//// loadBalancer struct is a wrapper around ServerPool
//type loadBalancer struct {
//	serverPool ServerPool
//}
//
//// Serve forwards the request to the backend with the least connections
//func (lb *loadBalancer) Serve(w http.ResponseWriter, r *http.Request) {
//	peer := lb.serverPool.GetNextValidPeer()
//	if peer != nil {
//		peer.Serve(w, r)
//		return
//	}
//	http.Error(w, "Service not available", http.StatusServiceUnavailable)
//}
//
//// HealthCheck checks the status of each backend periodically
//func HealthCheck(ctx context.Context, s ServerPool) {
//	for {
//		aliveChannel := make(chan bool, 1)
//
//		for _, b := range s.GetBackends() {
//
//			requestCtx, stop := context.WithTimeout(ctx, 10*time.Second)
//			defer stop()
//
//			go IsBackendAlive(requestCtx, aliveChannel, b.GetURL())
//
//			select {
//			case <-ctx.Done():
//				log.Println("Gracefully shutting down health check")
//				return
//			case alive := <-aliveChannel:
//				b.SetAlive(alive)
//			}
//		}
//
//		// Check every 30 seconds
//		time.Sleep(time.Minute)
//	}
//}
//
//// IsBackendAlive checks if the backend server is alive
//func IsBackendAlive(ctx context.Context, aliveChannel chan<- bool, url string) {
//	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
//	if err != nil {
//		log.Println("Error creating request:", err)
//		aliveChannel <- false
//		return
//	}
//
//	resp, err := http.DefaultClient.Do(req)
//	if err != nil || resp.StatusCode != http.StatusOK {
//		if err != nil {
//			log.Println("Request error:", err)
//		} else {
//			log.Println("Health check failed with status code:", resp.StatusCode)
//		}
//		aliveChannel <- false
//		return
//	}
//
//	aliveChannel <- true
//}
//
//// AllowRetry checks if the request can be retried
//func AllowRetry(r *http.Request) bool {
//	if r.Context().Value("RETRY_ATTEMPTED") == nil {
//		return true
//	}
//	return false
//}
//
//func main() {
//	URL1 := ("http://3.125.33.48:8081/swagger/index.html")
//	URL2 := ("http://3.125.33.48:8081/swagger/index.html")
//
//	backend1 := &backend{
//		url:   URL1,
//		alive: true,
//	}
//
//	backend2 := &backend{
//		url:   URL2,
//		alive: true,
//	}
//
//	serverPool := &lcServerPool{}
//	serverPool.AddBackend(backend1)
//	serverPool.AddBackend(backend2)
//
//	lb := &loadBalancer{
//		serverPool: serverPool,
//	}
//
//	http.HandleFunc("/", lb.Serve)
//
//	// Run health check in the background
//	go func() {
//		ctx := context.Background()
//		HealthCheck(ctx, serverPool)
//	}()
//
//	log.Println("Server is listening on port :8080")
//	log.Fatal(http.ListenAndServe(":8080", nil))
//}
