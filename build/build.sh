#!/usr/bin/env bash

PACKAGE_NAME=standup-reporter
OUTPUT_DIR=bin
APP=github.com/jeremy-miller/standup-reporter/cmd

PLATFORMS=("linux/amd64/linux/x86_64" "darwin/amd64/macOS/x86_64" "windows/amd64/windows/x86_64")

for PLATFORM in "${PLATFORMS[@]}"; do
    PLATFORM_SPLIT=(${PLATFORM//\// })
    GOOS=${PLATFORM_SPLIT[0]}
    GOARCH=${PLATFORM_SPLIT[1]}
    OUTPUT_OS=${PLATFORM_SPLIT[2]}
    OUTPUT_ARCH=${PLATFORM_SPLIT[3]}
    OUTPUT_NAME=${PACKAGE_NAME}"-"${OUTPUT_OS}"-"${OUTPUT_ARCH}

    if [[ ${GOOS} = "windows" ]]; then
        OUTPUT_NAME+=".exe"
    fi

    echo "Building ${OUTPUT_NAME}"

    env GOOS=${GOOS} GOARCH=${GOARCH} go build -o ${OUTPUT_DIR}/${OUTPUT_NAME} ${APP}

    if [[ $? -ne 0 ]]; then
        echo -e "\nError building ${OUTPUT_NAME}"
        exit 1
    fi
done
