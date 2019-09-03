// Code generated by go-bindata. DO NOT EDIT.
// sources:
// cats_dogs_pats.dat
// cats_dogs.wts

package main


import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func bindataRead(data []byte, name string) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}

	var buf bytes.Buffer
	_, err = io.Copy(&buf, gz)
	clErr := gz.Close()

	if err != nil {
		return nil, fmt.Errorf("Read %q: %v", name, err)
	}
	if clErr != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}


type asset struct {
	bytes []byte
	info  fileInfoEx
}

type fileInfoEx interface {
	os.FileInfo
	MD5Checksum() string
}

type bindataFileInfo struct {
	name        string
	size        int64
	mode        os.FileMode
	modTime     time.Time
	md5checksum string
}

func (fi bindataFileInfo) Name() string {
	return fi.name
}
func (fi bindataFileInfo) Size() int64 {
	return fi.size
}
func (fi bindataFileInfo) Mode() os.FileMode {
	return fi.mode
}
func (fi bindataFileInfo) ModTime() time.Time {
	return fi.modTime
}
func (fi bindataFileInfo) MD5Checksum() string {
	return fi.md5checksum
}
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _bindataCatsdogspatsdat = []byte(
	"\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xac\x91\x3d\x4b\xc6\x30\x14\x85\xe7\xbc\xbf\xc2\xc1\x77\xeb\x90\x9b\x7e" +
	"\x1a\x24\x8b\x5a\xec\x22\x82\x6e\x12\x4a\xb1\x19\x02\x9a\x94\x1a\x0b\xf5\xd7\x4b\x04\xe1\xe6\x36\x6e\x52\x28\x3c" +
	"\xcf\x3d\x3d\x1c\xe8\x78\x2f\xd9\xe5\xe3\x14\x82\x59\xdd\xc3\xf4\x6e\xd8\x39\xbe\x5f\x84\xe4\x05\xd7\xd7\x42\x42" +
	"\x01\x5c\x21\x09\x1a\x81\xc0\x50\x62\xa8\x30\xd4\x18\x1a\x0c\x2d\x86\x0e\xc3\x95\x66\xe7\x61\x36\x2e\xd8\xb0\x1f" +
	"\xd6\x24\x07\xa0\x49\x41\x45\x49\x45\x45\x45\x4d\x45\x43\x45\x4b\x45\x47\x45\x5c\x7c\xe3\xdf\xfc\x8a\xe7\x56\x0a" +
	"\x4b\x48\x22\x22\xa1\x38\xb2\x9f\x36\xbf\xda\x60\x7a\xef\x67\xda\x72\xb8\x41\xee\x03\x91\x93\xb1\xfa\xc9\x7e\x25" +
	"\x7f\xb5\x54\xc8\x01\x0e\xe0\x8a\x67\xbf\xff\xb5\xe3\xf7\x04\x99\x78\xae\xe2\x67\xc4\x62\x5e\xad\xf9\xc0\x95\x42" +
	"\xa5\x1a\xf4\x69\xbc\x95\x6c\x70\xcb\x67\xb8\xb8\xdb\x8c\x0b\x0c\x18\xff\x9f\xe7\xf4\x1d\x00\x00\xff\xff\xb5\x49" +
	"\x51\xd5\xed\x02\x00\x00")

func bindataCatsdogspatsdatBytes() ([]byte, error) {
	return bindataRead(
		_bindataCatsdogspatsdat,
		"cats_dogs_pats.dat",
	)
}



func bindataCatsdogspatsdat() (*asset, error) {
	bytes, err := bindataCatsdogspatsdatBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{
		name: "cats_dogs_pats.dat",
		size: 749,
		md5checksum: "",
		mode: os.FileMode(420),
		modTime: time.Unix(1566980700, 0),
	}

	a := &asset{bytes: bytes, info: info}

	return a, nil
}

var _bindataCatsdogswts = []byte(
	"\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\xec\x59\x5d\x6b\xdb\x30\x14\x7d\x96\x7f\x85\xd0\xb3\x28\x96\x3f\x62\x27" +
	"\x6f\xa1\x25\x63\xb0\x86\xb2\x0c\xf6\x50\xf2\x60\x12\x53\xb2\xb5\xf1\x70\x4c\x47\x57\xf2\xdf\x87\x13\x27\x76\x6c" +
	"\x39\xfa\x76\xd2\x62\x08\x6d\x7d\x2b\xce\xd1\x95\xee\x39\xf1\x95\xde\x2d\x80\xa6\x71\xf6\x37\x49\x7f\xa3\x11\x44" +
	"\xb7\x51\xb6\x19\xaf\x97\x77\xc9\xd3\x06\x61\x0b\xa0\x6f\xd1\x5b\x9c\x6e\xd0\x08\x3e\x5a\x00\xbc\x5b\x00\x14\xa1" +
	"\x7c\xec\x34\x7a\x89\xf3\x41\x00\xa0\xfb\x38\x8b\xee\xa2\x2c\x42\x23\xb8\x1b\x04\xd0\x78\x91\xdd\x8f\x5f\x9f\xf2" +
	"\x71\xf6\x0d\xf1\xf7\xe3\x76\xe1\x87\x6a\x38\x8f\x6e\xf7\x18\x0f\xe9\xaf\xf5\x81\xa9\xe0\x02\x00\x4d\xd2\xe4\x25" +
	"\x1f\xfd\x75\x19\xaf\xb3\x55\xf6\x56\x00\x51\x28\x01\x40\x5f\x66\x8b\xe8\x39\xce\x87\x13\xb4\x8f\x6d\x0f\xc3\xbf" +
	"\x97\xd0\x47\xf0\x3c\xbc\x42\x23\x68\xe3\xe3\xf3\x14\x8d\x20\x29\x1f\x67\xf9\xbf\x1f\xa1\x0d\xe7\x65\xec\x67\xb6" +
	"\x8b\x11\x38\x2f\x42\x07\x8e\x3a\x2c\x61\xc1\x12\x29\x58\x87\x05\xeb\x48\xc1\xba\x2c\x58\x57\x0a\xd6\x63\xc1\x7a" +
	"\x52\xb0\x3e\x0b\xd6\x97\x82\x1d\xb0\x60\x07\x52\xb0\x01\x0b\x36\x90\x82\x0d\x59\xb0\xa1\x14\xec\x90\x05\x3b\x3c" +
	"\x0f\xbb\xff\xbd\x7f\xdc\x3d\xe4\x7f\xee\xc8\x6a\x06\x72\xaa\xe9\x0e\x4c\xa4\x74\xac\xde\x40\xce\xc0\xf6\x06\xd2" +
	"\x1b\xc8\xd5\x18\x08\xa6\x2a\xf9\x36\x79\x4e\xd2\x0e\xa4\xec\x35\xa5\x8c\x21\xc1\xd0\xc1\xd4\x72\xb6\x71\xf1\x11" +
	"\x16\xb7\x28\xd1\x8d\x8f\x8b\x1f\x38\x77\x17\x31\xc9\x9b\x25\x73\xbb\x5a\x42\x4f\x95\x88\x70\x66\xe4\x2b\x11\x91" +
	"\x43\x46\x6c\xa2\x81\x86\xa5\x23\x3c\x44\x41\x97\x05\x11\x76\x49\x36\xd4\xbe\x84\x1c\x76\x34\x89\x5e\x93\x74\x95" +
	"\xc5\x93\x24\x59\x5e\xa5\x2b\xf1\x56\xba\x9a\x2b\x09\x54\xba\xa2\x23\xf1\x67\xa4\xc3\x8d\xb8\x24\xf5\x41\xdc\x48" +
	"\x20\x23\x1d\x6e\xc4\x65\xe4\x8a\x6e\xc4\x9f\x91\xa2\x13\xf1\x13\xe9\x70\xa1\xd3\xa5\xe3\x70\xa1\xd9\xea\x5f\x17" +
	"\xed\x8d\xdb\x96\x4c\xbb\x23\x88\xfa\x8e\x09\x0a\xc7\x3c\x85\x2b\x4f\x21\xe9\x32\x26\x28\x7c\x25\x0a\x2e\xc1\x0f" +
	"\xcc\x67\x11\x98\xa7\x08\xcd\x53\x0c\x35\xee\x05\x8f\x83\xfc\x89\x17\xab\x78\xd3\x81\x89\x38\xd4\x4c\x5a\xb4\x27" +
	"\x6a\x1f\x7a\xc1\x1d\x93\xe0\xae\x49\x70\xcf\x24\xb8\x2f\x09\x2e\x63\x10\x7a\xc1\x03\x93\xe0\xa1\x49\xf0\xa1\x16" +
	"\x70\x0e\x23\x38\x34\x34\x3f\x92\x2e\x6e\x5c\xcc\xb5\x19\x8a\xa7\x2c\xfc\x2f\xe5\x6a\xfd\x8c\x40\x46\x6a\xfd\x8c" +
	"\x00\x51\xdf\xcf\xd0\x88\xfa\x7e\xc6\x44\x3f\x63\x9d\xb9\x25\xaa\x1c\xf5\x7e\xb2\x7b\x66\xfb\xcc\x2a\x62\xe8\x61" +
	"\xe8\x63\x38\xc0\x30\xc0\x30\xc4\xd4\x33\x73\x1b\xd7\x4e\xc7\x8e\xa2\x2b\xa3\xc2\xb7\x4c\xda\x67\x55\x7a\x8e\xc0" +
	"\xac\xea\x97\x54\x5a\x66\x75\xfa\x21\xfc\x82\xae\xdf\x6d\x29\xcf\x86\x34\x26\x61\x53\x9c\x99\x5b\x23\xcd\xf3\xc7" +
	"\x5e\x2a\x8d\xed\xa7\x2c\xb5\xdc\x5b\x84\xae\xed\x27\xd4\x0a\xb8\xa0\x34\x6a\xd3\x92\x69\x60\xba\x12\x2a\xb7\x34" +
	"\xca\x43\xb1\x5e\x12\x8d\x22\x24\x6d\xaa\xb8\xcc\xb7\x46\xb5\xf0\x70\x65\x7e\x17\xff\xb6\xa8\x2e\x92\x54\x11\x56" +
	"\xcf\x55\xfa\x3a\x6c\xa9\x43\x4a\x35\x5e\xb0\x0e\xed\xd3\x22\x24\x72\x3b\xdf\x68\xa4\xfb\xdd\x6f\xec\xbe\xdd\x66" +
	"\x44\x17\xd9\xfd\xd6\x57\xb3\x2b\x71\xa1\x0f\xf9\xc5\x6c\x81\xb9\xb5\xb5\xfe\x07\x00\x00\xff\xff\xb0\x47\x9b\xf4" +
	"\x6d\x2c\x00\x00")

func bindataCatsdogswtsBytes() ([]byte, error) {
	return bindataRead(
		_bindataCatsdogswts,
		"cats_dogs.wts",
	)
}



func bindataCatsdogswts() (*asset, error) {
	bytes, err := bindataCatsdogswtsBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{
		name: "cats_dogs.wts",
		size: 11373,
		md5checksum: "",
		mode: os.FileMode(420),
		modTime: time.Unix(1566980796, 0),
	}

	a := &asset{bytes: bytes, info: info}

	return a, nil
}


//
// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
//
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
}

//
// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
// nolint: deadcode
//
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

//
// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or could not be loaded.
//
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, &os.PathError{Op: "open", Path: name, Err: os.ErrNotExist}
}

//
// AssetNames returns the names of the assets.
// nolint: deadcode
//
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

//
// _bindata is a table, holding each asset generator, mapped to its name.
//
var _bindata = map[string]func() (*asset, error){
	"cats_dogs_pats.dat": bindataCatsdogspatsdat,
	"cats_dogs.wts":      bindataCatsdogswts,
}

//
// AssetDir returns the file names below a certain
// directory embedded in the file by go-bindata.
// For example if you run go-bindata on data/... and data contains the
// following hierarchy:
//     data/
//       foo.txt
//       img/
//         a.png
//         b.png
// then AssetDir("data") would return []string{"foo.txt", "img"}
// AssetDir("data/img") would return []string{"a.png", "b.png"}
// AssetDir("foo.txt") and AssetDir("notexist") would return an error
// AssetDir("") will return []string{"data"}.
//
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, &os.PathError{
					Op: "open",
					Path: name,
					Err: os.ErrNotExist,
				}
			}
		}
	}
	if node.Func != nil {
		return nil, &os.PathError{
			Op: "open",
			Path: name,
			Err: os.ErrNotExist,
		}
	}
	rv := make([]string, 0, len(node.Children))
	for childName := range node.Children {
		rv = append(rv, childName)
	}
	return rv, nil
}


type bintree struct {
	Func     func() (*asset, error)
	Children map[string]*bintree
}

var _bintree = &bintree{Func: nil, Children: map[string]*bintree{
	"cats_dogs.wts": {Func: bindataCatsdogswts, Children: map[string]*bintree{}},
	"cats_dogs_pats.dat": {Func: bindataCatsdogspatsdat, Children: map[string]*bintree{}},
}}

// RestoreAsset restores an asset under the given directory
func RestoreAsset(dir, name string) error {
	data, err := Asset(name)
	if err != nil {
		return err
	}
	info, err := AssetInfo(name)
	if err != nil {
		return err
	}
	err = os.MkdirAll(_filePath(dir, filepath.Dir(name)), os.FileMode(0755))
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(_filePath(dir, name), data, info.Mode())
	if err != nil {
		return err
	}
	return os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
}

// RestoreAssets restores an asset under the given directory recursively
func RestoreAssets(dir, name string) error {
	children, err := AssetDir(name)
	// File
	if err != nil {
		return RestoreAsset(dir, name)
	}
	// Dir
	for _, child := range children {
		err = RestoreAssets(dir, filepath.Join(name, child))
		if err != nil {
			return err
		}
	}
	return nil
}

func _filePath(dir, name string) string {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	return filepath.Join(append([]string{dir}, strings.Split(cannonicalName, "/")...)...)
}