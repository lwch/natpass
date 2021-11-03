package define

import "syscall"

var (
	libSas, _      = syscall.LoadLibrary("Sas.dll")
	FuncSendSAS, _ = syscall.GetProcAddress(syscall.Handle(libSas), "SendSAS")
)
