package define

import "syscall"

var (
	libSas, _ = syscall.LoadLibrary("Sas.dll")
	// FuncSendSAS https://docs.microsoft.com/en-us/windows/win32/api/sas/nf-sas-sendsas
	FuncSendSAS, _ = syscall.GetProcAddress(syscall.Handle(libSas), "SendSAS")
)
