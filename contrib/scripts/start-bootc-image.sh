#!/bin/sh

set -exu

BOOTC_IMG_PATH=$1
qemu-img convert -f qcow2 -O raw ${BOOTC_IMG_PATH} bootc-overlay.img

#cp -c bootc-overlay.img snapshopt.img

./out/vfkit --cpus 2 --memory 2048 \
    --bootloader efi,variable-store=./efi-variable-store,create \
    --device virtio-blk,path=bootc-overlay.img \
    --device virtio-serial,stdio \
    --device virtio-net,nat,mac=72:20:43:d4:38:62 \
    --device virtio-rng
