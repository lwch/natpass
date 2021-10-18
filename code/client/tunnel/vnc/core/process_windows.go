package core

import "golang.org/x/sys/windows"

// Process process
type Process struct {
}

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

func getSessionUserTokenWin() windows.Token {
	pid := getLogonPid(getSessionID())
	process, err := windows.OpenProcess(PROCESS_ALL_ACCESS, false, pid)
	if err != nil {
		return 0
	}
	defer windows.CloseHandle(process)
	var ret windows.Token
	windows.OpenProcessToken(process, windows.TOKEN_ALL_ACCESS, &ret)
	return ret
}

// CreateWorkerProcess create worker process
func CreateWorkerProcess() (*Process, error) {
	tk := getSessionUserTokenWin()
}
