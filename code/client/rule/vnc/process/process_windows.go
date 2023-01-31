package process

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"unicode/utf16"
	"unsafe"

	"github.com/lwch/natpass/code/client/rule/vnc/define"
	"github.com/lwch/natpass/code/client/rule/vnc/vncnetwork"
	"golang.org/x/sys/windows"
)

func getLogonPid(sessionID uintptr) uint32 {
	snapshot, err := syscall.CreateToolhelp32Snapshot(syscall.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return 0
	}
	defer syscall.CloseHandle(snapshot)
	var procEntry syscall.ProcessEntry32
	procEntry.Size = uint32(unsafe.Sizeof(procEntry))
	err = syscall.Process32First(snapshot, &procEntry)
	if err != nil {
		return 0
	}
	var ret uint32
	for {
		var sid uint32
		name := parseProcessName(procEntry.ExeFile)
		if strings.ToLower(name) != "winlogon.exe" {
			goto next
		}
		err = windows.ProcessIdToSessionId(procEntry.ProcessID, &sid)
		if err != nil {
			return ret
		}
		if sid == uint32(sessionID) {
			ret = procEntry.ProcessID
		}
	next:
		err = syscall.Process32Next(snapshot, &procEntry)
		if err != nil {
			return ret
		}
	}
}

func parseProcessName(exeFile [syscall.MAX_PATH]uint16) string {
	for i, v := range exeFile {
		if v <= 0 {
			return string(utf16.Decode(exeFile[:i]))
		}
	}
	return ""
}

func getSessionID() uintptr {
	id, _, _ := syscall.Syscall(define.FuncWTSGetActiveConsoleSessionID, 0, 0, 0, 0)
	return id
}

func getSessionUserTokenWin() windows.Token {
	pid := getLogonPid(getSessionID())
	process, err := windows.OpenProcess(define.PROCESSALLACCESS, false, pid)
	if err != nil {
		return 0
	}
	defer windows.CloseHandle(process)
	var ret windows.Token
	windows.OpenProcessToken(process, windows.TOKEN_ALL_ACCESS, &ret)
	return ret
}

// CreateWorker create worker process
func CreateWorker(name, confDir string, showCursor bool) (*Process, error) {
	tk := getSessionUserTokenWin()
	if tk != 0 {
		defer windows.CloseHandle(windows.Handle(tk))
	}
	return createWorker(name, confDir, tk, showCursor)
}

func createWorker(name, confDir string, tk windows.Token, showCursor bool) (*Process, error) {
	dir, err := os.Executable()
	if err != nil {
		return nil, err
	}
	var p Process
	p.chWrite = make(chan *vncnetwork.VncMsg)
	p.chImage = make(chan *vncnetwork.ImageData)
	p.chClipboard = make(chan *vncnetwork.ClipboardData)
	port, err := p.listenAndServe()
	if err != nil {
		return nil, err
	}
	var startup windows.StartupInfo
	var process windows.ProcessInformation
	startup.Cb = uint32(unsafe.Sizeof(startup))
	startup.Desktop = windows.StringToUTF16Ptr("WinSta0\\default")
	startup.Flags = windows.STARTF_USESHOWWINDOW
	str := dir + fmt.Sprintf(" vnc --conf %s --name %s --port %d", confDir, name, port)
	if showCursor {
		str += "--cursor"
	}
	cmd := windows.StringToUTF16Ptr(str)
	if tk == 0 {
		// only for debug
		startup.Flags = 0
		err = windows.CreateProcess(nil, cmd, nil, nil, false, windows.CREATE_NEW_CONSOLE, nil, nil, &startup, &process)
	} else {
		err = windows.CreateProcessAsUser(tk, nil, cmd, nil, nil, false, windows.DETACHED_PROCESS, nil, nil, &startup, &process)
	}
	if err != nil {
		return nil, err
	}
	p.pid = int(process.ProcessId)
	return &p, nil
}
