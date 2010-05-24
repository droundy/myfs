package main

import (
	"fmt"
	"os"
	"bitbucket.org/taruti/go-extra/fuse"
)

type FileSystem struct {
	maxino uint64
	nodes map[uint64]*Inode
}

func NewFS() *FileSystem {
	x := FileSystem{0, make(map[uint64]*Inode)}
	x.MkDir(1, ".") // Make root directory...
	return &x
}

func (fs *FileSystem) MkDir(parent uint64, name string) *Inode {
	if parent != 1 {
		// Let's check that the parent actually exists...
		if _,ok := fs.nodes[parent]; !ok {
			panic("yikes, the parent doesn't exist!")
		}
	}
	var a = new(Inode);
	a.Ino = fs.maxino + 1
	fs.maxino = a.Ino
	a.Mode = fuse.S_IFDIR + 0755
	a.Uid = 1137
	a.Gid = 1137
	sec, nsec, _ := os.Time()
	a.Mtime = uint64(sec) - 3600
	a.Mtimensec = uint32(nsec)
	a.Ctime = uint64(sec) - 3600
	a.Ctimensec = uint32(nsec)
	a.DirContents = make(map[string]uint64)
	// Create . and .. entries here
	//a.DirContents["."] = a.Ino
	//a.DirContents[".."] = parent
	fs.nodes[a.Ino] = a;
	// Add entry to parent directoory
	if name != "." {
		fs.nodes[parent].DirContents[name] = a.Ino
	}
	return a
}

func (fs *FileSystem) Create(parent uint64, name string) *Inode {
	if _,ok := fs.nodes[parent]; !ok {
		panic("yikes, the parent doesn't exist!")
	}
	var a = new(Inode);
	a.Ino = fs.maxino + 1
	fs.maxino = a.Ino
	a.Mode = fuse.S_IFREG + 0644
	a.Uid = 1137
	a.Gid = 1137
	sec, nsec, _ := os.Time()
	a.Mtime = uint64(sec) - 3600
	a.Mtimensec = uint32(nsec)
	a.Ctime = uint64(sec) - 3600
	a.Ctimensec = uint32(nsec)
	a.FileContents = ""
	fs.nodes[a.Ino] = a;
	fs.nodes[parent].DirContents[name] = a.Ino
	return a
}

func (_ *FileSystem) Log(v ...interface{}) {
	fmt.Println(v)
}

func (fs *FileSystem) Dump(s string) {
	fmt.Println(s, "with maxino", fs.maxino)
	for ino, v := range fs.nodes {
		fmt.Println(ino, *v)
	}
}

func (fs *FileSystem) FuseLookup(h *fuse.Header, p []byte) (*fuse.Attr, fuse.Error) {
	fs.Log("FuseLookup", h.Nodeid, p)
	if par,ok := fs.nodes[h.Nodeid]; ok {
		if inum,ok := par.DirContents[string(p)]; ok {
			if i,ok := fs.nodes[inum]; ok {
				return i.ToAttr(), fuse.OK
			}
		}
	}
	return nil, fuse.ENOENT
}

func (fs *FileSystem) FuseForget(f *fuse.Forget) {
	fs.Log("FuseForget", f)
}

func (fs *FileSystem) FuseGetAttr(p *fuse.GetAttr) (*fuse.Attr, fuse.Error) {
	fs.Log("FuseGetAttr", p, "with Nodeid", p.Nodeid)
	if i,ok := fs.nodes[p.Nodeid]; ok {
		return i.ToAttr(), fuse.OK
	}
	return nil, fuse.ENOENT
}

func (fs *FileSystem) FuseMkDir(m *fuse.MkDir, p []byte) (*fuse.Attr, fuse.Error) {
	fs.Log("FuseMkDir", m, p)
	inew := fs.MkDir(m.Nodeid, string(p))
	inew.Uid, inew.Gid = uint32(m.Uid), uint32(m.Gid);
	fs.Dump("FuseMkDir")
	return inew.ToAttr(), fuse.OK
}

func (fs *FileSystem) FuseCreate(c *fuse.Create, p []byte) (flags fuse.Flags, fh fuse.FuseFile, a *fuse.Attr, e fuse.Error) {
	fs.Log("FuseCreate", c)
	inew := fs.Create(c.Nodeid, string(p))
	inew.Uid, inew.Gid = uint32(c.Uid), uint32(c.Gid);
	return c.Flags, inew, inew.ToAttr(), fuse.OK
}

func (fs *FileSystem) FuseOpen(o *fuse.Open) (flags fuse.Flags, fh fuse.FuseFile, e fuse.Error) {
	fs.Log("FuseOpen", o)
	flags =  o.Flags
	if i,ok := fs.nodes[o.Nodeid]; ok {
		return flags, i, fuse.OK
	}
	return 0, nil, fuse.ENOENT
}

func (fs *FileSystem) FuseGetXAttrSize(h *fuse.Header, p []byte) (int, fuse.Error) {
	fs.Log("FuseGetXAttrSize", h, p)
	return 0, fuse.OK
}

func (fs *FileSystem) FuseGetXAttr(h *fuse.Header, p []byte) ([]byte, fuse.Error) {
	fs.Log("FuseGetXAttr", h, p)
	return nil, fuse.OK
}

func (fs *FileSystem) FuseListXAttr(h *fuse.Header, l *fuse.XAttrList) fuse.Error {
	fs.Log("FuseListXAttr", h, l)
	return fuse.OK
}

func (fs *FileSystem) FuseInit(i *fuse.Init) fuse.Error {
	fs.Log("FuseInit", i)
	return fuse.OK
}

func (fs *FileSystem) FuseAccess(a *fuse.Access) fuse.Error {
	fs.Log("FuseAccess", a)
	return fuse.OK
}

func (fs *FileSystem) NilFile() fuse.FuseFile {
	fs.Log("NilFile")
	var i Inode
	return &i
}

type Inode struct {
	Ino uint64
	Atime, Mtime, Ctime uint64
	Atimensec, Mtimensec, Ctimensec uint32
	Mode uint32
	Nlink uint32
	Uid, Gid uint32
	Nreading, Nwriting uint32
	FileContents string
	DirContents map[string]uint64
}

func (i *Inode) Log(v ...interface{}) {
	fmt.Println("inode", i.Ino, ":", v)
}

func (i *Inode) IsDir() bool {
	return (i.Mode & (fuse.S_IFDIR | fuse.S_IFREG)) == fuse.S_IFDIR
}

func (i *Inode) FuseRead(io *fuse.Io) ([]byte, fuse.Error) {
	i.Log("FuseRead", io) // the following is wrong...
	i.Log("offset is", io.Offset)
	i.Log("size is", io.Size)
	if i.IsDir() {
		de := new(fuse.DEntryList)
		var num uint64 = 0
		for path,ino := range i.DirContents {
			i.Log("::: ",path,ino)
			if num >= io.Offset {
				de.AddString(path, ino, fuse.S_IFDIR)
				num += 1
			}
		}
		return de.Bytes(), fuse.OK
	}
	return []byte(i.FileContents[io.Offset:]), fuse.OK
}

func (i *Inode) FuseWrite(io *fuse.Io, d []byte) (uint64, fuse.Error) {
	i.Log("FuseWrite", io, string(d))
	i.FileContents = string(d) // wrong!
	return uint64(len(d)), fuse.OK
}

func (i *Inode) FuseRelease(r *fuse.Release) fuse.Error {
	i.Log("FuseRelease", r)
	if i.Nreading > 0 { i.Nreading -= 1 }
	if i.Nwriting > 0 { i.Nwriting -= 1 }
	if !i.IsDir() {
		i.Log("File contents:", i.FileContents)
	}
	return fuse.OK
}

func (i *Inode) FuseFlush(f *fuse.Flush) fuse.Error {
	i.Log("FuseFlush", f)
	return fuse.OK
}

func (i *Inode) ToAttr() *fuse.Attr {
	var a fuse.Attr
	a.Ino = i.Ino
	a.Size = uint64(len(i.FileContents))
	a.Atime, a.Mtime, a.Ctime = i.Atime, i.Mtime, i.Ctime
	a.Atimensec, a.Mtimensec, a.Ctimensec = i.Atimensec, i.Mtimensec, i.Ctimensec
	a.Mode = i.Mode
	a.Nlink = i.Nlink
	a.Uid, a.Gid = i.Uid, i.Gid
	return &a
}
