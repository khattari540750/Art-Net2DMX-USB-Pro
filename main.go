package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/jsimonetti/go-artnet/packet"
	"github.com/tarm/serial"
)

// Send DMX data to Enttec DMXUSB Pro
func sendDMXToUSBPro(portName string, dmxData []byte) error {
	// Enttec DMXUSB Pro packet structure
	// [0]=0x7E, [1]=Label(6), [2]=LenLo, [3]=LenHi, [4]=StartCode(0), [5:]=DMX, [n]=0xE7
	if len(dmxData) > 512 {
		dmxData = dmxData[:512]
	}
	packetLen := len(dmxData) + 1 // +1 for StartCode
	pkt := make([]byte, 0, packetLen+5)
	pkt = append(pkt, 0x7E)
	pkt = append(pkt, 6) // Label: 6 (Send DMX Packet)
	pkt = append(pkt, byte(packetLen&0xFF), byte((packetLen>>8)&0xFF))
	pkt = append(pkt, 0x00) // DMX StartCode
	pkt = append(pkt, dmxData...)
	pkt = append(pkt, 0xE7)

	c := &serial.Config{Name: portName, Baud: 57600}
	s, err := serial.OpenPort(c)
	if err != nil {
		return err
	}
	defer s.Close()
	_, err = s.Write(pkt)
	return err
}

func main() {
	myApp := app.New()
	w := myApp.NewWindow("Artnet to DMXUSB Pro")
	w.Resize(fyne.NewSize(400, 200))

	// Create context and cancel function
	ctx, cancel := context.WithCancel(context.Background())

	// Setup signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// Cleanup function
	cleanup := func() {
		log.Println("Shutting down application...")
		cancel() // Stop goroutines
		log.Println("Cleanup completed")
	}

	// Signal monitoring goroutine
	go func() {
		<-sigCh
		cleanup()
		os.Exit(0)
	}()

	statusLabel := widget.NewLabel("Device not connected")
	artnetLabel := widget.NewLabel("Art-Net not received")

	// Widgets for universe setting
	universeEntry := widget.NewEntry()
	universeEntry.SetText("0")
	universeEntry.SetPlaceHolder("Universe number (0-32767)")

	// Function to get universe number
	getTargetUniverse := func() uint16 {
		universeText := universeEntry.Text
		if universeText == "" {
			return 0
		}
		universe, err := strconv.ParseUint(universeText, 10, 16)
		if err != nil || universe > 32767 {
			return 0
		}
		return uint16(universe)
	}

	w.SetContent(container.NewVBox(
		widget.NewLabel("ArtNet to Enttec DMXUSB Pro"),
		container.NewHBox(
			widget.NewLabel("Universe:"),
			universeEntry,
		),
		widget.NewSeparator(),
		statusLabel,
		artnetLabel,
	))

	// Window close callback
	w.SetCloseIntercept(func() {
		cleanup()
		w.Close()
	})

	// Run Art-Net reception and DMX transmission in background only (no UI updates)
	go func() {
		addr := net.UDPAddr{
			Port: 6455,
			IP:   net.ParseIP("0.0.0.0"),
		}
		conn, err := net.ListenUDP("udp", &addr)
		if err != nil {
			log.Println("Art-Net reception error:", err)
			return
		}
		defer func() {
			conn.Close()
			log.Println("UDP port closed")
		}()

		buf := make([]byte, 1024)
		for {
			select {
			case <-ctx.Done():
				log.Println("Stopping Art-Net reception goroutine")
				return
			default:
				// Read with timeout
				err := conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
				if err != nil {
					continue
				}

				n, remoteAddr, err := conn.ReadFromUDP(buf)
				if err != nil {
					// Ignore timeout errors
					continue
				}
				pkt, err := packet.Unmarshal(buf[:n])
				if err == nil {
					if dmx, ok := pkt.(*packet.ArtDMXPacket); ok {
						// Process only specified universe
						targetUniverse := getTargetUniverse()
						// Check ArtDMXPacket Universe field (verify field name)
						packetUniverse := uint16(dmx.SubUni) + uint16(dmx.Net)*256 // Calculate universe number with Net + SubUni
						if packetUniverse == targetUniverse {
							log.Printf("Art-Net received: %s, Universe: %d, Channels: %d",
								remoteAddr.String(), packetUniverse, len(dmx.Data))
							// Send to DMXUSB Pro
							err := sendDMXToUSBPro("/dev/tty.usbserial", dmx.Data[:dmx.Length])
							if err != nil {
								log.Println("DMXUSB Pro transmission error:", err)
							}
						}
					}
				}
			}
		}
	}()

	// Initial status check
	port := "/dev/tty.usbserial"
	c := &serial.Config{Name: port, Baud: 57600}
	s, err := serial.OpenPort(c)
	if err == nil {
		statusLabel.SetText("Device connected: " + port)
		s.Close()
	} else {
		statusLabel.SetText("Device not connected")
	}

	w.ShowAndRun()
}
