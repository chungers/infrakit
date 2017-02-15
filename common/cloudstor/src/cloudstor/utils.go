package main

import (
  "bufio"
  "fmt"
  "os"
  "os/exec"
  "strings"

  log "github.com/Sirupsen/logrus"
)

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
