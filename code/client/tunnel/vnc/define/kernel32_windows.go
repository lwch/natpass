package define

import "syscall"

var (
	libKernel32, _ = syscall.LoadLibrary("kernel32.dll")
	// FuncWTSGetActiveConsoleSessionId https://docs.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-wtsgetactiveconsolesessionid
	FuncWTSGetActiveConsoleSessionId, _ = syscall.GetProcAddress(syscall.Handle(libKernel32), "WTSGetActiveConsoleSessionId")
	// FuncGlobalAlloc https://docs.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-globalalloc
	FuncGlobalAlloc, _ = syscall.GetProcAddress(syscall.Handle(libKernel32), "GlobalAlloc")
	// FuncGlobalFree https://docs.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-globalfree
	FuncGlobalFree, _ = syscall.GetProcAddress(syscall.Handle(libKernel32), "GlobalFree")
)

// PROCESS_ALL_ACCESS https://docs.microsoft.com/en-us/windows/win32/procthread/process-security-and-access-rights
const PROCESS_ALL_ACCESS = 0x1F0FFF

const (
	// GHND https://docs.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-globalalloc
	GHND = 0x0042
	// GMEM_FIXED https://docs.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-globalalloc
	GMEM_FIXED = 0x0000
	// GMEM_MOVEABLE https://docs.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-globalalloc
	GMEM_MOVEABLE = 0x0002
	// GMEM_ZEROINIT https://docs.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-globalalloc
	GMEM_ZEROINIT = 0x0040
	// GPTR https://docs.microsoft.com/en-us/windows/win32/api/winbase/nf-winbase-globalalloc
	GPTR = 0x0040
)
