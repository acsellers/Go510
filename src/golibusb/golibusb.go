package golibusb
/*
 #cgo LDFLAGS: -lusb-1.0
 #include <libusb-1.0/libusb.h>
 */
import "C"

import (
  "fmt"
  "log"
  "guinput"
)

var (
  KEYMAP map[uint8] uint8
  SPECDUO map[uint8] uint8
  FAKEMACRO map[uint8] uint8
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
  current_keys := make([]uint8, 8)
  key_endpoint := C.uchar(0x81)
  transferred_bytes := C.int(0)

  for {
    e := C.libusb_interrupt_transfer(handle, key_endpoint, &data_buffer[0], 8, &transferred_bytes,100)
    switch e {
    case 0:
      //you got the dataz!!
      //TODO: use a slice instead of starting at 2
      fmt.Printf("Received usb packet: %02x|%02x.%02x.%02x.%02x.%02x.%02x\n",data_buffer[1],data_buffer[2],data_buffer[3],data_buffer[4],data_buffer[5],data_buffer[6],data_buffer[7])
      for i := 2; i < 8; i++ {
        if data_buffer[i] == 0x00 && current_keys[i] != 0 {
          output <-guinput.KeyEvent{guinput.KEY_UP_CODE, current_keys[i]}
          current_keys[i] = 0x00
        }
        if data_buffer[i] != 0x00 {
          log.Printf("Key pressed:%02x",data_buffer[i])
          ucode := KEYMAP[uint8(data_buffer[i])]
          if ucode != current_keys[i] && current_keys[i] != 0 {
            output <-guinput.KeyEvent{guinput.KEY_UP_CODE, current_keys[i]}
          }
          current_keys[i] = ucode
          output <- guinput.KeyEvent{guinput.KEY_DOWN_CODE, current_keys[i]}
        }
      }
      if uint8(data_buffer[0]) != current_keys[0] {
        ParseModifiers(uint8(data_buffer[0]),current_keys[0],output)
        current_keys[0]=uint8(data_buffer[0])
      }
    case -7:
      //nothing happened
    default:
      //augh, panic
      log.Printf("Normal libusb goroutine encountered %d", e)
    }
  }
}

func ParseModifiers(pressed_keys uint8, previous_keys uint8, output chan guinput.KeyEvent) {
  if pressed_keys | previous_keys > previous_keys { //additional keys were pressed

    if pressed_keys & 0x10 != 0{
      output <- guinput.KeyEvent{guinput.KEY_DOWN_CODE, guinput.KEY_RIGHTCTRL}
    }
    if pressed_keys & 0x20 != 0{
      output <- guinput.KeyEvent{guinput.KEY_DOWN_CODE, guinput.KEY_RIGHTSHIFT}
    }
    if pressed_keys & 0x40 != 0{
      output <- guinput.KeyEvent{guinput.KEY_DOWN_CODE, guinput.KEY_RIGHTALT}
    }
    if pressed_keys & 0x80 != 0{
      output <- guinput.KeyEvent{guinput.KEY_DOWN_CODE, guinput.KEY_RIGHTMETA}
    }
    if pressed_keys & 0x01 != 0{
      output <- guinput.KeyEvent{guinput.KEY_DOWN_CODE, guinput.KEY_LEFTCTRL}
    }
    if pressed_keys & 0x02 != 0{
      output <- guinput.KeyEvent{guinput.KEY_DOWN_CODE, guinput.KEY_LEFTSHIFT}
    }
    if pressed_keys & 0x04 != 0{
      output <- guinput.KeyEvent{guinput.KEY_DOWN_CODE, guinput.KEY_LEFTALT}
    }
    if pressed_keys & 0x08 != 0{
      output <- guinput.KeyEvent{guinput.KEY_DOWN_CODE, guinput.KEY_LEFTMETA}
    }
  }


  if ^pressed_keys & previous_keys != 0 { //keys were released
    drop_keys := previous_keys^pressed_keys&pressed_keys
    log.Print("Lost a key")

    if drop_keys & 0x10 != 0{
      output <- guinput.KeyEvent{guinput.KEY_UP_CODE, guinput.KEY_RIGHTCTRL}
    }
    if drop_keys & 0x20 != 0{
      output <- guinput.KeyEvent{guinput.KEY_UP_CODE, guinput.KEY_RIGHTSHIFT}
    }
    if drop_keys & 0x40 != 0{
      output <- guinput.KeyEvent{guinput.KEY_UP_CODE, guinput.KEY_RIGHTALT}
    }
    if drop_keys & 0x80 != 0{
      output <- guinput.KeyEvent{guinput.KEY_UP_CODE, guinput.KEY_RIGHTMETA}
    }
    if drop_keys & 0x01 != 0{
      output <- guinput.KeyEvent{guinput.KEY_UP_CODE, guinput.KEY_LEFTCTRL}
    }
    if drop_keys & 0x02 != 0{
      output <- guinput.KeyEvent{guinput.KEY_UP_CODE, guinput.KEY_LEFTSHIFT}
    }
    if drop_keys & 0x04 != 0{
      output <- guinput.KeyEvent{guinput.KEY_UP_CODE, guinput.KEY_LEFTALT}
    }
    if drop_keys & 0x08 != 0{
      output <- guinput.KeyEvent{guinput.KEY_UP_CODE, guinput.KEY_LEFTMETA}
    }
  }
}

func SpecialKeyMonitor(handle *C.struct_libusb_device_handle, output chan guinput.KeyEvent) {

  data_buffer := make([]C.uchar, 512)
  key_endpoint := C.uchar(0x82)
  transferred_bytes := C.int(0)
  pressed_duos := uint8(0)

  for {
    e := C.libusb_interrupt_transfer(handle, key_endpoint, &data_buffer[0], 512, &transferred_bytes, 10)
    switch e {
    case 0:
      //we did it
      log.Print("Special Key data received")

      //the media buttons and dials in the top uh right
      //THey transfer 2 bytes and the first byte is 2, so duos
      if transferred_bytes == 2 && data_buffer[0] == 0x02 {
        ParseDuos(uint8(data_buffer[1]),pressed_duos,output)
        pressed_duos = uint8(data_buffer[1])
      }
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

func ParseDuos(pressed_keys uint8, previous_keys uint8, output chan guinput.KeyEvent) {
  if pressed_keys | previous_keys > previous_keys { //additional keys were pressed
    if pressed_keys & 0x10 != 0{
      output <- guinput.KeyEvent{guinput.KEY_DOWN_CODE, guinput.KEY_MUTE}
    }
    if pressed_keys & 0x20 != 0{
      output <- guinput.KeyEvent{guinput.KEY_DOWN_CODE, guinput.KEY_VOLUMEUP}
    }
    if pressed_keys & 0x40 != 0{
      output <- guinput.KeyEvent{guinput.KEY_DOWN_CODE, guinput.KEY_VOLUMEDOWN}
    }
    if pressed_keys & 0x01 != 0{
      output <- guinput.KeyEvent{guinput.KEY_DOWN_CODE, guinput.KEY_NEXTSONG}
    }
    if pressed_keys & 0x02 != 0{
      output <- guinput.KeyEvent{guinput.KEY_DOWN_CODE, guinput.KEY_PREVIOUSSONG}
    }
    if pressed_keys & 0x04 != 0{
      output <- guinput.KeyEvent{guinput.KEY_DOWN_CODE, guinput.KEY_STOP}
    }
    if pressed_keys & 0x08 != 0{
      output <- guinput.KeyEvent{guinput.KEY_DOWN_CODE, guinput.KEY_PLAYPAUSE}
    }
  }


  if ^pressed_keys & previous_keys != 0 { //keys were released
    drop_keys := previous_keys^pressed_keys&pressed_keys
    log.Print("Lost a key")

    if drop_keys & 0x10 != 0{
      output <- guinput.KeyEvent{guinput.KEY_UP_CODE, guinput.KEY_MUTE}
    }
    if drop_keys & 0x20 != 0{
      output <- guinput.KeyEvent{guinput.KEY_UP_CODE, guinput.KEY_VOLUMEUP}
    }
    if drop_keys & 0x40 != 0{
      output <- guinput.KeyEvent{guinput.KEY_UP_CODE, guinput.KEY_VOLUMEDOWN}
    }
    if drop_keys & 0x01 != 0{
      output <- guinput.KeyEvent{guinput.KEY_UP_CODE, guinput.KEY_NEXTSONG}
    }
    if drop_keys & 0x02 != 0{
      output <- guinput.KeyEvent{guinput.KEY_UP_CODE, guinput.KEY_PREVIOUSSONG}
    }
    if drop_keys & 0x04 != 0{
      output <- guinput.KeyEvent{guinput.KEY_UP_CODE, guinput.KEY_STOP}
    }
    if drop_keys & 0x08 != 0{
      output <- guinput.KeyEvent{guinput.KEY_UP_CODE, guinput.KEY_PLAYPAUSE}
    }
  }
}
/*
func FilterMacroKeys(normal_stream chan uint8, special_stream chan uint8, output chan guinput.KeyEvent) {
  norm1,norm2,spec1,spec2 := uint8(0),uint8(0),uint8(0),uint8(0)

  norm1 <- normal_stream

  spec1 <- special_stream

  norm2 <- normal_stream

  spec2 <- special_stream

}
*/
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
  KEYMAP = map[uint8] uint8  {
    0x29 : guinput.KEY_ESC,
    0x35 : guinput.KEY_GRAVE,
    0x1e : guinput.KEY_1,
    0x1f : guinput.KEY_2,
    0x20 : guinput.KEY_3,
    0x21 : guinput.KEY_4,
    0x22 : guinput.KEY_5,
    0x23 : guinput.KEY_6,
    0x24 : guinput.KEY_7,
    0x25 : guinput.KEY_8,
    0x26 : guinput.KEY_9,
    0x27 : guinput.KEY_0,
    0x2d : guinput.KEY_MINUS,
    0x2e : guinput.KEY_EQUAL,
    0x2a : guinput.KEY_BACKSPACE,
    0x2b : guinput.KEY_TAB,
    0x14 : guinput.KEY_Q,
    0x1a : guinput.KEY_W,
    0x08 : guinput.KEY_E,
    0x15 : guinput.KEY_R,
    0x17 : guinput.KEY_T,
    0x1c : guinput.KEY_Y,
    0x18 : guinput.KEY_U,
    0x0c : guinput.KEY_I,
    0x12 : guinput.KEY_O,
    0x13 : guinput.KEY_P,
    0x2f : guinput.KEY_LEFTBRACE,
    0x30 : guinput.KEY_RIGHTBRACE,
    0x31 : guinput.KEY_BACKSLASH,
    0x39 : guinput.KEY_CAPSLOCK,
    0x04 : guinput.KEY_A,
    0x16 : guinput.KEY_S,
    0x07 : guinput.KEY_D,
    0x09 : guinput.KEY_F,
    0x0a : guinput.KEY_G,
    0x0b : guinput.KEY_H,
    0x0d : guinput.KEY_J,
    0x0e : guinput.KEY_K,
    0x0f : guinput.KEY_L,
    0x33 : guinput.KEY_SEMICOLON,
    0x34 : guinput.KEY_APOSTROPHE,
    0x28 : guinput.KEY_ENTER,
    0x1d : guinput.KEY_Z,
    0x1b : guinput.KEY_X,
    0x06 : guinput.KEY_C,
    0x19 : guinput.KEY_V,
    0x05 : guinput.KEY_B,
    0x11 : guinput.KEY_N,
    0x10 : guinput.KEY_M,
    0x36 : guinput.KEY_COMMA,
    0x37 : guinput.KEY_DOT,
    0x38 : guinput.KEY_SLASH,
    0x2c : guinput.KEY_SPACE,
    0x49 : guinput.KEY_INSERT,
    0x4c : guinput.KEY_DELETE,
    0x4a : guinput.KEY_HOME,
    0x4d : guinput.KEY_END,
    0x4b : guinput.KEY_PAGEUP,
    0x4e : guinput.KEY_PAGEDOWN,
    0x46 : guinput.KEY_SYSRQ,
    0x47 : guinput.KEY_SCROLLLOCK,
    0x48 : guinput.KEY_PAUSE,
    0x52 : guinput.KEY_UP,
    0x51 : guinput.KEY_DOWN,
    0x50 : guinput.KEY_LEFT,
    0x4f : guinput.KEY_RIGHT,
    0x53 : guinput.KEY_NUMLOCK,
    0x54 : guinput.KEY_KPSLASH,
    0x55 : guinput.KEY_KPASTERISK,
    0x56 : guinput.KEY_KPMINUS,
    0x57 : guinput.KEY_KPPLUS,
    0x63 : guinput.KEY_KPDOT,
    0x58 : guinput.KEY_KPENTER,
    0x62 : guinput.KEY_KP0,
    0x59 : guinput.KEY_KP1,
    0x5a : guinput.KEY_KP2,
    0x5b : guinput.KEY_KP3,
    0x5c : guinput.KEY_KP4,
    0x5d : guinput.KEY_KP5,
    0x5e : guinput.KEY_KP6,
    0x5f : guinput.KEY_KP7,
    0x60 : guinput.KEY_KP8,
    0x61 : guinput.KEY_KP9,
    0x3a : guinput.KEY_F1,
    0x3b : guinput.KEY_F2,
    0x3c : guinput.KEY_F3,
    0x3d : guinput.KEY_F4,
    0x3e : guinput.KEY_F5,
    0x3f : guinput.KEY_F6,
    0x40 : guinput.KEY_F7,
    0x41 : guinput.KEY_F8,
    0x42 : guinput.KEY_F9,
    0x43 : guinput.KEY_F10,
    0x44 : guinput.KEY_F11,
    0x45 : guinput.KEY_F12,
  }

  SPECDUO = map[uint8] uint8  {
    0x40 : guinput.KEY_VOLUMEDOWN,
    0x20 : guinput.KEY_VOLUMEUP,
    0x10 : guinput.KEY_MUTE,
    0x08 : guinput.KEY_PLAYPAUSE,
    0x04 : guinput.KEY_STOP,
    0x02 : guinput.KEY_PREVIOUSSONG,
    0x01 : guinput.KEY_NEXTSONG,
  }

  //These are the scan codes that the macro keys will imitate mapped back to macro num
  FAKEMACRO= map[uint8] uint8 {
    0x3a : 1,
    0x3b : 2,
    0x3c : 3,
    0x3d : 4,
    0x3e : 5,
    0x3f : 6,
    0x40 : 7,
    0x41 : 8,
    0x42 : 9,
    0x43 : 10,
    0x44 : 11,
    0x45 : 12,
    0x1e : 13,
    0x1f : 14,
    0x20 : 15,
    0x21 : 16,
    0x22 : 17,
    0x23 : 18,
  }
}
