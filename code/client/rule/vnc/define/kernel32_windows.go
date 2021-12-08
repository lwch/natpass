package define

import "syscall"

var (
	libKernel32, _ = syscall.LoadLibrary("kernel32.dll")
	// FuncWTSGetActiveConsoleSessionID https://docs.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-wtsgetactiveconsolesessionid
	FuncWTSGetActiveConsoleSessionID, _ = syscall.GetProcAddress(syscall.Handle(libKernel32), "WTSGetActiveConsoleSessionId")
)

// PROCESSALLACCESS https://docs.microsoft.com/en-us/windows/win32/procthread/process-security-and-access-rights
const PROCESSALLACCESS = 0x1F0FFF
