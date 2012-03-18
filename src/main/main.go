package main

import (
	"guinput"
	"golibusb"
  "os/signal"
  "log"
  "os"
)

func main() {
  log.Print("Starting go510 Daemon")

  //TODO: move these into a function sometime
  uinput_control := make(chan guinput.KeyEvent)
  color_control  := make(chan byte)
  //lcd_control := make(chan golibusb.LCDFrame)

  handle := golibusb.Start(color_control, uinput_control)
  //guinput.FakeStart(uinput_control)
  guinput.Start(uinput_control)


  //wait for it... (it being Ctrl-C)
  signal_channel := make(chan os.Signal)
  signal.Notify(signal_channel)
  <-signal_channel
  golibusb.End(handle)
  kill_event :=  guinput.KeyEvent{guinput.KEY_DETACH, 0}
  uinput_control <- kill_event


}
