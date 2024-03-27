#!/bin/sh

set -exu

YQ=${YQ:-yq}
BUNDLE_PATH=$1
KERNEL=$(cat ${BUNDLE_PATH}/crc-bundle-info.json | ${YQ} .nodes[0].kernel)
INITRD=$(cat ${BUNDLE_PATH}/crc-bundle-info.json | ${YQ} .nodes[0].initramfs)
CMDLINE=$(cat ${BUNDLE_PATH}/crc-bundle-info.json | ${YQ} .nodes[0].kernelCmdLine)
DISKIMG=$(cat ${BUNDLE_PATH}/crc-bundle-info.json | ${YQ} .storage.diskImages[0].name)
cp -c ${BUNDLE_PATH}/${DISKIMG} overlay.img

SERIAL_CONSOLE_ARGS="--device virtio-serial,pty"
if [[ -n "${VFKIT_USE_STDIO-}" ]]; then
  SERIAL_CONSOLE_ARGS="--device virtio-serial,stdio"
fi

if [[ -n "${VFKIT_USE_GUI-}" ]]; then
  GUI_ARGS="--gui --device virtio-input,keyboard --device virtio-input,pointing --device virtio-gpu,width=1920,height=1080"
fi

./bin/vfkit --cpus 2 --memory 2048 \
    --kernel "${BUNDLE_PATH}/${KERNEL}" \
    --initrd "${BUNDLE_PATH}/${INITRD}" \
    --kernel-cmdline "${CMDLINE}" \
    --device virtio-blk,path=overlay.img \
    --device virtio-net,nat,mac=72:20:43:d4:38:62 \
    --device virtio-rng \
    --restful-uri unix:///Users/teuf/dev/vfkit/rest.sock \
    ${SERIAL_CONSOLE_ARGS} \
    ${GUI_ARGS-}
