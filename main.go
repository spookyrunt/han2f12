package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	whKeyboardLL        = 13
	wmKeyDown           = 0x0100
	wmSysKeyDown        = 0x0104
	vkHangul            = 0x15
	vkF12               = 0x7B
	inputKeyboard       = 1
	keyEventFKeyUp      = 0x0002
	processQueryLimited = 0x1000
)

var (
	user32   = windows.NewLazySystemDLL("user32.dll")
	kernel32 = windows.NewLazySystemDLL("kernel32.dll")

	setWindowsHookEx          = user32.NewProc("SetWindowsHookExW")
	callNextHookEx            = user32.NewProc("CallNextHookEx")
	unhookWindowsHookEx       = user32.NewProc("UnhookWindowsHookEx")
	getMessageW               = user32.NewProc("GetMessageW")
	getForegroundWindow       = user32.NewProc("GetForegroundWindow")
	getWindowThreadProcessId  = user32.NewProc("GetWindowThreadProcessId")
	sendInputProc             = user32.NewProc("SendInput")
	queryFullProcessImageName = kernel32.NewProc("QueryFullProcessImageNameW")
)

type kbdHookStruct struct {
	vkCode, scanCode, flags, time uint32
	dwExtraInfo                   uintptr
}

type msgStruct struct {
	hwnd           uintptr
	message        uint32
	wParam, lParam uintptr
	time           uint32
	ptX, ptY       int32
}

const inputSize = 40

func pressKey(vk uint16, up bool) {
	var b [inputSize]byte
	*(*uint32)(unsafe.Pointer(&b[0])) = inputKeyboard
	*(*uint16)(unsafe.Pointer(&b[8])) = vk
	if up {
		*(*uint32)(unsafe.Pointer(&b[12])) = keyEventFKeyUp
	}
	sendInputProc.Call(1, uintptr(unsafe.Pointer(&b[0])), inputSize)
}

func foregroundExe() string {
	hwnd, _, _ := getForegroundWindow.Call()
	if hwnd == 0 {
		return ""
	}
	var pid uint32
	getWindowThreadProcessId.Call(hwnd, uintptr(unsafe.Pointer(&pid)))

	h, err := windows.OpenProcess(processQueryLimited, false, pid)
	if err != nil {
		return ""
	}
	defer windows.CloseHandle(h)

	var buf [260]uint16
	n := uint32(len(buf))
	queryFullProcessImageName.Call(uintptr(h), 0, uintptr(unsafe.Pointer(&buf[0])), uintptr(unsafe.Pointer(&n)))

	path := windows.UTF16ToString(buf[:n])
	i := strings.LastIndexAny(path, `\/`)
	return strings.ToLower(path[i+1:])
}

var hookH uintptr

func hookCb(code int, wp, lp uintptr) uintptr {
	if code >= 0 && (wp == wmKeyDown || wp == wmSysKeyDown) {
		s := (*kbdHookStruct)(unsafe.Pointer(lp))
		if s.vkCode == vkHangul && foregroundExe() == "windowsterminal.exe" {
			pressKey(vkF12, false)
			pressKey(vkF12, true)
			return 1
		}
	}
	r, _, _ := callNextHookEx.Call(hookH, uintptr(code), wp, lp)
	return r
}

func installHook() error {
	cb := syscall.NewCallback(hookCb)
	hookH, _, _ = setWindowsHookEx.Call(whKeyboardLL, cb, 0, 0)
	if hookH == 0 {
		return fmt.Errorf("훅 설치 실패")
	}
	return nil
}

func main() {
	ready := make(chan error, 1)

	go func() {
		runtime.LockOSThread()

		if err := installHook(); err != nil {
			ready <- err
			return
		}
		ready <- nil

		var m msgStruct
		for {
			ret, _, _ := getMessageW.Call(uintptr(unsafe.Pointer(&m)), 0, 0, 0)
			if ret == 0 || ret == ^uintptr(0) {
				return
			}
		}
	}()

	if err := <-ready; err != nil {
		fmt.Println(err)
		return
	}
	defer unhookWindowsHookEx.Call(hookH)

	fmt.Println("실행 중. Windows Terminal 한/영 → F12")
	fmt.Println("종료: Enter 또는 Ctrl+C")

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	enter := make(chan struct{})
	go func() {
		fmt.Scanln()
		close(enter)
	}()

	select {
	case <-sig:
	case <-enter:
	}
}
