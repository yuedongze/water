//go:build linux && go1.11
// +build linux,go1.11

package water

import (
	"log"
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
		var (
			ifr *unix.Ifreq
		)
		ifr, err = unix.NewIfreq(name)
		if err != nil {
			return
		}
		err = unix.IoctlIfreq(int(fd), unix.TUNGETIFF, ifr)
		if err != nil {
			return
		}
		got := ifr.Uint16()
		if got&unix.IFF_VNET_HDR == 0 {
			log.Panicln("IFF_VNET_HDR is not enabled")
		}

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
