package main

import (
	"fmt"
	"bitbucket.org/taruti/go-extra/fuse"
)

type MyFile struct {}

func (_ *MyFile) Log(v ...interface{}) {
	fmt.Println(v)
}

func (mf *MyFile) FuseRead(io *fuse.Io) ([]byte, fuse.Error) {
	mf.Log("FuseRead", io)
	return []byte("hello world"), fuse.OK
}

func (mf *MyFile) FuseWrite(io *fuse.Io, d []byte) (uint64, fuse.Error) {
	mf.Log("FuseWrite", io, d)
	return 0, fuse.OK
}

func (mf *MyFile) FuseRelease(r *fuse.Release) fuse.Error {
	mf.Log("FuseRelease", r)
	return fuse.OK
}

func (mf *MyFile) FuseFlush(f *fuse.Flush) fuse.Error {
	mf.Log("FuseFlush", f)
	return fuse.OK
}

type MyFS struct {}

func (_ *MyFS) Log(v ...interface{}) {
	fmt.Println(v)
}

func (mfs *MyFS) FuseLookup(h *fuse.Header, p []byte) (*fuse.Attr, fuse.Error) {
	mfs.Log("FuseLookup", h, p)
	if string(p) == "hello" {
		return newFileAttr(h.Nodeid + 137), fuse.OK
	} else if string(p) == "hallo" {
		return newDirAttr(h.Nodeid + 1), fuse.OK
	}
	return nil, fuse.ENOENT
}

func (mfs *MyFS) FuseForget(f *fuse.Forget) {
	mfs.Log("FuseForget", f)
}

func newDirAttr(ino uint64) *fuse.Attr {
	a := new(fuse.Attr)
	a.Ino = ino
	a.Mode = fuse.S_IFDIR + 0755
	a.Uid = 1137
	a.Gid = 1137
	return a
}

func newFileAttr(ino uint64) *fuse.Attr {
	a := new(fuse.Attr)
	a.Ino = ino
	a.Mode = fuse.S_IFREG + 0644
	a.Size = 137
	a.Nlink = 1
	a.Uid = 1137
	a.Gid = 1137
	a.Mtime = 100000000
	return a
}

func (mfs *MyFS) FuseGetAttr(p *fuse.GetAttr) (*fuse.Attr, fuse.Error) {
	mfs.Log("FuseGetAttr", p)
	if p.Nodeid > 100 {
		return newFileAttr(p.Nodeid), fuse.OK
	} else {
		return newDirAttr(p.Nodeid), fuse.OK
	}
	var a fuse.Attr
  a.Ino = p.Nodeid
  a.Mode = fuse.S_IFDIR
  return &a, fuse.OK
}

func (mfs *MyFS) FuseMkDir(m *fuse.MkDir, p []byte) (*fuse.Attr, fuse.Error) {
	mfs.Log("FuseMkDir", m, p)
	return nil, fuse.OK
}

func (mfs *MyFS) FuseOpen(o *fuse.Open) (flags fuse.Flags, fh fuse.FuseFile, e fuse.Error) {
	mfs.Log("FuseOpen", o)
	flags =  o.Flags
	return 0, &Dir{}, fuse.OK
}

func (mfs *MyFS) FuseGetXAttrSize(h *fuse.Header, p []byte) (int, fuse.Error) {
	mfs.Log("FuseGetXAttrSize", h, p)
	return 0, fuse.OK
}

func (mfs *MyFS) FuseGetXAttr(h *fuse.Header, p []byte) ([]byte, fuse.Error) {
	mfs.Log("FuseGetXAttr", h, p)
	return nil, fuse.OK
}

func (mfs *MyFS) FuseListXAttr(h *fuse.Header, l *fuse.XAttrList) fuse.Error {
	mfs.Log("FuseListXAttr", h, l)
	return fuse.OK
}

func (mfs *MyFS) FuseInit(i *fuse.Init) fuse.Error {
	mfs.Log("FuseInit", i)
	return fuse.OK
}

func (mfs *MyFS) FuseAccess(a *fuse.Access) fuse.Error {
	mfs.Log("FuseAccess", a)
	return fuse.OK
}

func (mfs *MyFS) NilFile() fuse.FuseFile {
	mfs.Log("NilFile")
	return &MyFile{}
}


type Dir struct{}

func (*Dir) FuseFlush(*fuse.Flush) fuse.Error { return fuse.OK }
func (*Dir) FuseRead(r *fuse.Io) ([]uint8, fuse.Error) {
	offset := r.Offset
	fmt.Printf("read offset = %d new\n", offset)
	if offset > 0 {
		return nil, fuse.OK
	} // FIXME

	de := new(fuse.DEntryList)
	if r.Nodeid < 10 {
		de.AddString("hello", r.Nodeid+137, fuse.S_IFREG)
		de.AddString("hallo", r.Nodeid+1, fuse.S_IFDIR)
	}
	return de.Bytes(), fuse.OK
}
func (*Dir) FuseRelease(*fuse.Release) fuse.Error {
	return fuse.OK
}

func (*Dir) FuseWrite(*fuse.Io, []uint8) (uint64, fuse.Error) {
	return 0, fuse.EPERM
}
