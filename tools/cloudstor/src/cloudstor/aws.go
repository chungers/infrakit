package main

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/docker/go-plugins-helpers/volume"
)

const (
	mountPointRegular             = "/mnt/efs/reg"
	mountPointMaxIO               = "/mnt/efs/max"
	mountPointEBS                 = "/mnt/ebs"
	metaDataURL                   = "http://169.254.169.254/latest/dynamic/instance-identity/document"
	nfsOptionVersion              = "nfsvers=4.1"
	nfsOptionRsize                = "rsize=1048576"
	nfsOptionWsize                = "wsize=1048576"
	nfsOptionTimeout              = "timeo=600"
	nfsOptionRetransmit           = "retrans=2"
	nfsOptionHardlink             = "hard"
	backingTypeLocal              = "local"
	backingTypeShared             = "shared"
	perfmodeMaxIO                 = "maxio"
	perfmodeRegIO                 = "regio"
	cloudstorEBSVolumeTag         = "CloudstorVolume"
	createEBSTimeout              = 600 * time.Second
	detachEBSTimeout              = 600 * time.Second
	attachEBSTimeout              = 600 * time.Second
	createSnapTimeout             = 600 * time.Second
	retryInterval                 = 10 * time.Second
	snapshotLoopInterval          = 300 * time.Second
	snapshotLoopIterationInterval = 10 * time.Second
	volumeTypeProvisionedIOPS     = "io1"
)

type awsDriver struct {
	m            sync.Mutex
	cl           *ec2.EC2
	az           string
	instanceID   string
	stackID      string
	region       string
	efsSupported bool
}

type metaData struct {
	AvailZone  string `json:"availabilityZone,omitempty"`
	Region     string `json:"region,omitempty"`
	InstanceId string `json:"instanceId,omitempty"`
}

type awsVol struct {
	backingType string
	efsType     string
}

func newAWSDriver(efsIDRegular, efsIDMaxIO, metadataRoot, stackID string, efsSupported bool) (*awsDriver, error) {
	sess, err := session.NewSession(aws.NewConfig().WithLogLevel(aws.LogDebugWithRequestRetries | aws.LogDebugWithRequestErrors))
	if err != nil {
		return nil, fmt.Errorf("Error initializing session: %v", err)
	}

	cl := ec2.New(sess)

	md, err := fetchAWSMetaData()
	if err != nil {
		return nil, fmt.Errorf("Error resolving AWS metadata: %v", err)
	}

	if efsSupported {
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
	}

	if err := os.MkdirAll(mountPointEBS, 0755); err != nil {
		return nil, fmt.Errorf("error creating EBS mount point: %v", err)
	}

	stackIDmd5 := fmt.Sprintf("%x", md5.Sum([]byte(stackID)))

	v := &awsDriver{
		cl:           cl,
		az:           md.AvailZone,
		region:       md.Region,
		instanceID:   md.InstanceId,
		stackID:      stackIDmd5,
		efsSupported: efsSupported,
	}

	go v.snapshotVolumesLoop()
	return v, nil
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

func (v *awsDriver) Capabilities(req volume.Request) (resp volume.Response) {
	resp.Capabilities = volume.Capability{Scope: "local"}
	return
}

func (v *awsDriver) Create(req volume.Request) (resp volume.Response) {
	v.m.Lock()
	defer v.m.Unlock()

	logctx := log.WithFields(log.Fields{
		"operation": "create",
		"name":      req.Name,
		"options":   req.Options})

	// default to EFS when EFS supported (since it has less restrictions)
	// otherwise default to EBS
	voltype := req.Options["backing"]
	if voltype == "" {
		if v.efsSupported {
			voltype = backingTypeShared
		} else {
			voltype = backingTypeLocal
		}
	}
	if !strings.EqualFold(voltype, backingTypeLocal) && !strings.EqualFold(voltype, backingTypeShared) {
		resp.Err = fmt.Sprintf("invalid backing type specified: %q", voltype)
		logctx.Error(resp.Err)
		return
	}

	logctx.Debug("request accepted")

	var err error
	if voltype == backingTypeShared {
		err = v.createEFS(req)
	} else {
		err = v.createEBS(req)
	}
	if err != nil {
		resp.Err = fmt.Sprintf("volume creation failed: %v", err)
		logctx.Error(resp.Err)
		return
	}

	return
}

func (v *awsDriver) Path(req volume.Request) (resp volume.Response) {
	v.m.Lock()
	defer v.m.Unlock()

	logctx := log.WithFields(log.Fields{
		"operation": "path",
		"name":      req.Name,
	})
	logctx.Debug("request accepted")

	vol, err := v.getAWSVolume(req.Name)
	if err != nil {
		resp.Err = fmt.Sprintf("could not fetch volume: %v", err)
		logctx.Error(resp.Err)
		return
	}

	if vol.backingType == backingTypeLocal {
		resp.Mountpoint = v.pathForEBSVolume(req.Name)
	} else {
		resp.Mountpoint = v.pathForEFSVolume(req.Name, vol.efsType)
	}

	return
}

func (v *awsDriver) Mount(req volume.MountRequest) (resp volume.Response) {
	v.m.Lock()
	defer v.m.Unlock()

	logctx := log.WithFields(log.Fields{
		"operation": "mount",
		"name":      req.Name,
	})
	logctx.Debug("request accepted")

	vol, err := v.getAWSVolume(req.Name)
	if err != nil {
		resp.Err = fmt.Sprintf("could not fetch volume: %v", err)
		logctx.Error(resp.Err)
		return
	}

	if vol.backingType == backingTypeLocal {
		mountPath, err := v.mountEBS(req)
		if err != nil {
			resp.Err = fmt.Sprintf("error mounting volume: %v", err)
			logctx.Error(resp.Err)
			return
		}
		resp.Mountpoint = mountPath
	} else {
		resp.Mountpoint = v.pathForEFSVolume(req.Name, vol.efsType)
	}

	return
}

func (v *awsDriver) Unmount(req volume.UnmountRequest) (resp volume.Response) {
	v.m.Lock()
	defer v.m.Unlock()

	logctx := log.WithFields(log.Fields{
		"operation": "unmount",
		"name":      req.Name,
	})

	logctx.Debug("request accepted")

	vol, err := v.getAWSVolume(req.Name)
	if err != nil {
		resp.Err = fmt.Sprintf("could not fetch volume: %v", err)
		logctx.Error(resp.Err)
		return
	}

	if vol.backingType == backingTypeLocal {
		path := v.pathForEBSVolume(req.Name)
		// we do not support multiple mounts so this is okay.
		// need to reference count mounts if do want multiple mounts
		if err := unmount(path); err != nil {
			resp.Err = fmt.Sprintf("unmount failed: %v", err)
			logctx.Error(resp.Err)
			return
		}
	}
	return
}

func (v *awsDriver) Remove(req volume.Request) (resp volume.Response) {
	v.m.Lock()
	defer v.m.Unlock()

	logctx := log.WithFields(log.Fields{
		"operation": "remove",
		"name":      req.Name,
	})
	logctx.Debug("request accepted")

	vol, err := v.getAWSVolume(req.Name)
	if err != nil {
		resp.Err = fmt.Sprintf("could not fetch volume: %v", err)
		logctx.Error(resp.Err)
		return
	}

	if vol.backingType == backingTypeShared {
		err = v.removeEFS(v.pathForEFSVolume(req.Name, vol.efsType))
	} else {
		err = v.removeEBS(req)
	}
	if err != nil {
		resp.Err = fmt.Sprintf("could not remove volume: %v", err)
		logctx.Error(resp.Err)
		return
	}
	return
}

func (v *awsDriver) Get(req volume.Request) (resp volume.Response) {
	v.m.Lock()
	defer v.m.Unlock()
	logctx := log.WithFields(log.Fields{
		"operation": "get",
		"name":      req.Name,
	})
	logctx.Debug("request accepted")

	vol, err := v.getAWSVolume(req.Name)
	if err != nil {
		resp.Err = fmt.Sprintf("could not fetch volume: %v", err)
		logctx.Error(resp.Err)
		return
	}

	if vol.backingType == backingTypeShared {
		resp.Volume = &volume.Volume{Name: req.Name, Mountpoint: v.pathForEFSVolume(req.Name, vol.efsType)}
	} else {
		resp.Volume = &volume.Volume{Name: req.Name, Mountpoint: v.pathForEBSVolume(req.Name)}
	}
	return
}

func (v *awsDriver) List(req volume.Request) (resp volume.Response) {
	v.m.Lock()
	defer v.m.Unlock()

	logctx := log.WithFields(log.Fields{
		"operation": "list",
	})
	logctx.Debug("request accepted")

	volsEBS, err := v.getEBSVolumesByStack()
	if err != nil {
		logctx.Error(fmt.Sprintf("Failed to get volumes attached to instance: %v", err))
		return
	}

	nametbl := map[string]bool{}
	for _, vol := range volsEBS {
		for _, tag := range vol.Tags {
			if *tag.Key == "CloudstorVolumeName" {
				name := *tag.Value
				if seen := nametbl[name]; !seen {
					nametbl[name] = true
					resp.Volumes = append(resp.Volumes, &volume.Volume{Name: name, Mountpoint: v.pathForEBSVolume(name)})
				}
			}
		}
	}

	if v.efsSupported {
		files, err := ioutil.ReadDir(mountPointRegular)
		if err != nil {
			logctx.Error(fmt.Sprintf("Failed to get regular EFS volumes: %v", err))
			return
		}
		for _, file := range files {
			name := file.Name()
			if file.IsDir() {
				resp.Volumes = append(resp.Volumes, &volume.Volume{Name: name, Mountpoint: v.pathForEFSVolume(name, perfmodeRegIO)})
			}
		}

		files, err = ioutil.ReadDir(mountPointMaxIO)
		if err != nil {
			logctx.Error(fmt.Sprintf("Failed to get MaxIO EFS volumes: %v", err))
			return
		}
		for _, file := range files {
			name := file.Name()
			if file.IsDir() {
				resp.Volumes = append(resp.Volumes, &volume.Volume{Name: name, Mountpoint: v.pathForEFSVolume(name, perfmodeMaxIO)})
			}
		}
	}

	logctx.Debugf("response has %d items", len(resp.Volumes))
	return
}

func (v *awsDriver) getAWSVolume(name string) (*awsVol, error) {
	if v.efsSupported {
		_, err := os.Stat(v.pathForEFSVolume(name, perfmodeRegIO))
		if err == nil {
			return &awsVol{backingType: backingTypeShared, efsType: perfmodeRegIO}, nil
		}

		_, err = os.Stat(v.pathForEFSVolume(name, perfmodeMaxIO))
		if err == nil {
			return &awsVol{backingType: backingTypeShared, efsType: perfmodeMaxIO}, nil
		}
	}

	_, err := v.getEBSByName(name)
	if err == nil {
		return &awsVol{backingType: backingTypeLocal, efsType: ""}, nil
	}

	return nil, fmt.Errorf("Volume Not Found")
}

func (v *awsDriver) pathForEFSVolume(name, perfmode string) string {
	if perfmode == perfmodeMaxIO {
		return filepath.Join(mountPointMaxIO, name)
	} else {
		return filepath.Join(mountPointRegular, name)
	}
}

func (v *awsDriver) createEFS(req volume.Request) error {
	mountpoint := mountPointRegular
	perfmode := req.Options["perfmode"]
	if perfmode == perfmodeMaxIO {
		mountpoint = mountPointMaxIO
	}

	path := filepath.Join(mountpoint, req.Name)
	if err := os.MkdirAll(path, 0755); err != nil {
		log.Error(err)
		return err
	}
	return nil
}

func (v *awsDriver) removeEFS(path string) error {
	return os.RemoveAll(path)
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

func (v *awsDriver) snapshotVolumesLoop() {
	for {
		time.Sleep(snapshotLoopInterval)
		v.snapshotVolumesCore()
	}
}

func (v *awsDriver) snapshotVolumesCore() {
	v.m.Lock()
	defer v.m.Unlock()
	/*
		1. Retrieve all volumes attached to instance
		2. For each volume:
			2.1. If no snapshots present, intiate snapshot creation.
			2.2. If snapshot present,
				2.2.1. If in pending state, skip
				2.2.2  If in done state, delete older snapshots and initiate fresh snapshot creation
	*/
	logctx := log.WithFields(log.Fields{
		"operation": "snapshotVolumes"})

	vols, err := v.getCloudstorVolumesByAttachment(v.instanceID)
	if err != nil {
		logctx.Error(fmt.Sprintf("Failed to get volumes attached to instance: %v", err))
		return
	}

	for _, vol := range vols {
		v.m.Unlock()
		time.Sleep(snapshotLoopIterationInterval)
		v.m.Lock()

		logctx.Debug(fmt.Sprintf("Examining volume: %v", vol))

		snaps, err := v.getSnapshotsOfEBS(*vol.VolumeId)
		if err != nil {
			logctx.Error(fmt.Sprintf("Failed to get snapshots for volume: %v", err))
			continue
		}

		t := time.Time{}
		var latestSnap *ec2.Snapshot
		for _, snap := range snaps {
			logctx.Debug(fmt.Sprintf("Snapshot obtained: %v", snap))
			if t.Before(*snap.StartTime) {
				t = *snap.StartTime
				latestSnap = snap
			}
		}

		snapcreate := &ec2.CreateSnapshotInput{
			VolumeId: aws.String(*vol.VolumeId),
			DryRun:   aws.Bool(false),
		}

		if len(snaps) == 0 {
			fmt.Println("Create Initial snapshot for volume")
			_, err := v.cl.CreateSnapshot(snapcreate)
			if err != nil {
				logctx.Error(fmt.Sprintf("Failed to create snapshot %v", err))
				continue
			}
		} else {
			if *latestSnap.State == ec2.SnapshotStatePending {
				logctx.Debug(fmt.Sprintf("Latest snapshot still in progress: %v", latestSnap))
				continue
			}
			for _, snap := range snaps {
				if (*latestSnap.SnapshotId == *snap.SnapshotId) && (*latestSnap.State == ec2.SnapshotStateCompleted) {
					logctx.Debug(fmt.Sprintf("Retain existing latest snapshot: %v", snap))
					continue
				}
				logctx.Debug(fmt.Sprintf("Delete old snapshot: %v", snap))

				snapdel := &ec2.DeleteSnapshotInput{
					SnapshotId: aws.String(*snap.SnapshotId),
					DryRun:     aws.Bool(false),
				}
				_, err = v.cl.DeleteSnapshot(snapdel)
				if err != nil {
					logctx.Error(fmt.Sprintf("Failed to delete snapshot %v", err))
					continue
				}
			}
			logctx.Debug("Create new latest snapshot for volume")
			_, err = v.cl.CreateSnapshot(snapcreate)
			if err != nil {
				logctx.Error(fmt.Sprintf("Failed to create snapshot %v", err))
				continue
			}
		}
	}
}

func (v *awsDriver) pathForEBSVolume(name string) string {
	return filepath.Join(mountPointEBS, name)
}

func (v *awsDriver) createEBS(req volume.Request) error {
	logctx := log.WithFields(log.Fields{
		"operation": "createEBS",
		"name":      req.Name,
		"options":   req.Options})

	vol, _ := v.getEBSByName(req.Name)
	if vol == nil {
		// volume does not exist at all
		logctx.Infof("Volume does not exist. Create fresh EBS")
		return v.createEBSNew(req)
	}
	if *vol.AvailabilityZone == v.az {
		// volume already exists in AZ
		logctx.Infof("Volume already exists in AZ")
		return nil
	}
	// volume exists in another AZ
	_, err := v.createEBSFromSnapshot(req.Name)
	return err
}

func (v *awsDriver) createEBSNew(req volume.Request) error {
	logctx := log.WithFields(log.Fields{
		"operation": "createNewEBS",
		"name":      req.Name,
		"options":   req.Options})

	szGB, err := strconv.ParseInt(req.Options["size"], 10, 64)
	if err != nil {
		return fmt.Errorf("Invalid volume size: %v", err)
	}

	var iops int64
	if req.Options["ebstype"] == volumeTypeProvisionedIOPS {
		iops, err = strconv.ParseInt(req.Options["iops"], 10, 64)
		if err != nil {
			return fmt.Errorf("Invalid volume IOPs specified for io1: %v", err)
		}
	}

	vol, err := v.createEBSCore(req.Name, req.Options["ebstype"], "", szGB, iops, false)
	if err != nil {
		logctx.Error(fmt.Sprintf("Failed to create volume in new AZ: %v", err))
		return err
	}

	logctx.Infof(fmt.Sprintf("Volume creation in new AZ succeeded: %v", vol))
	return nil
}

func (v *awsDriver) createEBSFromSnapshot(volname string) (*ec2.Volume, error) {
	/*
		For volume create in different AZ:
		1. Create a new snapshot of volume, wait to complete.
		2. Delete old snapshots.
		3. Restore from latest snapshots.
			3.1 Optional warmed mode: full restore using dd - todo
	*/

	logctx := log.WithFields(log.Fields{
		"operation": "createEBSFromSnapshot",
		"name":      volname})

	snap, err := v.snapEBS(volname)
	if err != nil {
		logctx.Error(fmt.Sprintf("Failed to create snapshot of volume in original AZ: %v", err))
		return nil, err
	}
	logctx.Infof(fmt.Sprintf("Snapshot created of volume in original AZ: %v", snap))

	vol, err := v.createEBSCore(volname, "", *snap.SnapshotId, 0, 0, false)
	if err != nil {
		logctx.Error(fmt.Sprintf("Failed to create volume in new AZ: %v", err))
		return nil, err
	}
	logctx.Infof(fmt.Sprintf("Volume creation in new AZ succeeded: %v", vol))

	err = v.detachEBS(*snap.VolumeId)
	if err != nil {
		logctx.Error(fmt.Sprintf("DetachVolume failed: %v", err))
		return nil, err
	}

	logctx.Debug("Delete old version of volume in original AZ")
	volDelete := &ec2.DeleteVolumeInput{
		DryRun:   aws.Bool(false),
		VolumeId: aws.String(*snap.VolumeId),
	}

	_, err = v.cl.DeleteVolume(volDelete)
	if err != nil {
		logctx.Error(fmt.Sprintf("DeleteVolume failed: %v", err))
		return nil, err
	}
	logctx.Infof("Old Volume deleted")

	snaps, err := v.getSnapshotsOfEBS(*snap.VolumeId)
	if err != nil {
		logctx.Error(fmt.Sprintf("Failed to get snapshots of old volume: %v", err))
	}

	for _, snap := range snaps {
		logctx.Infof(fmt.Sprintf("Delete snapshot of original volume: %v", snap))
		snapdel := &ec2.DeleteSnapshotInput{
			SnapshotId: aws.String(*snap.SnapshotId),
			DryRun:     aws.Bool(false),
		}
		_, err = v.cl.DeleteSnapshot(snapdel)
		if err != nil {
			logctx.Error(fmt.Sprintf("Failed to delete snapshot: %v", err))
			continue
		}
	}
	return vol, nil
}

func (v *awsDriver) mountEBS(req volume.MountRequest) (string, error) {
	logctx := log.WithFields(log.Fields{
		"operation": "mountEBS",
		"name":      req.Name})

	vol, err := v.getEBSByName(req.Name)
	if err != nil {
		logctx.Error(fmt.Sprintf("Failed to get volume details: %v", err))
		return "", err
	}

	var device string
	var devicePath string

	attached := false
	detach := false
	for _, attachment := range vol.Attachments {
		if (*attachment.InstanceId == v.instanceID) && (*attachment.State == ec2.AttachmentStatusAttached) {
			attached = true
			devicePath = *attachment.Device
		}
		if (*attachment.InstanceId != v.instanceID) && (*attachment.State != ec2.AttachmentStatusDetached) {
			detach = true
		}
	}

	if !attached {
		device, err = v.findUnusedDevice()
		devicePath = fmt.Sprintf("/dev/%s", device)
		if err != nil {
			logctx.Error(fmt.Sprintf("findUnusedDevice failed: %v", err))
			return "", err
		}

		if detach {
			if *vol.AvailabilityZone == v.az {
				// volume exists in same AZ
				if err := v.detachEBS(*vol.VolumeId); err != nil {
					logctx.Error(fmt.Sprintf("Failed to detach volume: %v", err))
					return "", err
				}
			} else {
				// volume exists in another AZ - transfer it to current AZ
				vol, err = v.createEBSFromSnapshot(req.Name)
				if err != nil {
					logctx.Error(fmt.Sprintf("Failed to transfer volume to az: %v", err))
					return "", err
				}
			}
		}

		if err := v.attachEBS(*vol.VolumeId, device); err != nil {
			logctx.Error(fmt.Sprintf("Failed to attach volume: %v", err))
			return "", err
		}

		logctx.Infof("Volume attached to %q", v.instanceID)
	}

	formatted, err := isExtFS(devicePath)
	if err != nil {
		logctx.Error(fmt.Sprintf("failed to probe volume FS: %v", err))
		return "", err
	}

	if !formatted {
		cmd := exec.Command("mkfs.ext4", "-F", devicePath)
		if out, err := cmd.CombinedOutput(); err != nil {
			logctx.Error(fmt.Sprintf("format failed: %v\noutput=%q", err, out))
			return "", err
		}
		logctx.Infof("Formatting succeeded for device: %q", devicePath)
	}

	path := v.pathForEBSVolume(req.Name)

	if err := os.MkdirAll(path, 0755); err != nil {
		logctx.Error(fmt.Sprintf("could not create mount point: %v", err))
		return "", err
	}

	mounted, err := isMounted(path)
	if err != nil {
		logctx.Error(fmt.Sprintf("could not evaluate mount points: %v", err))
		return "", err
	}

	if !mounted {
		cmd := exec.Command("mount", "-t", "ext4", "-v", devicePath, path)
		if out, err := cmd.CombinedOutput(); err != nil {
			logctx.Error(fmt.Sprintf("mount failed: %v\noutput=%q", err, out))
			return "", err
		}
		logctx.Infof("mount succeeded for dev: %q", devicePath)
	}

	return path, nil
}

func (v *awsDriver) removeEBS(req volume.Request) error {
	logctx := log.WithFields(log.Fields{
		"operation": "removeEBS",
		"name":      req.Name})

	vol, err := v.getEBSByName(req.Name)
	if err != nil {
		logctx.Error(fmt.Sprintf("Failed to get volume details: %v", err))
		return err
	}
	volumeID := *vol.VolumeId

	detach := false
	for _, attachment := range vol.Attachments {
		if *attachment.State != ec2.AttachmentStatusDetached {
			detach = true
		}
	}

	if detach {
		err := v.detachEBS(volumeID)
		if err != nil {
			logctx.Error(fmt.Sprintf("DetachVolume failed: %v", err))
			return err
		}
	}

	volDelete := &ec2.DeleteVolumeInput{
		DryRun:   aws.Bool(false),
		VolumeId: aws.String(volumeID),
	}

	_, err = v.cl.DeleteVolume(volDelete)
	if err != nil {
		logctx.Error(fmt.Sprintf("DeleteVolume failed: %v", err))
		return err
	}
	return nil
}

func (v *awsDriver) createEBSCore(volumeName, volumeType, snapshotID string, volumeSize, volumeIOPs int64, encrypted bool) (*ec2.Volume, error) {
	nametag := &ec2.TagSpecification{
		ResourceType: aws.String(ec2.ResourceTypeVolume),
		Tags:         []*ec2.Tag{},
	}

	nametag.Tags = append(
		nametag.Tags,
		&ec2.Tag{
			Key:   aws.String("CloudstorVolumeName"),
			Value: &volumeName,
		})

	nametag.Tags = append(
		nametag.Tags,
		&ec2.Tag{
			Key:   aws.String("StackID"),
			Value: &v.stackID,
		})

	volCreate := &ec2.CreateVolumeInput{
		AvailabilityZone:  aws.String(v.az),
		DryRun:            aws.Bool(false),
		TagSpecifications: []*ec2.TagSpecification{},
	}

	if snapshotID != "" {
		volCreate.SnapshotId = aws.String(snapshotID)
	}

	if volumeType != "" {
		volCreate.VolumeType = aws.String(volumeType)
	}

	if volumeSize != 0 {
		volCreate.Size = aws.Int64(int64(volumeSize))
	}

	if volumeIOPs != 0 {
		volCreate.Iops = aws.Int64(int64(volumeIOPs))
	}

	volCreate.TagSpecifications = append(
		volCreate.TagSpecifications,
		nametag)

	vol, err := v.cl.CreateVolume(volCreate)
	if err != nil {
		return nil, err
	}
	volID := *vol.VolumeId

	volExists := false
	for start := time.Now(); !volExists && time.Since(start) < createEBSTimeout; time.Sleep(retryInterval) {
		vol, err = v.getEBSByID(volID)
		if err != nil {
			continue
		}
		if *vol.State == ec2.VolumeStateAvailable {
			volExists = true
		}
	}

	if !volExists {
		err = fmt.Errorf("Volume never transitioned to Available state")
		return nil, err
	}
	return vol, nil
}

func (v *awsDriver) attachEBS(volumeID, device string) error {
	volAttach := &ec2.AttachVolumeInput{
		Device:     aws.String(fmt.Sprintf("/dev/%s", device)),
		InstanceId: aws.String(v.instanceID),
		VolumeId:   aws.String(volumeID),
		DryRun:     aws.Bool(false),
	}

	_, err := v.cl.AttachVolume(volAttach)
	if err != nil {
		return err
	}

	deviceExists := false
	for start := time.Now(); !deviceExists && time.Since(start) < attachEBSTimeout; time.Sleep(retryInterval) {
		if _, err := os.Stat(fmt.Sprintf("/sys/block/%s", device)); err == nil {
			deviceExists = true
		}
	}

	if !deviceExists {
		err = fmt.Errorf("Volume never attached to Instance")
		return err
	}

	return nil
}

func (v *awsDriver) detachEBS(volumeID string) error {
	volDetach := &ec2.DetachVolumeInput{
		DryRun:   aws.Bool(false),
		Force:    aws.Bool(true),
		VolumeId: aws.String(volumeID),
	}

	_, err := v.cl.DetachVolume(volDetach)
	if err != nil {
		return err
	}

	detached := false
	for start := time.Now(); !detached && time.Since(start) < detachEBSTimeout; time.Sleep(retryInterval) {
		vol, err := v.getEBSByID(volumeID)
		if err != nil {
			continue
		}
		detached = true
		for _, attachment := range vol.Attachments {
			if *attachment.State != ec2.AttachmentStatusDetached {
				detached = false
			}
		}
	}

	if !detached {
		err = fmt.Errorf("Volume never detached from Instance")
		return err
	}

	return nil
}

func (v *awsDriver) snapEBS(volumeName string) (*ec2.Snapshot, error) {
	logctx := log.WithFields(log.Fields{
		"operation": "snapEBS",
		"name":      volumeName})

	vol, err := v.getEBSByName(volumeName)
	if err != nil {
		err = fmt.Errorf("Failed to get volume ID for snapshot: %v", err)
		return nil, err
	}

	snapcreate := &ec2.CreateSnapshotInput{
		VolumeId: aws.String(*vol.VolumeId),
		DryRun:   aws.Bool(false),
	}

	snap, err := v.cl.CreateSnapshot(snapcreate)
	if err != nil {
		err = fmt.Errorf("Failed to create snapshot %v", err)
		return nil, err
	}

	logctx.Debug(fmt.Sprintf("Snapshot created: %v", snap))

	snapdescr := &ec2.DescribeSnapshotsInput{
		SnapshotIds: []*string{aws.String(*snap.SnapshotId)},
		DryRun:      aws.Bool(false),
	}

	snapComplete := false
	for start := time.Now(); !snapComplete && time.Since(start) < createSnapTimeout; time.Sleep(retryInterval) {
		resp, err := v.cl.DescribeSnapshots(snapdescr)
		if err != nil {
			continue
		}
		if len(resp.Snapshots) == 0 {
			continue
		}
		snap := resp.Snapshots[0]
		if *snap.State == ec2.SnapshotStateCompleted {
			snapComplete = true
		} else {
			logctx.Debug("Snapshot status: " + *snap.State)
		}
	}

	return snap, nil
}

func (v *awsDriver) getEBSByID(volumeId string) (*ec2.Volume, error) {
	input := &ec2.DescribeVolumesInput{
		DryRun: aws.Bool(false),
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("volume-id"),
				Values: []*string{aws.String(volumeId)},
			},
		},
		VolumeIds: []*string{
			aws.String(volumeId),
		},
	}

	resp, err := v.cl.DescribeVolumes(input)
	if err != nil {
		return nil, err
	}
	if len(resp.Volumes) == 0 {
		return nil, fmt.Errorf("Volume not found")
	}
	return resp.Volumes[0], nil
}

func (v *awsDriver) getEBSByName(volumeName string) (*ec2.Volume, error) {
	input := &ec2.DescribeVolumesInput{
		DryRun: aws.Bool(false),
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag:CloudstorVolumeName"),
				Values: []*string{aws.String(volumeName)},
			},
			{
				Name:   aws.String("tag:StackID"),
				Values: []*string{aws.String(v.stackID)},
			},
		},
	}

	resp, err := v.cl.DescribeVolumes(input)
	if err != nil {
		return nil, err
	}
	if len(resp.Volumes) == 0 {
		return nil, fmt.Errorf("Volume not found")
	}
	return resp.Volumes[0], nil
}

func (v *awsDriver) getEBSVolumesByStack() ([]*ec2.Volume, error) {
	input := &ec2.DescribeVolumesInput{
		DryRun: aws.Bool(false),
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("tag:StackID"),
				Values: []*string{aws.String(v.stackID)},
			},
		},
	}

	resp, err := v.cl.DescribeVolumes(input)
	if err != nil {
		return nil, err
	}
	return resp.Volumes, nil
}

func (v *awsDriver) getCloudstorVolumesByAttachment(instanceID string) ([]*ec2.Volume, error) {
	input := &ec2.DescribeVolumesInput{
		DryRun: aws.Bool(false),
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("attachment.instance-id"),
				Values: []*string{aws.String(instanceID)},
			},
			{
				Name:   aws.String("tag-key"),
				Values: []*string{aws.String("CloudstorVolumeName")},
			},
		},
	}

	resp, err := v.cl.DescribeVolumes(input)
	if err != nil {
		return nil, err
	}
	return resp.Volumes, nil
}

func (v *awsDriver) getSnapshotsOfEBS(volumeID string) ([]*ec2.Snapshot, error) {
	input := &ec2.DescribeSnapshotsInput{
		DryRun: aws.Bool(false),
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("volume-id"),
				Values: []*string{aws.String(volumeID)},
			},
		},
	}

	resp, err := v.cl.DescribeSnapshots(input)
	if err != nil {
		return nil, err
	}
	return resp.Snapshots, nil
}

func (v *awsDriver) findUnusedDevice() (string, error) {
	// according the AWS docs, recommended device letters are xvdf to xvdp
	ec2vbds := []string{"f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p"}
	for _, dev := range ec2vbds {
		devpath := fmt.Sprintf("/sys/block/xvd%s", dev)
		if _, err := os.Stat(devpath); os.IsNotExist(err) {
			return fmt.Sprintf("xvd%s", dev), nil
		}
	}
	return "", fmt.Errorf("All device names used!")
}
