package webdav

import (
	"io"
	"os"
	"time"

	"github.com/edulution-io/edulution-fileproxy/modules/smb"
)

// #####################################################################################

type virtualFileInfo struct {
	name string
}

func (f virtualFileInfo) Name() string {
	return f.name

}
func (f virtualFileInfo) Size() int64 {
	return 0
}
func (f virtualFileInfo) Mode() os.FileMode {
	return os.ModeDir | 0755
}
func (f virtualFileInfo) ModTime() time.Time {
	return time.Now()
}
func (f virtualFileInfo) IsDir() bool {
	return true
}
func (f virtualFileInfo) Sys() any {
	return nil
}

// #####################################################################################

type virtualDir struct {
	info    os.FileInfo
	content map[string]smb.FS
}

func (v *virtualDir) Read(p []byte) (int, error) {
	return 0, io.EOF
}

func (v *virtualDir) Write(p []byte) (int, error) {
	return 0, os.ErrPermission
}

func (v *virtualDir) Seek(offset int64, whence int) (int64, error) {
	return 0, nil
}

func (v *virtualDir) Close() error {
	return nil
}

func (v *virtualDir) Stat() (os.FileInfo, error) {
	return v.info, nil
}

func (v *virtualDir) Readdir(count int) ([]os.FileInfo, error) {
	var infos []os.FileInfo
	for name := range v.content {
		infos = append(infos, virtualFileInfo{name: name})
	}

	return infos, nil
}
