package main

/*
	#cgo CFLAGS: -I../install/include
	#cgo LDFLAGS: -L../install/bin -lcsecorec
	#include "../install/include/cse.h"

	int hello_world() {
		char buf1[20];
		int ret = cse_hello_world(buf1, 20);
		return ret;
	}
*/
import "C"

func cse_hello() (number int) {
	number = int(C.hello_world())
	return
}