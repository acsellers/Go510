Go510 Keyboard Driver
=====================

## OUTDATED: See a rebuilt version at github.com/acsellers/G510
## Seriously this is 4 years old and pre-dates Go 1.0.

Dependencies
------------

 * Go-gb build system
 * Linux (tested on Ubuntu
 * uinput kernel module loaded


### Finished
 * Turning usb packets into keystrokes 
 * Taking over the keyboard

### Short Term Goals
 * Make LED's for Num-Lock etc. perform correctly
 * Expose Color to be set by the command line
 * Upstart daemonization

### Longer Term Goals
 * Expose Color/Macro to socket setting
 * Build simple Python settings app

### Final Goals
 * LCD protocol over socket (I can draw to it now just not in a easy way)
 * Multiple Apps that use the LCD


Note: go1 branch has a version that build on recent weekly go versions
that doesn't require go-gb.
