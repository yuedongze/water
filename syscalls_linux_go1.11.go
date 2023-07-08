//go:build linux && go1.11
// +build linux,go1.11

package water

import (
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

func openDev(config Config) (ifce *Interface, err error) {
	var fdInt int
	if fdInt, err = syscall.Open(
		"/dev/net/tun", os.O_RDWR|syscall.O_NONBLOCK, 0); err != nil {
		return nil, err
	}

	name, err := setupFd(config, uintptr(fdInt))
	if err != nil {
		return nil, err
	}

	f := os.NewFile(uintptr(fdInt), "tun")
	sc, err := f.SyscallConn()
	if err != nil {
		return nil, err
	}
	sc.Control(func(fd uintptr) {
		tunOffloads := unix.TUN_F_CSUM | unix.TUN_F_TSO4 | unix.TUN_F_TSO6
		err = unix.IoctlSetInt(int(fd), unix.TUNSETOFFLOAD, tunOffloads)
		if err != nil {
			panic(err)
		}
	})

	return &Interface{
		isTAP:           config.DeviceType == TAP,
		ReadWriteCloser: f,
		name:            name,
	}, nil
}
