package main

import (
	"bufio"
	"encoding/json"
	"io"
	"log/slog"
	"machine"
	"net/netip"
	"sync"
	"time"

	_ "embed"

	"github.com/soypat/cyw43439"
	"github.com/soypat/cyw43439/examples/common"
	"github.com/soypat/seqs/httpx"
	"github.com/soypat/seqs/stacks"
)

const (
	connTimeout = 3 * time.Second
	maxconns    = 3
	tcpbufsize  = 2030
	hostname    = "picopanic"
	listenPort  = 80
)

var (
	//go:embed index.html
	webPage         []byte
	buttonPinNumber machine.Pin = machine.GPIO0

	// panic stuff
	mlock      sync.Mutex
	panicOcurr bool
)

var logger *slog.Logger

func init() {
	logger = slog.New(
		slog.NewTextHandler(machine.Serial, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
}

func changeLEDState(dev *cyw43439.Device, state bool) {
	if err := dev.GPIOSet(0, state); err != nil {
		logger.Error("failed to change LED state:", slog.String("err", err.Error()))
	}
}

func setupDevice() (*stacks.PortStack, *cyw43439.Device) {
	_, stack, dev, err := common.SetupWithDHCP(common.SetupConfig{
		Hostname: hostname,
		Logger:   logger,
		TCPPorts: 1,
	})
	if err != nil {
		panic("setup DHCP:" + err.Error())
	}
	// Turn LED on
	changeLEDState(dev, true)
	return stack, dev
}

func newListener(stack *stacks.PortStack) *stacks.TCPListener {
	// Start TCP server.
	listenAddr := netip.AddrPortFrom(stack.Addr(), listenPort)
	listener, err := stacks.NewTCPListener(
		stack, stacks.TCPListenerConfig{
			MaxConnections: maxconns,
			ConnTxBufSize:  tcpbufsize,
			ConnRxBufSize:  tcpbufsize,
		})

	if err != nil {
		panic("listener create:" + err.Error())
	}

	err = listener.StartListening(listenPort)
	if err != nil {
		panic("listener start:" + err.Error())
	}

	logger.Info("listening",
		slog.String("addr", "http://"+listenAddr.String()),
	)

	return listener
}

type pulseCheck struct {
	Up bool `json:"up"`
}

// always return true
func getPulseCheck() *pulseCheck {
	return &pulseCheck{
		true,
	}
}

type panicCheck struct {
	Panic     bool      `json:"panic"`
	Timestamp time.Time `json:"timestamp"`
}

func getPanicCheck() *panicCheck {
	// always retun false afer a check
	mlock.Lock()
	defer func() {
		panicOcurr = false
		return
	}()
	defer mlock.Unlock()

	if !panicOcurr {
		return &panicCheck{
			Panic: false,
		}
	}

	return &panicCheck{
		panicOcurr,
		time.Now(),
	}
}

func blinkLED(dev *cyw43439.Device, blink chan uint) {
	for {
		select {
		case n := <-blink:
			lastLedState := true
			if n == 0 {
				n = 5
			}

			for i := uint(0); i < n; i++ {
				lastLedState = !lastLedState
				changeLEDState(dev, lastLedState)
				time.Sleep(500 * time.Millisecond)
			}

			changeLEDState(dev, true)
		}
	}
}

func respondJsonOrError(response any, respWriter io.Writer, resp *httpx.ResponseHeader) {
	body, err := json.Marshal(response)
	if err != nil {
		logger.Error(
			"humidity json:", slog.String("err", err.Error()),
		)
		resp.SetStatusCode(500)
	} else {
		resp.SetContentType("application/json")
		resp.SetContentLength(len(body))
	}
	respWriter.Write(resp.Header())
	respWriter.Write(body)
}

func HTTPHandler(respWriter io.Writer, resp *httpx.ResponseHeader, req *httpx.RequestHeader) {
	uri := string(req.RequestURI())
	resp.SetConnectionClose()

	switch uri {
	case "/":
		logger.Info("Got webpage request...")
		resp.SetContentType("text/html")
		resp.SetContentLength(len(webPage))
		respWriter.Write(resp.Header())
		respWriter.Write(webPage)

	case "/panic":
		logger.Info("Got panic request...")
		respondJsonOrError(getPanicCheck(), respWriter, resp)
		return

	case "/pulse":
		logger.Info("Got pulse request...")
		respondJsonOrError(getPulseCheck(), respWriter, resp)
		return

	case "/boot":
		logger.Info("Got boot request...")
		machine.EnterBootloader()

	case "/force_panic":
		logger.Info("Got force panic request...")
		panicOcurr = true
		resp.SetStatusCode(201)
		respWriter.Write(resp.Header())
		return

	default:
		println("Path not found:", uri)
		resp.SetStatusCode(404)
		respWriter.Write(resp.Header())
	}
}

func handleConnection(listener *stacks.TCPListener, blink chan uint) { // Reuse the same buffers for each
	// connection to avoid heap allocations.
	var req httpx.RequestHeader
	var resp httpx.ResponseHeader
	buf := bufio.NewReaderSize(nil, 1024)
	for {
		conn, err := listener.Accept()
		if err != nil {
			logger.Error(
				"listener accept:", slog.String("err", err.Error()))
			time.Sleep(time.Second)
			continue
		}

		logger.Info(
			"new connection", slog.String("remote",
				conn.RemoteAddr().String()),
		)

		err = conn.SetDeadline(time.Now().Add(connTimeout))
		if err != nil {
			conn.Close()
			logger.Error(
				"conn set deadline:",
				slog.String("err", err.Error()))
			continue
		}

		buf.Reset(conn)
		err = req.Read(buf)
		if err != nil {
			logger.Error("hdr read:", slog.String("err", err.Error()))
			conn.Close()
			continue
		}
		resp.Reset()
		HTTPHandler(conn, &resp, &req)
		conn.Close()
		blink <- 5
	}
}

func handleButton(blink chan<- uint) {
	buttonPinNumber.Configure(machine.PinConfig{
		Mode: machine.PinInputPulldown,
	})

	for {
		if value := buttonPinNumber.Get(); value {

			// long press logic
			var (
				start = time.Now()
			)
			for {
				if pressed := buttonPinNumber.Get(); pressed {
					if time.Now().Sub(start) > 3*time.Second {
						slog.Info("Entering boot loader mode...")
						machine.EnterBootloader()
						return
					}
				}
				break
			}

			// short press logic
			slog.Info("Button pushed...")

			panicOcurr = true
			blink <- 3
			time.Sleep(1 * time.Second)
		}
		time.Sleep(200 * time.Millisecond)
	}
}

func main() {
	// life savier
	time.Sleep(3 * time.Second)

	stack, dev := setupDevice()
	blink := make(chan uint, 3)
	go blinkLED(dev, blink)

	go handleButton(blink)

	time.Sleep(2 * time.Second)
	listener := newListener(stack)
	go handleConnection(listener, blink)

	// init ADCH pin

	for {
		select {
		case <-time.After(1 * time.Minute):
			logger.Info("Waiting for connections...")
		}
	}

}
