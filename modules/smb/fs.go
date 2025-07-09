package smb

import (
	"context"
	"os"

	"github.com/hirochachacha/go-smb2"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/webdav"
)

type FS struct{ Share string }

type smbFile struct {
	*smb2.File
	share *smb2.Share
}

func (f *smbFile) Close() error {
	logrus.Debug("Closing file and unmounting share")
	err := f.File.Close()
	f.share.Umount()
	return err
}

func (fs FS) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	logrus.Debugf("Mkdir called: %s", name)
	sess := ctx.Value("smbSess").(*smb2.Session)
	share, err := sess.Mount(fs.Share)
	if err != nil {
		logrus.Errorf("SMB open error: %v", err)
		return err
	}
	defer share.Umount()
	return share.Mkdir(name, perm)
}

func (fs FS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	logrus.Debugf("OpenFile called: %s", name)
	sess := ctx.Value("smbSess").(*smb2.Session)
	share, err := sess.Mount(fs.Share)
	if err != nil {
		logrus.Errorf("SMB open error: %v", err)
		return nil, err
	}
	file, err := share.OpenFile(name, flag, perm)
	if err != nil {
		share.Umount()
		logrus.Errorf("SMB open file error: %v", err)
		return nil, err
	}
	return &smbFile{File: file, share: share}, nil
}

func (fs FS) RemoveAll(ctx context.Context, name string) error {
	logrus.Debugf("RemoveAll called: %s", name)
	sess := ctx.Value("smbSess").(*smb2.Session)
	share, err := sess.Mount(fs.Share)
	if err != nil {
		logrus.Errorf("SMB remove error: %v", err)
		return err
	}
	defer share.Umount()
	return share.Remove(name)
}

func (fs FS) Rename(ctx context.Context, oldName, newName string) error {
	logrus.Debugf("Rename called: %s -> %s", oldName, newName)
	sess := ctx.Value("smbSess").(*smb2.Session)
	share, err := sess.Mount(fs.Share)
	if err != nil {
		logrus.Errorf("SMB rename error: %v", err)
		return err
	}
	defer share.Umount()
	return share.Rename(oldName, newName)
}

func (fs FS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	logrus.Debugf("Stat called: %s", name)
	sess := ctx.Value("smbSess").(*smb2.Session)
	share, err := sess.Mount(fs.Share)
	if err != nil {
		logrus.Errorf("SMB stat error: %v", err)
		return nil, err
	}
	defer share.Umount()
	return share.Stat(name)
}

func (fs FS) ReadDir(ctx context.Context, name string) ([]os.FileInfo, error) {
	logrus.Infof("ReadDir called: %s", name)
	sess := ctx.Value("smbSess").(*smb2.Session)
	share, err := sess.Mount(fs.Share)
	if err != nil {
		logrus.Errorf("SMB readdir error: %v", err)
		return nil, err
	}
	defer share.Umount()
	return share.ReadDir(name)
}
