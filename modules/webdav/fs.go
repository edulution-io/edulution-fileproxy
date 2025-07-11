package webdav

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/edulution-io/edulution-fileproxy/modules/smb"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/webdav"
)

type RouterFS struct {
	Prefix string
	Shares map[string]smb.FS
}

type virtDir struct {
	name string
}

func (v *virtDir) Name() string       { return v.name }
func (v *virtDir) Size() int64        { return 0 }
func (v *virtDir) Mode() os.FileMode  { return os.ModeDir | 0555 }
func (v *virtDir) ModTime() time.Time { return time.Now() }
func (v *virtDir) IsDir() bool        { return true }
func (v *virtDir) Sys() interface{}   { return nil }

func (r *RouterFS) Stat(ctx context.Context, name string) (os.FileInfo, error) {
	logrus.Debugf("[WEBDAV] Stat: %s", name)

	clean := "/" + trimSlashes(name)

	if clean == "/" {
		logrus.Debugf("> Root requested! Returning shares...")
		return virtualFileInfo{name: clean[1:]}, nil
	}

	shareName, path := splitPath(clean)
	fs, ok := r.Shares[shareName]
	if !ok {
		return nil, os.ErrNotExist
	}
	return fs.Stat(ctx, path)
}

func (r *RouterFS) ReadDir(ctx context.Context, name string) ([]os.FileInfo, error) {
	logrus.Debugf("[WEBDAV] ReadDir: %s", name)

	clean := "/" + trimSlashes(name)

	if clean == "/" {
		logrus.Debugf("> Root requested! Returning shares...")
		out := make([]os.FileInfo, 0, len(r.Shares))
		for share := range r.Shares {
			out = append(out, &virtDir{name: share})
		}
		return out, nil
	}

	shareName, path := splitPath(clean)
	fs, ok := r.Shares[shareName]
	if !ok {
		return nil, os.ErrNotExist
	}
	return fs.ReadDir(ctx, path)
}

func (r *RouterFS) OpenFile(ctx context.Context, name string, flag int, perm os.FileMode) (webdav.File, error) {
	logrus.Debugf("[WEBDAV] OpenFile: %s", name)

	clean := "/" + trimSlashes(name)

	if clean == "/" {
		logrus.Debugf("> Root requested! Returning shares...")
		return &virtualDir{
			info:    virtualFileInfo{name: clean[1:]},
			content: r.Shares,
		}, nil
	}

	shareName, path := splitPath(clean)

	fs, ok := r.Shares[shareName]
	if !ok {
		return nil, os.ErrNotExist
	}
	return fs.OpenFile(ctx, path, flag, perm)
}

func (r *RouterFS) Mkdir(ctx context.Context, name string, perm os.FileMode) error {
	logrus.Debugf("[WEBDAV] Mkdir: %s", name)

	clean := "/" + trimSlashes(name)
	shareName, path := splitPath(clean)

	if path == "" {
		logrus.Debugf("> Root requested! Returning permission denied!")
		return os.ErrPermission
	}

	fs, ok := r.Shares[shareName]
	if !ok {
		return os.ErrNotExist
	}
	return fs.Mkdir(ctx, path, perm)
}

func (r *RouterFS) RemoveAll(ctx context.Context, name string) error {
	logrus.Debugf("[WEBDAV] RemoveAll: %s", name)

	clean := "/" + trimSlashes(name)
	shareName, path := splitPath(clean)

	if path == "" {
		logrus.Debugf("> Root requested! Returning permission denied!")
		return os.ErrPermission
	}

	fs, ok := r.Shares[shareName]
	if !ok {
		return os.ErrNotExist
	}
	return fs.RemoveAll(ctx, path)
}

func (r *RouterFS) Rename(ctx context.Context, oldName, newName string) error {
	logrus.Debugf("[WEBDAV] Rename: %s to %s", oldName, newName)

	oldClean := "/" + trimSlashes(oldName)
	newClean := "/" + trimSlashes(newName)

	oldShare, oldPath := splitPath(oldClean)
	newShare, newPath := splitPath(newClean)

	if oldPath == "" || newPath == "" {
		logrus.Debugf("> Root requested! Returning permission denied!")
		return os.ErrPermission
	}

	if oldShare != newShare {
		return errors.New("cross-share rename not supported")
	}
	fs, ok := r.Shares[oldShare]
	if !ok {
		return os.ErrNotExist
	}
	return fs.Rename(ctx, oldPath, newPath)
}

func trimSlashes(s string) string {
	return strings.Trim(s, "/")
}

func splitPath(path string) (string, string) {
	trimmed := strings.Trim(path, "/")
	parts := strings.SplitN(trimmed, "/", 2)

	if len(parts) == 1 {
		return parts[0], ""
	} else if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "", ""
}
