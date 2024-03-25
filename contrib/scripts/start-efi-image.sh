#!/bin/sh

set -exu

DISKIMG=$1
cp -c ${DISKIMG} overlay.img

if [[ -n "${VFKIT_USE_GUI-}" ]]; then
  GUI_ARGS="--gui --device virtio-input,keyboard --device virtio-input,pointing --device virtio-gpu,width=1920,height=1080"
fi

SERIAL_CONSOLE_ARGS="--device virtio-serial,pty"
if [[ -n "${VFKIT_USE_STDIO-}" ]]; then
  SERIAL_CONSOLE_ARGS="--device virtio-serial,stdio"
fi

./bin/vfkit --cpus 2 --memory 2048 \
    --bootloader efi,variable-store=efi-variable-store,create \
    --device virtio-blk,path=overlay.img \
    --device virtio-net,nat,mac=72:20:43:d4:38:62 \
    --device virtio-rng \
    --restful-uri unix:///Users/teuf/dev/vfkit/rest.sock \
    ${SERIAL_CONSOLE_ARGS} \
    ${GUI_ARGS-}
