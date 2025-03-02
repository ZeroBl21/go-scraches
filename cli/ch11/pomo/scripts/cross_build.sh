#!/bin/bash

OS_LIST="linux windows darwin"
ARCH_LIST="amd64 arm arm64"

for os in ${OS_LIST}; do
    for arch in ${ARCH_LIST}; do 
        if [[ "$os/$arch" =~ ^(windows/arm64|darwin/arm)$ ]]; then
            continue
        fi

        echo "Building binary for $os $arch"
        mkdir -p releases/${os}/${arch}
        
        # output_file="releases/${os}/${arch}/my_program"

        # [[ "$os" == "windows" ]] && output_file+=".exe"

        # Compilar
        CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -o releases/${os}/${arch}/

        if [[ $? -ne 0 ]]; then
            echo "Error: Compilation failed for $os/$arch"
            exit 1
        fi
    done
done

echo "Compilation finished."
