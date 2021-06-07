package helpers

import (
	"os"
	"strconv"
	"syscall"

	"github.com/osixia/container-baseimage/log"
)

// PIDs functions
// =============================

func ListPIDs() ([]int, error) {

	log.Trace("ListPIDs called")

	var ret []int

	d, err := os.Open("/proc")
	if err != nil {
		return nil, err
	}
	defer d.Close()

	fnames, err := d.Readdirnames(-1)
	if err != nil {
		return nil, err
	}
	for _, fname := range fnames {
		pid, err := strconv.ParseInt(fname, 10, 32)
		if err != nil {
			// if not numeric name, just skip
			continue
		}

		ipid := int(pid)

		// ignore self pid
		if ipid == os.Getpid() {
			continue
		}

		ret = append(ret, ipid)
	}

	return ret, nil
}

// Signals functions
// =============================

func KillAll(sig syscall.Signal) error {

	log.Tracef("KillAll called with sig: %v", sig)

	pids, err := ListPIDs()
	if err != nil {
		return nil
	}

	log.Tracef("pids: %v", pids)

	for _, pid := range pids {
		log.Tracef("Sending %v to pid %v ...", sig, pid)
		if err := syscall.Kill(pid, sig); err != nil {
			return err
		}
	}

	return nil
}
