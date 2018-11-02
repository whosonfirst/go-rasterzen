// Code generated by go-bindata.
// sources:
// templates/html/index.html
// DO NOT EDIT!

package templates

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
	info  os.FileInfo
}

type bindataFileInfo struct {
	name    string
	size    int64
	mode    os.FileMode
	modTime time.Time
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
func (fi bindataFileInfo) IsDir() bool {
	return false
}
func (fi bindataFileInfo) Sys() interface{} {
	return nil
}

var _templatesHtmlIndexHtml = []byte("\x1f\x8b\x08\x00\x00\x00\x00\x00\x00\xff\x8c\x53\xc1\x8e\x9b\x30\x10\xbd\xef\x57\x8c\xa8\x2a\x12\x6d\x16\x87\x8d\x9a\x8d\x08\xa4\xea\xad\x55\xa3\xaa\x87\xaa\x87\x56\x55\x35\x1b\x4f\x62\x67\x0d\x58\xb6\x45\xc2\x22\xfe\xbd\x32\x90\x84\xbd\xf5\xe6\xe7\x79\x7e\xbc\x79\x33\xa4\xc2\xe5\x6a\x73\x07\x90\x0a\x42\xee\x0f\x00\xa9\x93\x4e\xd1\xc6\xa0\x75\x64\x78\xca\x7a\xd8\x95\xfa\xba\x92\xc5\x0b\x18\x52\x59\x60\x5d\xad\xc8\x0a\x22\x17\x80\x30\xb4\xcf\x02\xb6\xb3\x96\x29\xc2\xbd\x22\x17\xed\xac\x0d\x80\x0d\xaa\x76\x67\xa4\x76\x60\xcd\x2e\x0b\xd8\x11\x2b\xec\x2f\xae\xe4\xa3\x0d\x36\x29\xeb\x2f\xff\xef\xc9\x83\x40\x2b\xde\xbe\xbb\x9a\x1c\x04\xbc\x3f\x70\xb5\xa6\x2c\x70\x74\x76\xde\x5d\xd0\x8b\x03\x3c\x97\xbc\x86\x06\x72\x34\x07\x59\x24\x30\xd7\xe7\x35\x68\xe4\x5c\x16\x87\x01\xb5\x03\xf3\x5d\x8e\x1a\x1a\x38\x49\xee\x44\x12\xcf\xe7\xef\xd7\x20\x48\x1e\x84\xf3\xa0\x12\x37\x62\x74\x71\xe6\xa4\x22\x68\xe0\xb9\x34\x9c\x4c\x02\x1c\xad\x20\x0e\x86\x38\xc4\x37\xe1\x94\x75\xfe\xba\xf8\xd9\x25\x7f\x0f\xbc\xb3\x21\x02\x2e\x2b\x90\x3c\x0b\x72\xd4\xbe\x4d\x2e\xab\xf1\x24\xc6\x69\x01\x54\x68\x00\xb5\xfc\xfb\x42\x35\x64\x10\x36\x0d\x44\xdf\xe8\xec\x5e\xa9\xf8\xf4\xfd\xcb\x57\xaa\xa1\x6d\xc3\xf5\x88\xbb\x2f\x4d\x8e\xce\x53\x6d\x75\x08\xd7\x77\xa3\x92\x6f\x38\x83\x6d\x94\xa3\x9e\x84\x39\xea\x70\x1a\x59\x72\x3f\x25\x9d\x26\xbf\x17\x4f\xd1\x32\x5e\xac\x3e\xac\x66\xf0\x10\x3f\x3e\x46\x8b\xa7\xc5\x72\xf9\x67\x06\xf1\x62\x3a\x56\x57\x58\x93\xe9\x44\x7c\x18\x5b\x8f\x26\xa1\x70\x4e\x27\x8c\xa9\x72\x87\x4a\x94\xd6\x25\xab\xf9\x6a\xce\x42\xb8\xbf\x98\xb9\x87\x90\x35\xaf\x2d\x6b\xce\x2d\x6b\xea\x36\x3a\xda\xb2\xf8\x38\x34\x95\x79\xde\x70\x9e\x41\x93\xe3\xf9\x57\x59\xe6\x09\xc4\xcb\x76\x7a\x75\xdf\x7d\x36\x42\xce\x7f\x94\x93\x1c\xf5\xb5\xe0\x2d\xf9\x7d\x81\x0c\x0a\x3a\xc1\x36\xfa\x8c\x56\x0c\x0c\xb8\x25\xfa\x66\x01\xbb\xc1\xf4\xc3\x48\x59\xff\xa7\xfc\x0b\x00\x00\xff\xff\x29\x70\x69\xb3\x31\x03\x00\x00")

func templatesHtmlIndexHtmlBytes() ([]byte, error) {
	return bindataRead(
		_templatesHtmlIndexHtml,
		"templates/html/index.html",
	)
}

func templatesHtmlIndexHtml() (*asset, error) {
	bytes, err := templatesHtmlIndexHtmlBytes()
	if err != nil {
		return nil, err
	}

	info := bindataFileInfo{name: "templates/html/index.html", size: 817, mode: os.FileMode(420), modTime: time.Unix(1541182944, 0)}
	a := &asset{bytes: bytes, info: info}
	return a, nil
}

// Asset loads and returns the asset for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func Asset(name string) ([]byte, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("Asset %s can't read by error: %v", name, err)
		}
		return a.bytes, nil
	}
	return nil, fmt.Errorf("Asset %s not found", name)
}

// MustAsset is like Asset but panics when Asset would return an error.
// It simplifies safe initialization of global variables.
func MustAsset(name string) []byte {
	a, err := Asset(name)
	if err != nil {
		panic("asset: Asset(" + name + "): " + err.Error())
	}

	return a
}

// AssetInfo loads and returns the asset info for the given name.
// It returns an error if the asset could not be found or
// could not be loaded.
func AssetInfo(name string) (os.FileInfo, error) {
	cannonicalName := strings.Replace(name, "\\", "/", -1)
	if f, ok := _bindata[cannonicalName]; ok {
		a, err := f()
		if err != nil {
			return nil, fmt.Errorf("AssetInfo %s can't read by error: %v", name, err)
		}
		return a.info, nil
	}
	return nil, fmt.Errorf("AssetInfo %s not found", name)
}

// AssetNames returns the names of the assets.
func AssetNames() []string {
	names := make([]string, 0, len(_bindata))
	for name := range _bindata {
		names = append(names, name)
	}
	return names
}

// _bindata is a table, holding each asset generator, mapped to its name.
var _bindata = map[string]func() (*asset, error){
	"templates/html/index.html": templatesHtmlIndexHtml,
}

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
func AssetDir(name string) ([]string, error) {
	node := _bintree
	if len(name) != 0 {
		cannonicalName := strings.Replace(name, "\\", "/", -1)
		pathList := strings.Split(cannonicalName, "/")
		for _, p := range pathList {
			node = node.Children[p]
			if node == nil {
				return nil, fmt.Errorf("Asset %s not found", name)
			}
		}
	}
	if node.Func != nil {
		return nil, fmt.Errorf("Asset %s not found", name)
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
var _bintree = &bintree{nil, map[string]*bintree{
	"templates": &bintree{nil, map[string]*bintree{
		"html": &bintree{nil, map[string]*bintree{
			"index.html": &bintree{templatesHtmlIndexHtml, map[string]*bintree{}},
		}},
	}},
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
	err = os.Chtimes(_filePath(dir, name), info.ModTime(), info.ModTime())
	if err != nil {
		return err
	}
	return nil
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

