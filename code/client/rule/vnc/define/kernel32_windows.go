package define

import "syscall"

var (
	libKernel32, _ = syscall.LoadLibrary("kernel32.dll")
	// FuncWTSGetActiveConsoleSessionID https://docs.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-wtsgetactiveconsolesessionid
	FuncWTSGetActiveConsoleSessionID, _ = syscall.GetProcAddress(syscall.Handle(libKernel32), "WTSGetActiveConsoleSessionId")
	// FuncGlobalAlloc https://docs.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-globalalloc
	FuncGlobalAlloc, _ = syscall.GetProcAddress(syscall.Handle(libKernel32), "GlobalAlloc")
	// FuncGlobalFree https://docs.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-globalfree
	FuncGlobalFree, _ = syscall.GetProcAddress(syscall.Handle(libKernel32), "GlobalFree")
)

// PROCESSALLACCESS https://docs.microsoft.com/en-us/windows/win32/procthread/process-security-and-access-rights
const PROCESSALLACCESS = 0x1F0FFF

const (
	// GHND https://docs.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-globalalloc
	GHND = 0x0042
	// GMEMFIXED https://docs.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-globalalloc
	GMEMFIXED = 0x0000
	// GMEMMOVEABLE https://docs.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-globalalloc
	GMEMMOVEABLE = 0x0002
	// GMEMZEROINIT https://docs.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-globalalloc
	GMEMZEROINIT = 0x0040
	// GPTR https://docs.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-globalalloc
	GPTR = 0x0040
)
