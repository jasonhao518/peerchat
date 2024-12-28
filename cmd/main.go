package main

import (
	"C"
	"flag"
	"fmt"
	"io"
	log2 "log"
	"net/http"
	"os"
	"time"

	"github.com/libp2p/go-libp2p/core/peer"

	"github.com/manishmeganathan/peerchat/protocol"
	"github.com/manishmeganathan/peerchat/src"
	"github.com/sirupsen/logrus"
)
import (
	"context"
	"encoding/json"
	"net"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"sync"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/manishmeganathan/peerchat/internal/log"
	"github.com/manishmeganathan/peerchat/pkg/xtermjs"
	"github.com/multiformats/go-multiaddr"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const figlet = `

W E L C O M E  T O
					     db                  db   
					     88                  88   
.8d888b. .d8888b. .d8888b. .d8888b. .d8888b. 88d888b. .d8888b. d8888P 
88'  '88 88ooood8 88ooood8 88'  '88 88'      88'  '88 88'  '88   88   
88.  .88 88.      88.      88       88.      88    88 88.  .88   88   
888888P' '88888P' '88888P' db       '88888P' db    db '8888888   '88P   
88                                                                    
dP     
`

var (
	cmd      *exec.Cmd
	sigs     = make(chan os.Signal, 1)
	selfExit = make(chan struct{}, 1)
)

type PodmanConnection struct {
	Name     string `json:"Name"`
	Identity string `json:"Identity"`
	URI      string `json:"URI"`
	Default  bool   `json:"Default"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for demo; restrict in production
	},
}

func init() {
	// Log as Text with color
	logrus.SetFormatter(&logrus.TextFormatter{
		ForceColors:     true,
		FullTimestamp:   true,
		TimestampFormat: time.RFC822,
	})

	// Log to stdout
	logrus.SetOutput(os.Stdout)
}

func main() {
	key := flag.String("key", "", "private key")
	port := flag.String("port", "3030", "http port.")
	ssh := flag.String("ssh", "2222", "http port.")
	socks5 := flag.String("socks5", "1082", "http port.")
	chatroom := flag.String("room", "", "chatroom to join.")
	workdir := flag.String("workdir", ".", "http port.")
	// Parse input flags
	flag.Parse()

	RunMain(C.CString(*key), C.CString(*port), C.CString(*ssh), C.CString(*socks5), C.CString(*workdir), C.CString(*chatroom))
}

//export RunMain
func RunMain(privKey *C.char, port *C.char, ssh *C.char, socks5 *C.char, workdir *C.char, room *C.char) {
	serverStr := C.GoString(privKey)
	portStr := C.GoString(port)
	sshStr := C.GoString(ssh)
	socks5Str := C.GoString(socks5)
	workingDirectory := C.GoString(workdir)
	chatroom := C.GoString(room)
	fmt.Println("Received string from C:", serverStr, portStr, sshStr, socks5Str, workingDirectory, chatroom)
	// Define input flags
	var command string

	// Determine the operating system
	switch runtime.GOOS {
	case "darwin": // macOS
		command = filepath.Join(filepath.Dir(filepath.Dir(workingDirectory)), "MacOS", "podman")
	default: // Other operating systems
		workingDirectory = "."
		command = filepath.Join(workingDirectory, "podman.exe")
	}

	// Output the command
	fmt.Println("Command:", command)

	username := flag.String("user", "", "username to use in the chatroom.")
	loglevel := flag.String("log", "", "level of logs to print.")
	discovery := flag.String("discover", "", "method to use for discovery.")
	// Parse input flags
	flag.Parse()

	// Set the log level
	switch *loglevel {
	case "panic", "PANIC":
		logrus.SetLevel(logrus.PanicLevel)
	case "fatal", "FATAL":
		logrus.SetLevel(logrus.FatalLevel)
	case "error", "ERROR":
		logrus.SetLevel(logrus.ErrorLevel)
	case "warn", "WARN":
		logrus.SetLevel(logrus.WarnLevel)
	case "info", "INFO":
		logrus.SetLevel(logrus.InfoLevel)
	case "debug", "DEBUG":
		logrus.SetLevel(logrus.DebugLevel)
	case "trace", "TRACE":
		logrus.SetLevel(logrus.TraceLevel)
	default:
		logrus.SetLevel(logrus.InfoLevel)
	}

	// Display the welcome figlet
	// fmt.Println(figlet)
	fmt.Println("The PeerChat Application is starting.")
	fmt.Println("This may take upto 30 seconds.")
	fmt.Println()

	// Create a new P2PHost
	p2phost := src.NewP2P(serverStr, command)
	logrus.Infoln("Completed P2P Setup")

	// Connect to peers with the chosen discovery method
	switch *discovery {
	case "announce":
		p2phost.AnnounceConnect()
	case "advertise":
		p2phost.AdvertiseConnect()
	default:
		p2phost.AdvertiseConnect()
	}
	logrus.Infoln("Connected to Service Peers")

	// Join the chat room
	chatapp, _ := src.JoinChatRoom(p2phost, *username, chatroom)
	logrus.Infof("Joined the '%s' chatroom as '%s'", chatapp.RoomName, chatapp.UserName)

	// Wait for network setup to complete
	time.Sleep(time.Second * 5)

	// initialise the logger
	log.Init(log.Format(conf.GetString("log-format")), log.Level(conf.GetString("log-level")))

	// debug stuff
	//command := conf.GetString("command")
	connectionErrorLimit := conf.GetInt("connection-error-limit")
	arguments := conf.GetStringSlice("arguments")
	allowedHostnames := conf.GetStringSlice("allowed-hostnames")
	keepalivePingTimeout := time.Duration(conf.GetInt("keepalive-ping-timeout")) * time.Second
	maxBufferSizeBytes := conf.GetInt("max-buffer-size-bytes")
	pathLiveness := conf.GetString("path-liveness")
	pathMetrics := conf.GetString("path-metrics")
	pathReadiness := conf.GetString("path-readiness")
	pathXTermJS := conf.GetString("path-xtermjs")
	serverAddress := conf.GetString("server-addr")
	serverPort := conf.GetInt("server-port")
	//workingDirectory := conf.GetString("workdir")
	if !path.IsAbs(workingDirectory) {
		wd, err := os.Getwd()
		if err != nil {
			message := fmt.Sprintf("failed to get working directory: %s", err)
			log.Error(message)
		}
		workingDirectory = path.Join(wd, workingDirectory)
	}
	log.Infof("working directory     : '%s'", workingDirectory)
	log.Infof("command               : '%s'", command)
	log.Infof("arguments             : ['%s']", strings.Join(arguments, "', '"))

	log.Infof("allowed hosts         : ['%s']", strings.Join(allowedHostnames, "', '"))
	log.Infof("connection error limit: %v", connectionErrorLimit)
	log.Infof("keepalive ping timeout: %v", keepalivePingTimeout)
	log.Infof("max buffer size       : %v bytes", maxBufferSizeBytes)
	log.Infof("server address        : '%s' ", serverAddress)
	log.Infof("server port           : %v", serverPort)

	log.Infof("liveness checks path  : '%s'", pathLiveness)
	log.Infof("readiness checks path : '%s'", pathReadiness)
	log.Infof("metrics endpoint path : '%s'", pathMetrics)
	log.Infof("xtermjs endpoint path : '%s'", pathXTermJS)

	// configure routing
	router := mux.NewRouter()

	// this is the endpoint for xterm.js to connect to
	xtermjsHandlerOptions := xtermjs.HandlerOpts{
		AllowedHostnames:     allowedHostnames,
		Arguments:            arguments,
		Command:              command,
		ConnectionErrorLimit: connectionErrorLimit,
		CreateLogger: func(connectionUUID string, r *http.Request) xtermjs.Logger {
			createRequestLog(r, map[string]interface{}{"connection_uuid": connectionUUID}).Infof("created logger for connection '%s'", connectionUUID)
			return createRequestLog(nil, map[string]interface{}{"connection_uuid": connectionUUID})
		},
		KeepalivePingTimeout: keepalivePingTimeout,
		MaxBufferSizeBytes:   maxBufferSizeBytes,
	}
	router.HandleFunc(pathXTermJS, xtermjs.GetHandler(xtermjsHandlerOptions))

	// readiness probe endpoint
	router.HandleFunc(pathReadiness, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	router.HandleFunc("/join", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")                            // Allow all origins
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")          // Allowed methods
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization") // Allowed headers
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read request body", http.StatusInternalServerError)
			return
		}
		defer r.Body.Close() // Ensure the body is closed
		bodyStr := string(body)
		// Create a reference to the current chatroom
		oldchatroom := chatapp

		// Create a new chatroom and join it
		newchatroom, err := src.JoinChatRoom(p2phost, "newuser", bodyStr)
		if err != nil {
			logrus.Debugf("could not change chat room - %s", err)

		}

		// Assign the new chat room to UI
		chatapp = newchatroom
		// Sleep for a second to give time for the queues to adapt
		time.Sleep(time.Second * 1)

		// Exit the old chatroom and pause for two seconds
		oldchatroom.Exit()
	})

	// liveness probe endpoint
	router.HandleFunc(pathLiveness, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	router.HandleFunc("/info", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")                            // Allow all origins
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")          // Allowed methods
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization") // Allowed headers

		w.WriteHeader(http.StatusOK)

		addrs := p2phost.Host.Addrs()
		var result []multiaddr.Multiaddr
		for _, addr := range addrs {
			// Extract the IP part from the multiaddress
			ip, err := addr.ValueForProtocol(multiaddr.P_IP4)
			if err != nil {
				// If not IPv4, try IPv6
				ip, err = addr.ValueForProtocol(multiaddr.P_IP6)
			}
			if err == nil && !net.ParseIP(ip).IsLoopback() {
				result = append(result, addr)
			}
		}
		// Convert addresses to string
		addrStrings := make([]string, len(result))
		for i, addr := range result {
			addrStrings[i] = addr.String() + "/" + p2phost.Host.ID().String()
		}
		// Encode as JSON and write response
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(addrStrings); err != nil {
			http.Error(w, "Failed to encode addresses", http.StatusInternalServerError)
		}
	})

	router.HandleFunc("/peer", func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")                            // Allow all origins
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")          // Allowed methods
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization") // Allowed headers

		// Handle preflight (OPTIONS) request
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent) // 204 No Content
			return
		} else if r.Method == "GET" {
			if len(chatapp.PeerList()) > 0 {
				jsonArray, err := json.Marshal(chatapp.PeerList())
				if err != nil {
					fmt.Println("Error:", err)
					return
				}
				fmt.Fprintln(w, string(jsonArray))
			} else {
				fmt.Fprintln(w, chatapp.PeerList())
			}

			// Print the JSON array as a string

		} else {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusInternalServerError)
				return
			}
			defer r.Body.Close() // Ensure the body is closed
			bodyStr := string(body)

			if bodyStr == "" || isIPAddress(bodyStr) {
				// switch to self when it's empty
				p2phost.Proxy.SetRemotePeer(p2phost.Host.ID())
				fmt.Fprintln(w, "reset remote peer to self", p2phost.Host.ID())
			} else {
				peerID, err := peer.Decode(bodyStr)
				if err != nil {
					http.Error(w, "Failed to read request body", http.StatusInternalServerError)
					return
				}
				// if multi address, need to connect to it at this time

				p2phost.Proxy.SetRemotePeer(peerID)
				fmt.Fprintln(w, "successfully set remote peer to", bodyStr)
			}

		}
	})

	// metrics endpoint
	router.Handle(pathMetrics, promhttp.Handler())

	// this is the endpoint for serving xterm.js assets
	depenenciesDirectory := path.Join(workingDirectory, "./node_modules")
	router.PathPrefix("/assets").Handler(http.StripPrefix("/assets", http.FileServer(http.Dir(depenenciesDirectory))))

	// this is the endpoint for the root path aka website
	publicAssetsDirectory := path.Join(workingDirectory, "./public")
	router.PathPrefix("/").Handler(http.FileServer(http.Dir(publicAssetsDirectory)))

	// start memory logging pulse
	logWithMemory := createMemoryLog()
	go func(tick *time.Ticker) {
		for {
			logWithMemory.Debug("tick")
			<-tick.C
		}
	}(time.NewTicker(time.Second * 30))

	go func() {

		// listen
		listenOnAddress := fmt.Sprintf("%s:%v", serverAddress, portStr)
		server := http.Server{
			Addr:    listenOnAddress,
			Handler: addIncomingRequestLogging(router),
		}

		log.Infof("starting server on interface:port '%s'...", listenOnAddress)
		server.ListenAndServe()
	}()

	//	go func() {
	//		if err := p2phost.Proxy.Serve("0.0.0.0:" + socks5Str); err != nil {
	//			protocol.Log.Fatal(err)
	//		}
	//	}()
	go func() {
		if err := p2phost.Proxy.ServeSsh("127.0.0.1:" + sshStr); err != nil {
			protocol.Log.Fatal(err)
		}
	}()

	wg := sync.WaitGroup{}

	// 启动代理协程
	wg.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		defer wg.Done()
		if err := Proxy(ctx, time.Duration(5*time.Second), "0.0.0.0:"+socks5Str, command); err != nil {
			log2.Println("proxy error: ", err)
		}

		log2.Println("proxy协程已退出")
		go func() {
			sigs <- syscall.SIGQUIT
		}()
	}()

	// 设置信号处理函数
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)
	wg.Add(1)
	go func() {
		defer wg.Done()
		sig := <-sigs
		log2.Println("Received signal:", sig)
		log2.Println("信号清理已经完成")
		cancel()
	}()

	wg.Wait()
	log2.Println("进程退出!")
	time.Sleep(time.Second * 4)
	selfExit <- struct{}{}
	// Create the Chat UI
	//ui := src.NewUI(chatapp)
	// Start the UI system
	//ui.Run()
}

func isIPAddress(input string) bool {
	// IPv4 regex
	ipv4Regex := `^(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.` +
		`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.` +
		`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\.` +
		`(25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)$`

	// IPv6 regex
	ipv6Regex := `^(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|` +
		`([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|` +
		`([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|` +
		`([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|` +
		`:((:[0-9a-fA-F]{1,4}){1,7}|:)|fe80:(:[0-9a-fA-F]{0,4}){0,4}%[0-9a-zA-Z]{1,}|` +
		`::(ffff(:0{1,4}){0,1}:){0,1}((25[0-5]|(2[0-4]|1{0,1}[0-9]|)[0-9])\.){3,3}` +
		`(25[0-5]|(2[0-4]|1{0,1}[0-9]|)[0-9])|([0-9a-fA-F]{1,4}:){1,4}:` +
		`((25[0-5]|(2[0-4]|1{0,1}[0-9]|)[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]|)[0-9]))$`

	ipv4Pattern := regexp.MustCompile(ipv4Regex)
	ipv6Pattern := regexp.MustCompile(ipv6Regex)

	return ipv4Pattern.MatchString(input) || ipv6Pattern.MatchString(input)
}
