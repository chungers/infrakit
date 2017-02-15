package main

import (
  "fmt"
  "os"
  "os/exec"
  "path/filepath"
  "strings"
  "sync"
  "time"
  "crypto/md5"

  azure "github.com/Azure/azure-sdk-for-go/storage"
  log "github.com/Sirupsen/logrus"
  "github.com/docker/go-plugins-helpers/volume"
)

type azfsDriver struct {
  m            sync.Mutex
  cl           azure.FileServiceClient
  meta         *metadataDriver
  accountName  string
  accountKey   string
  storageBase  string
  mountpoint   string
  removeShares bool
}

const (
  mountPoint = "/mnt/cloudstor"
  cifs_version_option = "vers=2.1"
  cifs_file_mode_option = "file_mode=0777"
  cifs_dir_mode_option = "dir_mode=0777"
  cifs_uid_option = "uid=0"
  cifs_gid_option = "gid=0"
)

func newAZFSDriver(accountName, accountKey, metadataRoot string) (*azfsDriver, error) {

  storageBase := azure.DefaultBaseURL
  storageClient, err := azure.NewClient(accountName, accountKey, storageBase, azure.DefaultAPIVersion, true)
  if err != nil {
    return nil, fmt.Errorf("error creating azure client: %v", err)
  }

  metaDriver, err := newMetadataDriver(metadataRoot)
  if err != nil {
    return nil, fmt.Errorf("cannot initialize metadata driver: %v", err)
  }

  return &azfsDriver{
    cl:           storageClient.GetFileService(),
    meta:         metaDriver,
    accountName:  accountName,
    accountKey:   accountKey,
    storageBase:  storageBase,
    mountpoint:   mountPoint,
  }, nil
}

func (v *azfsDriver) Capabilities(req volume.Request) (resp volume.Response) {
  resp.Capabilities = volume.Capability{Scope: "local"}
  return
}

func (v *azfsDriver) Create(req volume.Request) (resp volume.Response) {
  v.m.Lock()
  defer v.m.Unlock()

  logctx := log.WithFields(log.Fields{
    "operation": "create",
    "name":      req.Name,
    "options":   req.Options})

  volMeta, err := v.meta.Validate(req.Options)
  if err != nil {
    resp.Err = fmt.Sprintf("error validating metadata: %v", err)
    logctx.Error(resp.Err)
    return
  }

  volMeta.Account = v.accountName
  volMeta.CreatedAt = time.Now().UTC()

  share := req.Options["share"]
  if share == "" {
      share = fmt.Sprintf("%x", md5.Sum([]byte(req.Name)))
      volMeta.Options.Share = share
  }
  logctx.Debug("request accepted")

  if ok, err := v.cl.CreateShareIfNotExists(share); err != nil {
    resp.Err = fmt.Sprintf("error creating azure file share: %v", err)
    logctx.Error(resp.Err)
    return
  } else if ok {
    logctx.Infof("created azure file share %q", share)
  }

  if err := v.meta.Set(req.Name, volMeta); err != nil {
    resp.Err = fmt.Sprintf("error saving metadata: %v", err)
    logctx.Error(resp.Err)
    return
  }
  return
}

func (v *azfsDriver) Path(req volume.Request) (resp volume.Response) {
  v.m.Lock()
  defer v.m.Unlock()

  log.WithFields(log.Fields{
    "operation": "path", "name": req.Name,
  }).Debug("request accepted")

  resp.Mountpoint = v.pathForVolume(req.Name)
  return
}

func (v *azfsDriver) Mount(req volume.MountRequest) (resp volume.Response) {
  v.m.Lock()
  defer v.m.Unlock()

  logctx := log.WithFields(log.Fields{
    "operation": "mount",
    "name":      req.Name,
  })
  logctx.Debug("request accepted")

  path := v.pathForVolume(req.Name)
  if err := os.MkdirAll(path, 0755); err != nil {
    resp.Err = fmt.Sprintf("could not create mount point: %v", err)
    logctx.Error(resp.Err)
    return
  }

  meta, err := v.meta.Get(req.Name)
  if err != nil {
    resp.Err = fmt.Sprintf("could not fetch metadata: %v", err)
    logctx.Error(resp.Err)
    return
  }

  if err := azfsMount(v.accountName, v.accountKey, v.storageBase, path, meta.Options); err != nil {
    resp.Err = err.Error()
    logctx.Error(resp.Err)
    return
  }
  resp.Mountpoint = path
  return
}

func (v *azfsDriver) Unmount(req volume.UnmountRequest) (resp volume.Response) {
  v.m.Lock()
  defer v.m.Unlock()

  logctx := log.WithFields(log.Fields{
    "operation": "unmount",
    "name":      req.Name,
  })

  logctx.Debug("request accepted")
  path := v.pathForVolume(req.Name)
  if err := unmount(path); err != nil {
    resp.Err = err.Error()
    logctx.Error(resp.Err)
    return
  }
  logctx.Debug("unmount successful")

  isActive, err := isMounted(path)
  if err != nil {
    resp.Err = err.Error()
    logctx.Error(resp.Err)
    return
  }
  if isActive {
    logctx.Debug("mountpoint still has active mounts, not removing")
  } else {
    logctx.Debug("mountpoint has no further mounts, removing")
    if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
      resp.Err = fmt.Sprintf("error removing mountpoint: %v", err)
      logctx.Error(resp.Err)
      return
    }
  }
  return
}

func (v *azfsDriver) Remove(req volume.Request) (resp volume.Response) {
  v.m.Lock()
  defer v.m.Unlock()

  logctx := log.WithFields(log.Fields{
    "operation": "remove",
    "name":      req.Name,
  })
  logctx.Debug("request accepted")

  meta, err := v.meta.Get(req.Name)
  if err != nil {
    resp.Err = fmt.Sprintf("could not fetch metadata: %v", err)
    logctx.Error(resp.Err)
    return
  }

  share := meta.Options.Share
  if ok, err := v.cl.DeleteShareIfExists(share); err != nil {
    resp.Err = fmt.Sprintf("error removing azure file share %q: %v", share, err)
    logctx.Error(resp.Err)
    return
  } else if ok {
    logctx.Infof("removed azure file share %q", share)
  }

  logctx.Debug("removing volume metadata")
  if err != v.meta.Delete(req.Name) {
    resp.Err = err.Error()
    logctx.Error(resp.Err)
    return
  }
  return
}

func (v *azfsDriver) Get(req volume.Request) (resp volume.Response) {
  v.m.Lock()
  defer v.m.Unlock()
  logctx := log.WithFields(log.Fields{
    "operation": "get",
    "name":      req.Name,
  })
  logctx.Debug("request accepted")

  _, err := v.meta.Get(req.Name)
  if err != nil {
    resp.Err = fmt.Sprintf("could not fetch metadata: %v", err)
    logctx.Error(resp.Err)
    return
  }
  resp.Volume = v.volumeEntry(req.Name)
  return
}

func (v *azfsDriver) List(req volume.Request) (resp volume.Response) {
  v.m.Lock()
  defer v.m.Unlock()

  logctx := log.WithFields(log.Fields{
    "operation": "list",
  })
  logctx.Debug("request accepted")

  vols, err := v.meta.List()
  if err != nil {
    resp.Err = fmt.Sprintf("failed to list managed volumes: %v", err)
    logctx.Error(resp.Err)
    return
  }

  for _, vn := range vols {
    resp.Volumes = append(resp.Volumes, v.volumeEntry(vn))
  }
  logctx.Debugf("response has %d items", len(resp.Volumes))
  return
}

func (v *azfsDriver) volumeEntry(name string) *volume.Volume {
  return &volume.Volume{Name: name,
    Mountpoint: v.pathForVolume(name)}
}

func (v *azfsDriver) pathForVolume(name string) string {
  return filepath.Join(v.mountpoint, name)
}

func azfsMount(accountName, accountKey, storageBase, mountPath string, options VolumeOptions) error {

  mountURI := fmt.Sprintf("//%s.file.%s/%s", accountName, storageBase, options.Share)

  opts := []string{
    cifs_version_option,
    fmt.Sprintf("username=%s", accountName),
    fmt.Sprintf("password=%s", accountKey),
    cifs_file_mode_option,
    cifs_dir_mode_option,
    cifs_uid_option,
    cifs_gid_option,
  }

  mntcmd := "mount -t cifs " + mountURI + " " + mountPath + " -o " + strings.Join(opts, ",")
  logctx := log.WithFields(log.Fields{
    "operation": "mount",
  })
  logctx.Debug(fmt.Sprintf("mount cmd=%s", mntcmd))

  cmd := exec.Command(mntcmd)
  out, err := cmd.CombinedOutput()
  if err != nil {
    return fmt.Errorf("mount failed: %v\noutput=%q", err, out)
  }

  cmd = exec.Command("mount")
  out, err = cmd.CombinedOutput()
  logctx.Debug(fmt.Sprintf("mount output=%s", out))

  return nil
}
