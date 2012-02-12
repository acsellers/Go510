package golibusb

// #cgo LDFLAGS: -lusb-1.0
// #include <libusb-1.0/libusb.h>
import "C"

import (
  "fmt"
  "log"
  "guinput"
)

var (
  KEYMAP map[int] int
)

func ColorChange(handle *C.struct_libusb_device_handle, color chan byte) {
  command_buffer := [4]C.uchar{0x05,0x00,0x00,0x00}

  for {
    command_buffer[1] = C.uchar(<-color)
    command_buffer[2] = C.uchar(<-color)
    command_buffer[3] = C.uchar(<-color)

    r := C.libusb_control_transfer(handle,33, 9, 0x305, 1, &command_buffer[0], 0x4, 1000)
    log.Printf("Color Change Response %d\n",r)
  }
}

//Scan for keys on the normal interface
func NormalKeyMonitor(handle *C.struct_libusb_device_handle, output chan guinput.KeyEvent) {
  //libusb will drop data into data_buffer, our endpoint id is 0x81 (from lsusb)
  data_buffer := make([]C.uchar, 8)
  key_endpoint := C.uchar(0x81)
  transferred_bytes := C.int(0)

  for {
    e := C.libusb_interrupt_transfer(handle, key_endpoint, &data_buffer[0], 8, &transferred_bytes, 10)
    switch e {
    case 0:
      //you got the dataz!!
      for i := 0; i < int(transferred_bytes); i++ {
        fmt.Printf("%02x\n",data_buffer[i])
      }
    case -7:
      //nothing happened
    default:
      //augh, panic
      log.Print("Normal libusb goroutine encountered %d", e)
    }
  }
}

func SpecialKeyMonitor(handle *C.struct_libusb_device_handle, output chan guinput.KeyEvent) {

  data_buffer := make([]C.uchar, 512)
  key_endpoint := C.uchar(0x82)
  transferred_bytes := C.int(0)

  for {
    e := C.libusb_interrupt_transfer(handle, key_endpoint, &data_buffer[0], 512, &transferred_bytes, 10)
    switch e {
    case 0:
      //we did it
      log.Print("Special Key data received")
      for i := 0; i < int(transferred_bytes); i++ {
        fmt.Printf("%02x\n",data_buffer[i])
      }
    case -7:
      //nothing happened
    default:
      //augh, panic
      fmt.Println("Augh panic")
    }
  }

}


func Start(color chan byte, output chan guinput.KeyEvent)(*C.struct_libusb_device_handle){

	C.libusb_init(nil)
  fmt.Println(KEYMAP[0x21])

  key_device_handle := C.libusb_open_device_with_vid_pid(nil,0x046d,0xc22d)

  //detach any necessary kernel drivers from MY keyboard
  if C.libusb_kernel_driver_active(key_device_handle, 0) == 1 {
    log.Print("kernel driver active on main interface")
    e := C.libusb_detach_kernel_driver(key_device_handle, 0)
    if e != 0 {
      log.Fatal("Can't detach kernel driver")
    }
  }
  if C.libusb_kernel_driver_active(key_device_handle, 1) == 1 {
    fmt.Println("kernel driver active")
    e := C.libusb_detach_kernel_driver(key_device_handle, 1)
    if e != 0 {
      log.Fatal("Can't detach kernel driver")
    }
  }

  //Claim the interfaces we'll be listening on
  r := C.libusb_claim_interface(key_device_handle, 0)
  if r != 0 {
    log.Fatal("Can't claim main interface")
  }
  r = C.libusb_claim_interface(key_device_handle, 1)
  if r != 0 {
    log.Fatal("Can't claim special interface")
  }

  log.Print("Starting libusb goroutines\n")

  go ColorChange(key_device_handle, color)
  //go LCDChange(key_device_handle, lcd)
  go NormalKeyMonitor(key_device_handle, output)
  go SpecialKeyMonitor(key_device_handle, output)

  log.Print("Libusb Goroutines started\n")
  return key_device_handle
}

func End(handle *C.struct_libusb_device_handle) {

  log.Print("Libusb goroutine exiting\n")

  C.libusb_release_interface(handle, 0)
  C.libusb_release_interface(handle, 1)
  C.libusb_attach_kernel_driver(handle, 0)
  C.libusb_attach_kernel_driver(handle, 1)
	C.libusb_exit(nil)

  log.Print("Libusb goroutine exited\n")
}

func init(){
  KEYMAP = map[int] int  {
    0x21 : 123,
    0x23 : 1234,
  }
}
