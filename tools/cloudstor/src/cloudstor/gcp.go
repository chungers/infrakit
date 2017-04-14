package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/compute/metadata"
	log "github.com/Sirupsen/logrus"
	"github.com/docker/go-plugins-helpers/volume"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
)

type gcpDriver struct {
	m          sync.Mutex
	meta       *metadataDriver
	cl         *compute.Service
	mountpoint string
	project    string
	zone       string
	instance   string
}

const (
	diskTypeStandard = "pd-standard"
	diskTypeSSD      = "pd-ssd"
	formatcmd        = "mkfs.ext4"
	scans            = 10
	scanInterval     = 1000 //ms
	pollInterval     = 1000 //ms
)

func newGCPDriver(metadataRoot string) (*gcpDriver, error) {

	gcpclient := func() (*compute.Service, error) {
		client, err := google.DefaultClient(context.TODO(), compute.ComputeScope)
		if err != nil {
			fmt.Printf("failed to create client: %v", err)
			return nil, err
		}
		return compute.New(client)
	}
	service, err := gcpclient()
	if err != nil {
		return nil, fmt.Errorf("error initializing gcp client: %v", err)
	}

	project, err := metadata.ProjectID()
	if err != nil {
		return nil, fmt.Errorf("error obtaining gcp project: %v", err)
	}

	zone, err := metadata.Zone()
	if err != nil {
		return nil, fmt.Errorf("error obtaining gcp zone: %v", err)
	}

	instance, err := metadata.InstanceName()
	if err != nil {
		return nil, fmt.Errorf("error obtaining gcp instance: %v", err)
	}

	metaDriver, err := newMetadataDriver(metadataRoot)
	if err != nil {
		return nil, fmt.Errorf("cannot initialize metadata driver: %v", err)
	}

	if err := os.MkdirAll(mountPoint, 0755); err != nil {
		return nil, fmt.Errorf("error creating mount point: %v", err)
	}

	return &gcpDriver{
		cl:         service,
		meta:       metaDriver,
		mountpoint: mountPoint,
		project:    project,
		zone:       zone,
		instance:   instance,
	}, nil
}

func (v *gcpDriver) Capabilities(req volume.Request) (resp volume.Response) {
	resp.Capabilities = volume.Capability{Scope: "local"}
	return
}

func (v *gcpDriver) Create(req volume.Request) (resp volume.Response) {
	v.m.Lock()
	defer v.m.Unlock()

	logctx := log.WithFields(log.Fields{
		"operation": "create",
		"name":      req.Name,
		"options":   req.Options})

	volmeta, err := v.meta.Validate(req.Options)
	if err != nil {
		resp.Err = fmt.Sprintf("error validating metadata: %v", err)
		logctx.Error(resp.Err)
		return
	}

	volmeta.CreatedAt = time.Now().UTC()

	if v.diskExists(req.Name) {
		logctx.Infof("disk %q already exists in zone", req.Name)
		if err := v.meta.Set(req.Name, volmeta); err != nil {
			resp.Err = fmt.Sprintf("error saving metadata: %v", err)
			logctx.Error(resp.Err)
			return
		}
		return
	}

	size := req.Options["size"]
	szGB, err := strconv.ParseInt(size, 10, 64)
	if err != nil {
		resp.Err = fmt.Sprintf("invalid volume size: %v", err)
		logctx.Error(resp.Err)
		return
	}

	perfmode := req.Options["perfmode"]
	if perfmode == "" {
		perfmode = diskTypeStandard
	}

	if !strings.EqualFold(perfmode, diskTypeStandard) && !strings.EqualFold(perfmode, diskTypeSSD) {
		resp.Err = fmt.Sprintf("invalid perf mode: %q", perfmode)
		logctx.Error(resp.Err)
		return
	}
	logctx.Debug("request accepted")

	uri := fmt.Sprintf("zones/%s/diskTypes/%s", v.zone, perfmode)
	diskParams := &compute.Disk{
		Name:   req.Name,
		SizeGb: szGB,
		Type:   uri,
	}

	aop, err := v.cl.Disks.Insert(v.project, v.zone, diskParams).Do()
	if err != nil {
		resp.Err = fmt.Sprintf("error invoking insert disk API in GCP: %v", err)
		logctx.Error(resp.Err)
		return
	}

	if err := gcpwait(v.project, v.zone, aop, v.cl); err != nil {
		resp.Err = fmt.Sprintf("error detected while creating disk in GCP: %v", err)
		logctx.Error(resp.Err)
		return
	}

	logctx.Infof("disk creation succeded for %q", req.Name)

	if err := v.meta.Set(req.Name, volmeta); err != nil {
		resp.Err = fmt.Sprintf("error saving metadata: %v", err)
		logctx.Error(resp.Err)
		return
	}
	return
}

func (v *gcpDriver) Path(req volume.Request) (resp volume.Response) {
	v.m.Lock()
	defer v.m.Unlock()

	log.WithFields(log.Fields{
		"operation": "path", "name": req.Name,
	}).Debug("request accepted")

	resp.Mountpoint = v.pathForVolume(req.Name)
	return
}

func (v *gcpDriver) Mount(req volume.MountRequest) (resp volume.Response) {
	v.m.Lock()
	defer v.m.Unlock()

	logctx := log.WithFields(log.Fields{
		"operation": "mount",
		"name":      req.Name,
	})
	logctx.Debug("request accepted")

	path := v.pathForVolume(req.Name)

	/* Perform the following steps: attach, format, create mount pt, mount.
	   Check before each step whether it's needed so that the overall mount
	   request is idempotent and can process a volume that made it
	   partially through attach/format/mount when a node/docker daemon crashed
	*/
	attached, err := v.isDiskAttached(req.Name)
	if err != nil {
		resp.Err = fmt.Sprintf("failed to get volume attachment details: %v", err)
		logctx.Error(resp.Err)
		return
	}

	if !attached {
		if err := v.attachDisk(req.Name); err != nil {
			resp.Err = fmt.Sprintf("failed to attach disk: %v", err)
			logctx.Error(resp.Err)
			return
		}
		logctx.Infof("disk attach succeded for %q", req.Name)
	}

	dev, err := v.getBlkDev(req.Name)
	if err != nil {
		resp.Err = fmt.Sprintf("failed to scan device: %v", err)
		logctx.Error(resp.Err)
		return
	}
	dev = "/dev/" + dev

	/* only support EXT4 for now. We can expand to XFS, etc later */
	formatted, err := isExtFS(dev)
	if err != nil {
		resp.Err = fmt.Sprintf("failed to probe volume FS: %v", err)
		logctx.Error(resp.Err)
		return
	}

	if !formatted {
		cmd := exec.Command("mkfs.ext4", "-F", dev)
		if out, err := cmd.CombinedOutput(); err != nil {
			resp.Err = fmt.Sprintf("format failed: %v\noutput=%q", err, out)
			logctx.Error(resp.Err)
			return
		}
		logctx.Infof("formatting succeeded for dev: %q", dev)
	}

	if err := os.MkdirAll(path, 0755); err != nil {
		resp.Err = fmt.Sprintf("could not create mount point: %v", err)
		logctx.Error(resp.Err)
		return
	}

	mounted, err := isMounted(path)
	if err != nil {
		resp.Err = fmt.Sprintf("could not evaluate mount points: %v", err)
		logctx.Error(resp.Err)
		return
	}

	if !mounted {
		cmd := exec.Command("mount", "-t", "ext4", "-v", dev, path)
		if out, err := cmd.CombinedOutput(); err != nil {
			resp.Err = fmt.Sprintf("mount failed: %v\noutput=%q", err, out)
			logctx.Error(resp.Err)
			return
		}
		logctx.Infof("mount succeeded for dev: %q", dev)
	}

	resp.Mountpoint = path
	return
}

func (v *gcpDriver) Unmount(req volume.UnmountRequest) (resp volume.Response) {
	v.m.Lock()
	defer v.m.Unlock()

	logctx := log.WithFields(log.Fields{
		"operation": "unmount",
		"name":      req.Name,
	})

	logctx.Debug("request accepted")
	path := v.pathForVolume(req.Name)
	if err := unmount(path); err != nil {
		resp.Err = fmt.Sprintf("unmount failed: %v", err)
		logctx.Error(resp.Err)
		return
	}
	logctx.Debug("unmount successful")
	return
}

func (v *gcpDriver) Remove(req volume.Request) (resp volume.Response) {
	v.m.Lock()
	defer v.m.Unlock()

	logctx := log.WithFields(log.Fields{
		"operation": "remove",
		"name":      req.Name,
	})
	logctx.Debug("request accepted")

	_, err := v.meta.Get(req.Name)
	if err != nil {
		resp.Err = fmt.Sprintf("could not fetch metadata: %v", err)
		logctx.Error(resp.Err)
		return
	}

	aop, err := v.cl.Disks.Delete(v.project, v.zone, req.Name).Do()
	if err != nil {
		resp.Err = fmt.Sprintf("error deleting disk during remove: %v", err)
		logctx.Error(resp.Err)
		return
	}

	if err := gcpwait(v.project, v.zone, aop, v.cl); err != nil {
		resp.Err = fmt.Sprintf("error creating disk: %v", err)
		logctx.Error(resp.Err)
		return
	}

	logctx.Debug("removing volume metadata")
	if err != v.meta.Delete(req.Name) {
		resp.Err = err.Error()
		logctx.Error(resp.Err)
		return
	}
	return
}

func (v *gcpDriver) Get(req volume.Request) (resp volume.Response) {
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

func (v *gcpDriver) List(req volume.Request) (resp volume.Response) {
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

func (v *gcpDriver) diskExists(name string) bool {
	if _, err := v.cl.Disks.Get(v.project, v.zone, name).Do(); err != nil {
		return false
	}
	return true
}

func (v *gcpDriver) attachDisk(name string) error {
	attachParams := &compute.AttachedDisk{
		AutoDelete: false,
		Boot:       false,
		Source:     fmt.Sprintf("zones/%s/disks/%s", v.zone, name),
		DeviceName: name,
	}

	/* Determine if disk is still attached to another instance and detach first */
	disk, err := v.cl.Disks.Get(v.project, v.zone, name).Do()
	if err != nil {
		return fmt.Errorf("error querying disk to determine attachments: %v", err)
	}

	if disk == nil {
		return fmt.Errorf("disk not found when trying to determine attachments: %v", err)
	}

	if len(disk.Users) > 0 {
		if err := v.detachDisk(getNameFromUrl(disk.Users[0]), name); err != nil {
			return fmt.Errorf("error detaching disk while trying to attach: %v", err)
		}
	}

	aop, err := v.cl.Instances.AttachDisk(v.project, v.zone, v.instance, attachParams).Do()
	if err != nil {
		return fmt.Errorf("error initiating disk attach: %v", err)
	}

	if err := gcpwait(v.project, v.zone, aop, v.cl); err != nil {
		return fmt.Errorf("disk attach failed: %v", err)
	}
	return nil
}

func (v *gcpDriver) detachDisk(instanceName, devName string) error {
	aop, err := v.cl.Instances.DetachDisk(v.project, v.zone, instanceName, devName).Do()
	if err != nil {
		return fmt.Errorf("error initiating disk detach: %v", err)
	}

	if err := gcpwait(v.project, v.zone, aop, v.cl); err != nil {
		return fmt.Errorf("disk detach failed: %v", err)
	}
	return nil
}

func (v *gcpDriver) isDiskAttached(name string) (bool, error) {
	instance, err := v.cl.Instances.Get(v.project, v.zone, v.instance).Do()
	if err != nil {
		return false, err
	}

	for _, disk := range instance.Disks {
		if disk.DeviceName == name {
			return true, nil
		}
	}

	return false, err
}

func (v *gcpDriver) volumeEntry(name string) *volume.Volume {
	return &volume.Volume{Name: name,
		Mountpoint: v.pathForVolume(name)}
}

func (v *gcpDriver) pathForVolume(name string) string {
	return filepath.Join(v.mountpoint, name)
}

func (v *gcpDriver) getBlkDev(pg83ID string) (string, error) {
	/* scan through all SCSI devices looking for a match for ID in pg83 VPD */

	blkdevs, err := filepath.Glob("/sys/block/sd*")
	if err != nil {
		return "", fmt.Errorf("failed to scan devices: %v", err)
	}

	for i := 0; i < scans; i += 1 {
		for _, blkdev := range blkdevs {

			blkdevPath, err := filepath.EvalSymlinks(blkdev)
			if err != nil {
				return "", fmt.Errorf("failed to to read sys block link: %v", err)
			}

			blkdevIDFile := path.Join(blkdevPath, "/../../vpd_pg83")

			devID, err := ioutil.ReadFile(blkdevIDFile)
			if err != nil {
				return "", fmt.Errorf("failed to read contents of %q: %v", blkdevIDFile, err)
			}
			s := string(devID)
			if strings.Contains(s, pg83ID) {
				return path.Base(blkdev), nil
			}
		}
		time.Sleep(scanInterval * time.Millisecond)
	}
	return "", fmt.Errorf("device not found during scans")
}

func gcpwait(project, zone string, op *compute.Operation, svc *compute.Service) error {
	for {
		time.Sleep(pollInterval * time.Millisecond)
		op, err := svc.ZoneOperations.Get(project, zone, op.Name).Do()
		if err != nil {
			return err
		}

		switch op.Status {
		case "PENDING", "RUNNING":
			continue

		case "DONE":
			if op.Error != nil {
				bytes, _ := op.Error.MarshalJSON()
				return fmt.Errorf(string(bytes))
			}
			return nil

		default:
			return fmt.Errorf("unknown status.")
		}
	}
}
