package main

import (
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"
	"time"

	"github.com/blackjack/webcam"
	"github.com/godbus/dbus/v5"
)

// dbus-monitor --system interface=org.freedesktop.login1.Session

const (
	dev       = "/dev/video0"
	namespace = "/org/freedesktop/login1"
	name      = "org.freedesktop.login1.Session.Unlock"
	photopath = "/var/log/paparazzi"
)

var (
	format = fourcc('M', 'J', 'P', 'G')
)

func fourcc(a, b, c, d rune) webcam.PixelFormat {
	if a < 0 {
		panic("a < 0")
	}

	if b < 0 {
		panic("b < 0")
	}

	if c < 0 {
		panic("c < 0")
	}

	if d < 0 {
		panic("d < 0")
	}

	if a > 255 {
		panic("a > 255")
	}

	if b > 255 {
		panic("b > 255")
	}

	if c > 255 {
		panic("c > 255")
	}

	if d > 255 {
		panic("d > 255")
	}

	return webcam.PixelFormat(a | b<<8 | c<<16 | d<<24)
}

func capture() {
	cam, err := webcam.Open(dev)
	if err != nil {
		fmt.Println(err)
		return
	}

	cam.GetSupportedFrameSizes(webcam.PixelFormat(1196444237))

	var size webcam.FrameSize

	sizes := cam.GetSupportedFrameSizes(format)
	for _, s := range sizes {
		if s.MaxHeight*s.MaxWidth > size.MaxHeight*size.MaxWidth {
			size = s
		}
	}

	_, _, _, err = cam.SetImageFormat(format, size.MaxWidth, size.MaxHeight)

	if err != nil {
		fmt.Println("Error setting format", err)

		return
	}

	err = cam.StartStreaming()
	if err != nil {
		fmt.Println("Error streaming", err)

		return
	}

	for i := 0; i < 20; i++ {
		time.Sleep(100 * time.Millisecond)
		frame, err := cam.ReadFrame()
		if err != nil {
			continue
		}

		filename := fmt.Sprintf("login-%s.jpg", time.Now().Format(time.RFC3339))
		fullpath := path.Join(photopath, filename)
		_ = os.WriteFile(fullpath, frame, 0644)

		break
	}

	cam.StopStreaming()
	cam.Close()
}

func main() {
	conn, err := dbus.ConnectSystemBus()
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to connect to session bus:", err)
		os.Exit(1)
	}

	defer conn.Close()

	err = conn.AddMatchSignal(dbus.WithMatchPathNamespace(namespace))
	if err != nil {
		panic(err)
	}

	c := make(chan *dbus.Signal, 10)
	conn.Signal(c)

	// Capture CTRL+C
	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

		<-sig
		conn.RemoveSignal(c)
		close(c)
		fmt.Printf("\r")
	}()

	for signal := range c {
		if signal.Name == name {
			capture()
		}
	}
}
