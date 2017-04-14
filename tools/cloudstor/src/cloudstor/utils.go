package main

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	log "github.com/Sirupsen/logrus"
)

const (
	/* For EXT file systems, look for 0xEF 0x53 at offset 1080 */
	ext4MagicString = "\123\357" // octal for EF 53 reversed
	ext4MagicOffset = 0x438
)

func getNameFromUrl(url string) string {
	elements := strings.Split(url, "/")
	return elements[len(elements)-1]
}

func isExtFS(dev string) (bool, error) {
	ext4MagicBufferLen := ext4MagicOffset + uint64(len(ext4MagicString))
	file, err := os.Open(dev)
	if err != nil {
		return false, fmt.Errorf("failed to open device to probe ext4: %v", err)
	}
	defer file.Close()

	buf := make([]byte, ext4MagicBufferLen)
	l, err := file.Read(buf)
	if err != nil {
		return false, fmt.Errorf("failed to read magic bytes for ext4: %v", err)
	}

	if uint64(l) < ext4MagicBufferLen {
		return false, fmt.Errorf("failed to read all magic bytes for ext4")
	}

	if bytes.Equal(
		[]byte(ext4MagicString), buf[ext4MagicOffset:ext4MagicBufferLen]) {
		return true, nil
	}
	return false, nil
}

func unmount(mountpoint string) error {
	cmd := exec.Command("umount", mountpoint)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("unmount failed: %v\noutput=%q", err, out)
	}
	return nil
}

// isMounted reads /proc/self/mountinfo to see if the specified mountpoint is
// mounted.
func isMounted(mountpoint string) (bool, error) {
	f, err := os.Open("/proc/self/mountinfo")
	if err != nil {
		return false, fmt.Errorf("cannot read mountinfo: %v", err)
	}
	defer f.Close()

	// format of mountinfo:
	//    38 23 0:30 / /sys/fs/cgroup/devices rw,relatime - cgroup cgroup rw,devices
	//    39 23 0:31 / /sys/fs/cgroup/freezer rw,relatime - cgroup cgroup rw,freezer
	//    33 22 8:17 / /mnt rw,relatime - ext4 /dev/sdb1 rw,data=ordered
	// so we split the lines into the specified format and match the mountpoint
	// at 5th field.
	//
	// This code is adopted from https://github.com/docker/docker/blob/master/pkg/mount/mountinfo_linux.go

	oldFi, err := os.Stat(mountpoint)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("cannot stat mountpoint: %v", err)
	}

	s := bufio.NewScanner(f)
	for s.Scan() {
		t := s.Text()
		f := strings.Fields(t)
		if len(f) < 5 {
			return false, fmt.Errorf("mountinfo line %q has less than 5 fields, cannot parse mountpoint", t)
		}
		mp := f[4] // ID, Parent, Major, Minor, Root, *Mountpoint*, Opts, OptionalFields
		fi, err := os.Stat(mp)
		if err != nil {
			return false, fmt.Errorf("cannot stat %s: %v", mp, err)
		}
		same := os.SameFile(oldFi, fi)
		if same {
			return true, nil
		}
	}
	log.Debug("mountpoint not found")
	return false, nil
}
