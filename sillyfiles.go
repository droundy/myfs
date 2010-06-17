package main

import (
	"fmt"
	"os"
	"syscall"
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
	a.Nlink = 1
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
	a.Nlink = 1
	sec, nsec, _ := os.Time()
	a.Mtime = uint64(sec) - 3600
	a.Mtimensec = uint32(nsec)
	a.Ctime = uint64(sec) - 3600
	a.Ctimensec = uint32(nsec)
	a.FileContents = make([]byte, 0, 256)
	fs.nodes[a.Ino] = a;
	fs.nodes[parent].DirContents[name] = a.Ino
	return a
}

func (fs *FileSystem) Delete(parent uint64, name string) fuse.Error {
	if _,ok := fs.nodes[parent]; !ok {
		panic("yikes, the parent doesn't exist!")
	}
	if ino, ok := fs.nodes[parent].DirContents[name]; ok {
		if len(fs.nodes[ino].DirContents) > 0 {
			return syscall.ENOTEMPTY // is this right?
		}
		fs.nodes[parent].DirContents[name] = ino, false // delete entry
		fs.nodes[ino].Nlink -= 1  // uncount this entry
		if fs.nodes[ino].Nlink <= 0 {
			// Trash entry when last link is removed
			fs.nodes[ino] = fs.nodes[ino], false
		}
		return fuse.OK
	}
	return fuse.ENOENT
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

func (fs *FileSystem) FuseDelete(h *fuse.Header, name []byte) fuse.Error {
	fs.Log("FuseDelete", h.Nodeid, name)
	return fs.Delete(h.Nodeid, string(name))
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
	if c.Flags & os.O_WRONLY != 0 {
		fs.Log("O_WRONLY");
	} else if c.Flags & os.O_RDONLY != 0 {
		fs.Log("O_RDONLY");
	}	else if c.Flags & os.O_RDWR != 0 {
		fs.Log("O_RDWR");
	}
	if c.Flags & os.O_NONBLOCK != 0 {
		fs.Log("O_NONBLOCK");
	}
	if c.Flags & os.O_EXCL != 0 {
		fs.Log("O_EXCL");
	}
	// O_CREAT is always present in FuseCreate...
	inew := fs.Create(c.Nodeid, string(p))
	inew.Mode = fuse.S_IFREG | c.Mode
	inew.HandleFlags(c.Flags) // Handle O_TRUNC
	inew.Uid, inew.Gid = uint32(c.Uid), uint32(c.Gid);
	return c.Flags, &OpenInode{inew,0}, inew.ToAttr(), fuse.OK
}

func (fs *FileSystem) FuseOpen(o *fuse.Open) (flags fuse.Flags, fh fuse.FuseFile, e fuse.Error) {
	fs.Log("FuseOpen", o)
	flags =  o.Flags
	fs.Log("open flags are", o.Flags)
	if o.Flags & os.O_WRONLY != 0 {
		fs.Log("O_WRONLY");
	} else if o.Flags & os.O_RDONLY != 0 {
		fs.Log("O_RDONLY");
	}	else if o.Flags & os.O_RDWR != 0 {
		fs.Log("O_RDWR");
	}
	if o.Flags & os.O_NONBLOCK != 0 {
		fs.Log("O_NONBLOCK");
	}
	if o.Flags & os.O_EXCL != 0 {
		fs.Log("O_EXCL");
	}
	// O_CREAT isn't a possibility in FuseOpen...
	if i,ok := fs.nodes[o.Nodeid]; ok {
		i.HandleFlags(o.Flags) // handle O_TRUNC
		if o.Flags & os.O_APPEND != 0 {
			fs.Log("O_APPEND");
			i.HandleFlags(o.Flags) // handle O_TRUNC
			fh := &OpenInode { i, -1 }
			return flags, fh, fuse.OK
		} else {
			return flags, &OpenInode { i, 0 }, fuse.OK
		}
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
	return &OpenInode{&i,0}
}

type Inode struct {
	Ino uint64
	Atime, Mtime, Ctime uint64
	Atimensec, Mtimensec, Ctimensec uint32
	Mode uint32
	Nlink uint32
	Uid, Gid uint32
	Nreading, Nwriting uint32
	FileContents []byte
	DirContents map[string]uint64
}

type OpenInode struct {
	Ino *Inode
	Location int
}

func (i *Inode) Log(v ...interface{}) {
	fmt.Println("inode", i.Ino, ":", v)
}

func (i *OpenInode) Log(v ...interface{}) {
	fmt.Println("inode", i.Ino.Ino, ":", v)
}

func (i *Inode) IsDir() bool {
	return (i.Mode & (fuse.S_IFDIR | fuse.S_IFREG)) == fuse.S_IFDIR
}

func (i *Inode) HandleFlags(f fuse.Flags) {
	// Here we do things like O_TRUNC
	if f & os.O_TRUNC != 0 {
		i.Log("O_TRUNC");
		i.FileContents = i.FileContents[0:0] // truncate file!
	}
}

func (i *OpenInode) FuseRead(io *fuse.Io) ([]byte, fuse.Error) {
	i.Log("FuseRead", io) // the following is wrong...
	i.Log("offset is", io.Offset)
	i.Log("size is", io.Size)
	if i.Ino.IsDir() {
		de := new(fuse.DEntryList)
		var num uint64 = 0
		for path,ino := range i.Ino.DirContents {
			i.Log("::: ",path,ino)
			if num >= io.Offset {
				de.AddString(path, ino, fuse.S_IFDIR)
				num += 1
			}
		}
		return de.Bytes(), fuse.OK
	}
	return i.Ino.FileContents[i.Location:], fuse.OK
}

func (oi *OpenInode) FuseWrite(io *fuse.Io, d []byte) (uint64, fuse.Error) {
	oi.Log("FuseWrite", io, string(d))
	oi.Log("io.Offset as int is ", int(io.Offset))
	oi.Log("io.Offset is ", io.Offset)
	oi.Log("io itself is ", io)
	start := oi.Location + int(io.Offset)
	if oi.Location < 0 {
		start = len(oi.Ino.FileContents)
	}
	stop := start + len(d)
	if cap(oi.Ino.FileContents) < stop {
		oi.Log("Expanding FileContents...")
		fc := make([]byte, stop, 2*stop)
		for i,v := range oi.Ino.FileContents {
			fc[i] = v
		}
		oi.Ino.FileContents = fc
	} else {
		oi.Ino.FileContents = oi.Ino.FileContents[0:stop]
	}
	for i,v := range d {
		oi.Ino.FileContents[start + i] = v
	}
	oi.Log("FileContents is now: ", string(oi.Ino.FileContents))
	if oi.Location >= 0 {
		oi.Location += len(d)
	}
	return uint64(len(d)), fuse.OK
}

func (i *OpenInode) FuseRelease(r *fuse.Release) fuse.Error {
	i.Log("FuseRelease", r)
	if i.Ino.Nreading > 0 { i.Ino.Nreading -= 1 }
	if i.Ino.Nwriting > 0 { i.Ino.Nwriting -= 1 }
	if !i.Ino.IsDir() {
		i.Log("File contents:", string(i.Ino.FileContents))
	}
	return fuse.OK
}

func (i *OpenInode) FuseFlush(f *fuse.Flush) fuse.Error {
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
