package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/docker/go-plugins-helpers/volume"
)

const (
	mountPointRegular   = "/mnt/cloudstor/reg"
	mountPointMaxIO     = "/mnt/cloudstor/max"
	metaDataURL         = "http://169.254.169.254/latest/dynamic/instance-identity/document"
	nfsOptionVersion    = "nfsvers=4.1"
	nfsOptionRsize      = "rsize=1048576"
	nfsOptionWsize      = "wsize=1048576"
	nfsOptionTimeout    = "timeo=600"
	nfsOptionRetransmit = "retrans=2"
	nfsOptionHardlink   = "hard"
)

type efsDriver struct {
	m    sync.Mutex
	meta *metadataDriver
}

type metaData struct {
	AvailZone string `json:"availabilityZone,omitempty"`
	Region    string `json:"region,omitempty"`
}

func newEFSDriver(efsIDRegular, efsIDMaxIO, metadataRoot string) (*efsDriver, error) {
	md, err := fetchAWSMetaData()
	if err != nil {
		return nil, fmt.Errorf("Error resolving AWS metadata: %v", err)
	}

	mountURI := fmt.Sprintf("%s.efs.%s.amazonaws.com:/", efsIDRegular, md.Region)

	if err := os.MkdirAll(mountPointRegular, 0755); err != nil {
		return nil, fmt.Errorf("Error creating mount point: %v", err)
	}

	if err := efsMount(mountURI, mountPointRegular); err != nil {
		return nil, fmt.Errorf("Could not mount: %v", err)
	}

	mountURI = fmt.Sprintf("%s.efs.%s.amazonaws.com:/", efsIDMaxIO, md.Region)

	if err := os.MkdirAll(mountPointMaxIO, 0755); err != nil {
		return nil, fmt.Errorf("Error creating mount point: %v", err)
	}

	if err := efsMount(mountURI, mountPointMaxIO); err != nil {
		return nil, fmt.Errorf("Could not mount: %v", err)
	}

	metaDriver, err := newMetadataDriver(metadataRoot)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize metadata driver: %v", err)
	}

	return &efsDriver{
		meta: metaDriver,
	}, nil
}

func fetchAWSMetaData() (*metaData, error) {
	r, err := http.Get(metaDataURL)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	md := &metaData{}
	json.NewDecoder(r.Body).Decode(md)
	return md, nil
}

func (v *efsDriver) Capabilities(req volume.Request) (resp volume.Response) {
	resp.Capabilities = volume.Capability{Scope: "local"}
	return
}

func (v *efsDriver) Create(req volume.Request) (resp volume.Response) {
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

	mountpoint := mountPointRegular
	perfmode := req.Options["perfmode"]
	if perfmode == "maxio" {
		mountpoint = mountPointMaxIO
	}

	path := filepath.Join(mountpoint, req.Name)
	if err := os.MkdirAll(path, 0755); err != nil {
		resp.Err = fmt.Sprintf("Could not create volume: %v", err)
		logctx.Error(resp.Err)
		return
	}

	volMeta.VolPath = path
	volMeta.CreatedAt = time.Now().UTC()
	// Save volume metadata
	if err := v.meta.Set(req.Name, volMeta); err != nil {
		resp.Err = fmt.Sprintf("error saving metadata: %v", err)
		logctx.Error(resp.Err)
		return
	}

	return
}

func (v *efsDriver) Path(req volume.Request) (resp volume.Response) {
	v.m.Lock()
	defer v.m.Unlock()

	logctx := log.WithFields(log.Fields{
		"operation": "path",
		"name":      req.Name,
	})
	logctx.Debug("request accepted")

	meta, err := v.meta.Get(req.Name)
	if err != nil {
		resp.Err = fmt.Sprintf("could not fetch metadata: %v", err)
		logctx.Error(resp.Err)
		return
	}
	resp.Mountpoint = meta.VolPath
	return
}

func (v *efsDriver) Mount(req volume.MountRequest) (resp volume.Response) {
	v.m.Lock()
	defer v.m.Unlock()

	logctx := log.WithFields(log.Fields{
		"operation": "mount",
		"name":      req.Name,
	})
	logctx.Debug("request accepted")

	meta, err := v.meta.Get(req.Name)
	if err != nil {
		resp.Err = fmt.Sprintf("could not fetch metadata: %v", err)
		logctx.Error(resp.Err)
		return
	}
	resp.Mountpoint = meta.VolPath
	return
}

func (v *efsDriver) Unmount(req volume.UnmountRequest) (resp volume.Response) {
	v.m.Lock()
	defer v.m.Unlock()

	logctx := log.WithFields(log.Fields{
		"operation": "unmount",
		"name":      req.Name,
	})

	logctx.Debug("request accepted")
	return
}

func (v *efsDriver) Remove(req volume.Request) (resp volume.Response) {
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
	path := meta.VolPath

	if err := os.RemoveAll(path); err != nil {
		resp.Err = fmt.Sprintf("error removing path: %v", err)
		logctx.Error(resp.Err)
		return
	}

	logctx.Debug("removing volume metadata")
	if err := v.meta.Delete(req.Name); err != nil {
		resp.Err = err.Error()
		logctx.Error(resp.Err)
		return
	}

	return
}

func (v *efsDriver) Get(req volume.Request) (resp volume.Response) {
	v.m.Lock()
	defer v.m.Unlock()
	logctx := log.WithFields(log.Fields{
		"operation": "get",
		"name":      req.Name,
	})
	logctx.Debug("request accepted")

	resp.Volume = v.volumeEntry(req.Name)
	return
}

func (v *efsDriver) List(req volume.Request) (resp volume.Response) {
	v.m.Lock()
	defer v.m.Unlock()

	logctx := log.WithFields(log.Fields{
		"operation": "list",
	})
	logctx.Debug("request accepted")

	vols, err := v.meta.List()
	if err != nil {
		logctx.Error("failed to list managed volumes: %v", err)
		return
	}

	for _, vn := range vols {
		resp.Volumes = append(resp.Volumes, v.volumeEntry(vn))
	}
	logctx.Debugf("response has %d items", len(resp.Volumes))
	return
}

func (v *efsDriver) volumeEntry(name string) *volume.Volume {
	meta, err := v.meta.Get(name)
	if err != nil {
		log.Error("could not fetch metadata: %v", err)
		return nil
	}
	return &volume.Volume{Name: name, Mountpoint: meta.VolPath}
}

func efsMount(mountURI, mountPath string) error {
	// Set defaults
	opts := []string{
		nfsOptionVersion,
		nfsOptionRsize,
		nfsOptionWsize,
		nfsOptionTimeout,
		nfsOptionRetransmit,
		nfsOptionHardlink,
	}

	cmd := exec.Command("mount", "-t", "nfs4", "-o", strings.Join(opts, ","), mountURI, mountPath)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("mount failed: %v\noutput=%q", err, out)
	}

	cmd = exec.Command("mount")
	out, err = cmd.CombinedOutput()
	logctx := log.WithFields(log.Fields{
		"operation": "mount",
	})
	logctx.Debug(fmt.Sprintf("mount output=%s", out))

	return nil
}
