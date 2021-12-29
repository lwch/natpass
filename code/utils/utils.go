package utils

import (
	"os"
	"os/user"
	"strconv"

	"github.com/lwch/runtime"
)

func BuildLogDir(dir, u string) {
	runtime.Assert(os.MkdirAll(dir, 0755))
	if len(u) > 0 {
		us, err := user.Lookup(u)
		runtime.Assert(err)
		uid, _ := strconv.ParseInt(us.Uid, 10, 64)
		gid, _ := strconv.ParseInt(us.Gid, 10, 64)
		runtime.Assert(os.Chown(dir, int(uid), int(gid)))
	}
}
