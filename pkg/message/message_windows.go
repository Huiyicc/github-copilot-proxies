//go:build windows

package message

import (
	"syscall"
	"unsafe"
)

const (
	MB_OK              = 0x00000000
	MB_ICONINFORMATION = 0x00000040
)

var (
	user32      = syscall.NewLazyDLL("user32.dll")
	messageBoxW = user32.NewProc("MessageBoxW")
)

func ShowAppLaunchMessage() {
	titlePtr, _ := syscall.UTF16PtrFromString("运行成功")
	textPtr, _ := syscall.UTF16PtrFromString("服务已经启动, GUI的程序将会以后台方式运行, 如需关闭请手动结束进程.")
	messageBoxW.Call(0, uintptr(unsafe.Pointer(textPtr)), uintptr(unsafe.Pointer(titlePtr)), MB_OK|MB_ICONINFORMATION)
}
