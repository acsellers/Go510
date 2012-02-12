package guinput

// #include "guinput.h"
import "C"

import (
  "log"
  )

//TODO: move to input_constants
const (
  KEY_START = iota
  KEY_DOWN_CODE
  KEY_UP_CODE
  KEY_DETACH
  KEY_MULTI_DOWN
  KEY_MULTI_UP
)

func Scan(device C.int, input chan KeyEvent){

  event := new(KeyEvent)

  for event.Kind != KEY_DETACH {

    *event = <-input
    log.Printf("uinput event received kind %d with Key %d\n", event.Kind, event.Key)
    switch{
    case event.Kind == KEY_DOWN_CODE:
      C.UkeyKeyDown(device, C.int(event.Key))

    case event.Kind == KEY_UP_CODE:
      C.UkeyKeyUp(device, C.int(event.Key))

    case event.Kind == KEY_DETACH:
      C.UkeyDeviceDestroy(device)
    }
  }

  log.Print("Detaching uinput system\n")
}

func Start(input chan KeyEvent){
  log.Print("Starting uinput system\n")

  //The actual creation happens in the c-guinput.c file
  device := C.UkeyDeviceCreate()
  if device == -1 {
    log.Fatal("Could not create uinput interface\n")
  }

  //Start goroutine to watch the input channel for events
  go Scan(device, input)
}


func End(input chan KeyEvent){

  //tell the watcher goroutine to die, reverse of golibusb
  killer := KeyEvent{Kind: KEY_DETACH, Key: 0}
  input <- killer
}
