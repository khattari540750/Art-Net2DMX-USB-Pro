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
	// Enttec DMXUSB Proのパケット構造
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

	// コンテキストとキャンセル関数を作成
	ctx, cancel := context.WithCancel(context.Background())

	// シグナルハンドリングのセットアップ
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	// 終了処理
	cleanup := func() {
		log.Println("アプリケーションを終了しています...")
		cancel() // goroutineを停止
		log.Println("クリーンアップ完了")
	}

	// シグナル監視goroutine
	go func() {
		<-sigCh
		cleanup()
		os.Exit(0)
	}()

	statusLabel := widget.NewLabel("デバイス未接続")
	artnetLabel := widget.NewLabel("Artnet未受信")

	// ユニバース設定用のウィジェット
	universeEntry := widget.NewEntry()
	universeEntry.SetText("0")
	universeEntry.SetPlaceHolder("ユニバース番号 (0-32767)")

	// ユニバース番号を取得する関数
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
			widget.NewLabel("ユニバース:"),
			universeEntry,
		),
		widget.NewSeparator(),
		statusLabel,
		artnetLabel,
	))

	// ウィンドウ閉じるときのコールバック
	w.SetCloseIntercept(func() {
		cleanup()
		w.Close()
	})

	// バックグラウンドでArtnet受信とDMX送信のみ実行（UI更新なし）
	go func() {
		addr := net.UDPAddr{
			Port: 6455,
			IP:   net.ParseIP("0.0.0.0"),
		}
		conn, err := net.ListenUDP("udp", &addr)
		if err != nil {
			log.Println("Artnet受信エラー:", err)
			return
		}
		defer func() {
			conn.Close()
			log.Println("UDPポートを閉じました")
		}()

		buf := make([]byte, 1024)
		for {
			select {
			case <-ctx.Done():
				log.Println("Artnet受信goroutineを停止します")
				return
			default:
				// タイムアウト付きの読み取り
				err := conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
				if err != nil {
					continue
				}

				n, remoteAddr, err := conn.ReadFromUDP(buf)
				if err != nil {
					// タイムアウトエラーは無視
					continue
				}
				pkt, err := packet.Unmarshal(buf[:n])
				if err == nil {
					if dmx, ok := pkt.(*packet.ArtDMXPacket); ok {
						// 指定されたユニバースのみ処理
						targetUniverse := getTargetUniverse()
						// ArtDMXPacketのUniverseフィールドをチェック（フィールド名を確認）
						packetUniverse := uint16(dmx.SubUni) + uint16(dmx.Net)*256 // Net + SubUni でユニバース番号を計算
						if packetUniverse == targetUniverse {
							log.Printf("Artnet受信: %s, Universe: %d, Ch数: %d",
								remoteAddr.String(), packetUniverse, len(dmx.Data))
							// DMXUSB Proへ送信
							err := sendDMXToUSBPro("/dev/tty.usbserial", dmx.Data[:dmx.Length])
							if err != nil {
								log.Println("DMXUSB Pro送信エラー:", err)
							}
						}
					}
				}
			}
		}
	}()

	// 初期状態確認
	port := "/dev/tty.usbserial"
	c := &serial.Config{Name: port, Baud: 57600}
	s, err := serial.OpenPort(c)
	if err == nil {
		statusLabel.SetText("デバイス接続済み: " + port)
		s.Close()
	} else {
		statusLabel.SetText("デバイス未接続")
	}

	w.ShowAndRun()
}
