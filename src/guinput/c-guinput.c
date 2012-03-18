/*
  ukey - uinput shim for go510
  Copyright 2012 Andrew Sellers <andrew@andrewcsellers.com>

  This program is licensed under the Apache 2.0 license or the 
  GNU GPLv2, to the user's benefit.

*/

#include <fcntl.h>
#include <linux/input.h>
#include <linux/uinput.h>
//TODO: figure out if the next two includes are required
#include <unistd.h>
#include <stdlib.h>
#include <stdio.h>
#include <string.h>
#include "guinput.h"
//#include "_cgo_export.h"



int UkeyDeviceCreate(){
  int fd;
  struct uinput_user_dev keydev;
  int i;

  fd = open("/dev/uinput", O_WRONLY | O_NONBLOCK);
  if (fd < 0){
    return -1;
  }

  //I'm a keyboard :>)
  if(ioctl(fd,UI_SET_EVBIT, EV_KEY) < 0){
    return -1;
  }

  //we'll start with ~250 keys, more in a later version of this shim
  for (i=1; i<248; i++){
    if (ioctl(fd, UI_SET_KEYBIT, i) < 0){
      return -1;
    }
  }

  //setup a device for us
  memset(&keydev, 0, sizeof(keydev));
  snprintf(keydev.name, UINPUT_MAX_NAME_SIZE, "go510 Keyboard");
  keydev.id.bustype = BUS_USB;
  keydev.id.vendor = 0x1;
  keydev.id.product = 0x1;
  keydev.id.version = 1;

  if (write(fd, &keydev, sizeof(keydev)) < 0){
    return -1;
  }
  if (ioctl(fd, UI_DEV_CREATE) < 0){
    return -1;
  }

  sleep(2);

  return fd;
}

int UkeyDeviceDestroy(int ukey_fd){
  if(ioctl(ukey_fd, UI_DEV_DESTROY) < 0){
    return -1;
  }
  return 0;
}

int UkeyKeyDown(int ukey_fd, int key_val){
  struct input_event ev;

  memset(&ev, 0, sizeof(struct input_event));
  gettimeofday(&ev.time, 0);
  ev.type = EV_KEY;
  ev.code = key_val;
  ev.value = 1;

  if (write(ukey_fd, &ev, sizeof(struct input_event)) < 0){
    return -1;
  }

  ev.type = EV_SYN;
  ev.code = SYN_REPORT;
  ev.value = 0;
  if(write(ukey_fd, &ev, sizeof(struct input_event)) < 0){
    return -1;
  }

  return 0;
}

int UkeyKeyUp(int ukey_fd, int32_t key_val){
  struct input_event ev;

  memset(&ev, 0, sizeof(struct input_event));
  gettimeofday(&ev.time, 0);
  ev.type = EV_KEY;
  ev.code = key_val;
  ev.value = 0;

  if (write(ukey_fd, &ev, sizeof(struct input_event)) < 0){
    return -1;
  }

  ev.type = EV_SYN;
  ev.code = SYN_REPORT;
  ev.value = 0;
  if(write(ukey_fd, &ev, sizeof(struct input_event)) < 0){
    return -1;
  }

  return 0;
}
