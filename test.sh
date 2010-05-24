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
echo hello > bar
grep hello bar
cd ..

ls -lh testing-mountpoint
killall -SIGINT myfs

# give it a moment to unmount the FS...
sleep 1

# check that the contents are no longer there.
ls testing-mountpoint/bar && exit 1
grep hello testing-mountpoint/bar && exit 1

echo $0 passed!