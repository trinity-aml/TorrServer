#!/bin/bash

PLATFORMS=""
PLATFORMS="$PLATFORMS linux/amd64 linux/386"
PLATFORMS="$PLATFORMS windows/amd64 windows/386"
PLATFORMS="$PLATFORMS darwin/amd64 darwin/arm64"
PLATFORMS="$PLATFORMS freebsd/amd64"
PLATFORMS="$PLATFORMS linux/mips linux/mipsle linux/mips64 linux/mips64le"

type setopt >/dev/null 2>&1

GOBIN="go"

$GOBIN version

LDFLAGS="'-s -w'"
FAILURES=""
ROOT=${PWD}
OUTPUT="${ROOT}/dist/TorrServer"

rm -fr "${ROOT}/dist"
cd "${ROOT}/server"
#$GOBIN clean -i -r -cache
$GOBIN mod tidy

BUILD_FLAGS="-tags disable_libutp -ldflags=${LDFLAGS}"

################################################
### ARM build section                        ###
################################################

GOOS="linux"
GOARCH="arm64"
BIN_FILENAME="${OUTPUT}-${GOOS}-${GOARCH}"
CMD="GOOS=linux GOARCH=${GOARCH} ${GOBIN} build ${BUILD_FLAGS} -o ${BIN_FILENAME} ./cmd"
echo "${CMD}"
eval $CMD || FAILURES="${FAILURES} ${PLATFORM}"

GOARCH="arm"
GOARM="7"
BIN_FILENAME="${OUTPUT}-${GOOS}-${GOARCH}${GOARM}"
CMD="GOOS=${GOOS} GOARCH=${GOARCH} GOARM=${GOARM} ${GOBIN} build ${BUILD_FLAGS} -o ${BIN_FILENAME} ./cmd"
echo "${CMD}"
eval "${CMD}" || FAILURES="${FAILURES} ${GOOS}/${GOARCH}${GOARM}"

################################################
### X86, darwin, freebsd, mips build section ###
################################################

for PLATFORM in $PLATFORMS; do
  GOOS=${PLATFORM%/*}
  GOARCH=${PLATFORM#*/}
  BIN_FILENAME="${OUTPUT}-${GOOS}-${GOARCH}"
  if [[ "${GOOS}" == "windows" ]]; then BIN_FILENAME="${BIN_FILENAME}.exe"; fi
  if [[ "${GOOS}" == "linux" ]]; then
    CMD="GOOS=${GOOS} GOARCH=${GOARCH} ${GOBIN} build ${BUILD_FLAGS} -o ${BIN_FILENAME} ./cmd"
  else
    CMD="GOOS=${GOOS} GOARCH=${GOARCH} ${GOBIN} build ${BUILD_FLAGS} -o ${BIN_FILENAME} ./cmd"
  fi
  echo "${CMD}"
  eval $CMD || FAILURES="${FAILURES} ${PLATFORM}"
done

# eval errors
if [[ "${FAILURES}" != "" ]]; then
  echo ""
  echo "failed on: ${FAILURES}"
  exit 1
fi