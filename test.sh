#!/bin/sh

set -ev

make

fusermount -u testing-mountpoint || true
rm -rf testing-mountpoint
mkdir testing-mountpoint

grep hello testing-mountpoint/bar && exit 1

./myfs testing-mountpoint > /dev/null 2> /dev/null || true &

# sleep to let the file system be mounted
sleep 1

mount
grep hello testing-mountpoint/bar && exit 1

cd testing-mountpoint
mkdir boo
ls -l boo
echo foo > boo/foo
rmdir boo 2> err && exit 1
cat err
#grep 'Directory not empty' err
rm err
rm boo/foo
rmdir boo
ls -l boo && exit 1
echo hello > bar
cat bar
grep hello bar
# the size is 6
ls -l bar | grep ' 6 '
rm bar
ls -l bar && exit 1
cd ..

ls -lh testing-mountpoint
killall -SIGINT myfs

# give it a moment to unmount the FS...
sleep 1

# check that the contents are no longer there.
ls testing-mountpoint/bar && exit 1
grep hello testing-mountpoint/bar && exit 1

echo $0 passed!
