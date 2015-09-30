package main

/*
  #cgo LDFLAGS: -lusb-1.0
  #include <libusb-1.0/libusb.h>
*/
import "C"
import "errors"
import "flag"

import "log"
import "image"
import "image/color"
import _ "image/png"
import "math/rand"
import "os"
import "os/signal"
import "runtime"
import "time"

//import "os/signal"

var keyboardHandle *C.struct_libusb_device_handle
var Connected bool

func init() {
	Military = *flag.Bool("military", true, "Use military time on the clock")
}

func main() {
	runtime.LockOSThread()
	rand.Seed(int64(time.Now().UnixNano()))
	C.libusb_init(nil)

	StartLibUsb()

	clr := color.RGBA{0xff, 0, 0, 0}
	SetColor(clr)
	EndLibUsb()

	go StartClock()
	signals := make(chan os.Signal)
	signal.Notify(signals)
	<-signals

	C.libusb_exit(nil)
}

func SetColor(clr color.Color) {
	command_buffer := [4]C.uchar{0x05, 0x00, 0x00, 0x00}

	r, g, b, _ := clr.RGBA()
	command_buffer[1] = C.uchar(uint8(r))
	command_buffer[2] = C.uchar(uint8(g))
	command_buffer[3] = C.uchar(uint8(b))

	C.libusb_control_transfer(keyboardHandle, 33, 9, 0x305, 1, &command_buffer[0], 0x4, 1000)
}

func SetLCD(pix *image.Gray) {
	if !Connected {
		if err := StartLibUsb(); err != nil {
			return
		}
		defer EndLibUsb()
	}
	var img [992]C.uchar
	imgOffset := 32
	var curr, row uint8

	for offset := 0; offset < 5; offset++ {
		for col := 0; col < 160; col++ {
			curr = 0
			for row = 0; row < 8; row++ {
				if pix.Pix[(offset*8+int(row)-pix.Rect.Min.Y)*pix.Stride+(col-pix.Rect.Min.X)]>>4 > 0 {
					curr += 1 << row
				}
			}
			img[imgOffset] = C.uchar(curr)
			imgOffset++
		}
	}
	for col := 0; col < 160; col++ {
		curr = 0
		for row = 0; row < 3; row++ {
			if pix.Pix[(40+int(row)-pix.Rect.Min.Y)*pix.Stride+(col-pix.Rect.Min.X)]>>4 > 0 {
				curr += 1 << row
			}
		}
		img[imgOffset] = C.uchar(curr)
		imgOffset++
	}

	//magic byte from libg15 and USBTrace
	img[0] = 0x03

	transferred := C.int(0)
	C.libusb_interrupt_transfer(keyboardHandle, 0x3, &img[0], 992, &transferred, 1000)
}

func StartLibUsb() error {
	Connected = true

	key_device_handle := C.libusb_open_device_with_vid_pid(nil, 0x046d, 0xc22d)
	if key_device_handle != nil {

		if C.libusb_kernel_driver_active(key_device_handle, 1) == 1 {
			e := C.libusb_detach_kernel_driver(key_device_handle, 1)
			if e != 0 {
				log.Fatal("Can't detach kernel driver")
			}
		}

		r := C.libusb_claim_interface(key_device_handle, 1)
		if r != 0 {
			log.Fatal("Can't claim special interface")
		}

		keyboardHandle = key_device_handle

		return nil
	}
	return errors.New("could not open driver")
}

func EndLibUsb() {
	Connected = false
	C.libusb_release_interface(keyboardHandle, 1)
	C.libusb_attach_kernel_driver(keyboardHandle, 1)
}
func StartClock() {
	SetLCD(TimeNowImg())
	<-time.After(time.Duration(60-time.Now().Second()) * time.Second)
	SetLCD(TimeNowImg())
	tick := time.Tick(1 * time.Minute)
	for _ = range tick {
		SetLCD(TimeNowImg())
	}
}
