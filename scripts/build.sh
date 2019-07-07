#!/usr/bin/env bash

package_name=standup-reporter
output_directory=build

platforms=("linux/amd64" "darwin/amd64" "windows/amd64")

for platform in "${platforms[@]}"; do
    platform_split=(${platform//\// })
    GOOS=${platform_split[0]}
    GOARCH=${platform_split[1]}
    output_name=${package_name}"-"${GOOS}"-"${GOARCH}

    if [[ ${GOOS} = "windows" ]]; then
        output_name+=".exe"
    fi

    echo "Building ${output_name}..."

    env GOOS=${GOOS} GOARCH=${GOARCH} go build -o ${output_directory}/${output_name}

    if [[ $? -ne 0 ]]; then
        echo -e "\nAn error has occurred! Aborting the script execution..."
        exit 1
    fi
done
