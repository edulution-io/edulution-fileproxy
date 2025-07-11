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
	logrus.Debug("[SMB] Closing file and unmounting share")
	err := f.File.Close()
	f.share.Umount()
	return err
}

func (fs FS) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	logrus.Debugf("[SMB] Mkdir: %s", name)
	sess := ctx.Value("smbSess").(*smb2.Session)
	share, err := sess.Mount(fs.Share)
	if err != nil {
		logrus.Errorf("SMB open error: %v", err)
		return err
	}
	defer share.Umount()

	res_err := share.Mkdir(name, perm)
	if res_err != nil {
		logrus.Errorf("SMB mkdir error: %v", res_err)
	}

	logrus.Debugf("[SMB] mkdir %s successful!", name)
	return res_err
}

func (fs FS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	logrus.Debugf("[SMB] OpenFile: %s", name)
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
	logrus.Debugf("[SMB] RemoveAll: %s", name)
	sess := ctx.Value("smbSess").(*smb2.Session)
	share, err := sess.Mount(fs.Share)
	if err != nil {
		logrus.Errorf("SMB remove error: %v", err)
		return err
	}
	defer share.Umount()

	res_err := share.Remove(name)
	if res_err != nil {
		logrus.Errorf("SMB remove error: %v", res_err)
	}

	logrus.Debugf("[SMB] remove %s successful!", name)
	return res_err
}

func (fs FS) Rename(ctx context.Context, oldName, newName string) error {
	logrus.Debugf("[SMB] Rename: %s -> %s", oldName, newName)
	sess := ctx.Value("smbSess").(*smb2.Session)
	share, err := sess.Mount(fs.Share)
	if err != nil {
		logrus.Errorf("SMB rename error: %v", err)
		return err
	}
	defer share.Umount()

	res_err := share.Rename(oldName, newName)
	if res_err != nil {
		logrus.Errorf("SMB rename error: %v", res_err)
	}

	logrus.Debugf("[SMB] rename %s to %s successful!", oldName, newName)
	return res_err
}

func (fs FS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	logrus.Debugf("[SMB] Stat: %s", name)
	sess := ctx.Value("smbSess").(*smb2.Session)
	share, err := sess.Mount(fs.Share)
	if err != nil {
		logrus.Errorf("SMB stat error: %v", err)
		return nil, err
	}
	defer share.Umount()

	stat, res_err := share.Stat(name)
	if res_err != nil {
		logrus.Errorf("SMB stat error: %v", res_err)
	}

	logrus.Debugf("[SMB] stat %s successful!", name)
	return stat, res_err
}

func (fs FS) ReadDir(ctx context.Context, name string) ([]os.FileInfo, error) {
	logrus.Infof("[SMB] ReadDir: %s", name)
	sess := ctx.Value("smbSess").(*smb2.Session)
	share, err := sess.Mount(fs.Share)
	if err != nil {
		logrus.Errorf("SMB readdir error: %v", err)
		return nil, err
	}
	defer share.Umount()
	readdir, res_err := share.ReadDir(name)
	if res_err != nil {
		logrus.Errorf("SMB readdir error: %v", res_err)
	}

	logrus.Debugf("[SMB] readdir %s successful!", name)
	return readdir, res_err
}
