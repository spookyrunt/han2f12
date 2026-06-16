package main

import (
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"unsafe"

	"github.com/spf13/pflag"
	"golang.org/x/sys/windows"
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

func pressKey(vk uint16, up bool) {
	const (
		inputSize      = 40
		inputKeyboard  = 1
		keyEventFKeyUp = 0x0002
	)
	var b [inputSize]byte
	*(*uint32)(unsafe.Pointer(&b[0])) = inputKeyboard
	*(*uint16)(unsafe.Pointer(&b[8])) = vk
	if up {
		*(*uint32)(unsafe.Pointer(&b[12])) = keyEventFKeyUp
	}
	sendInputProc.Call(1, uintptr(unsafe.Pointer(&b[0])), inputSize)
}

func foregroundExe() string {
	const processQueryLimited = 0x1000
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

var (
	hookH     uintptr
	hangulKey uint16
	hanjaKey  uint16
)

func hookCb(code int, wp, lp uintptr) uintptr {
	const (
		wmKeyDown     = 0x0100
		wmSysKeyDown  = 0x0104
		wmKeyUp       = 0x0101
		wmSysKeyUp    = 0x0105
		vkHangul      = 0x15
		vkHanja       = 0x19
		llkhfInjected = 0x10
	)
	if code >= 0 && (wp == wmKeyDown || wp == wmSysKeyDown ||
		wp == wmKeyUp || wp == wmSysKeyUp) {
		s := (*kbdHookStruct)(unsafe.Pointer(lp))
		injected := s.flags&llkhfInjected != 0
		if !injected && foregroundExe() == "windowsterminal.exe" {
			switch s.vkCode {
			case vkHangul:
				if wp == wmKeyDown || wp == wmSysKeyDown {
					pressKey(hangulKey, false)
					pressKey(hangulKey, true)
				}
				return 1
			case vkHanja:
				if wp == wmKeyDown || wp == wmSysKeyDown {
					pressKey(hanjaKey, false)
					pressKey(hanjaKey, true)
				}
				return 1
			}
			if s.vkCode == uint32(hangulKey) {
				if wp == wmKeyDown || wp == wmSysKeyDown {
					pressKey(vkHangul, false)
					pressKey(vkHangul, true)
				}
				return 1
			}
			if s.vkCode == uint32(hanjaKey) {
				if wp == wmKeyDown || wp == wmSysKeyDown {
					pressKey(vkHanja, false)
					pressKey(vkHanja, true)
				}
				return 1
			}
		}
	}
	r, _, _ := callNextHookEx.Call(hookH, uintptr(code), wp, lp)
	return r
}

func installHook() error {
	const whKeyboardLL = 13
	cb := syscall.NewCallback(hookCb)
	hookH, _, _ = setWindowsHookEx.Call(whKeyboardLL, cb, 0, 0)
	if hookH == 0 {
		return fmt.Errorf("훅 설치 실패")
	}
	return nil
}

func parseKey(s string) uint16 {
	keys := map[string]uint16{
		"F1": 0x70, "F2": 0x71, "F3": 0x72, "F4": 0x73,
		"F5": 0x74, "F6": 0x75, "F7": 0x76, "F8": 0x77,
		"F9": 0x78, "F10": 0x79, "F11": 0x7A, "F12": 0x7B,
	}
	if v, ok := keys[strings.ToUpper(s)]; ok {
		return v
	}
	fmt.Fprintf(os.Stderr, "알 수 없는 키: %s\n", s)
	os.Exit(1)
	return 0
}

func main() {
	{
		name, _ := windows.UTF16PtrFromString("Local\\han2f12_mutex")
		_, err := windows.CreateMutex(nil, false, name)
		if err != nil {
			if err == windows.ERROR_ALREADY_EXISTS {
				fmt.Fprintln(os.Stderr, "이미 실행 중입니다:", err)
			} else {
				fmt.Fprintln(os.Stderr, "이미 실행 중인 것 같습니다:", err)
			}
			return
		}
	}

	var hangulStr, hanjaStr string
	pflag.StringVar(&hangulStr, "hangul", "F12", "한영 키 매핑")
	pflag.StringVar(&hanjaStr, "hanja", "F9", "한자 키 매핑")
	pflag.Parse()
	hangulKey = parseKey(hangulStr)
	hanjaKey = parseKey(hanjaStr)

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
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer unhookWindowsHookEx.Call(hookH)
	fmt.Printf("실행 중. Windows Terminal 한/영 <-> %s, 한자 <-> %s\n", hangulStr, hanjaStr)
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
