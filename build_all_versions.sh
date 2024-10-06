#!/bin/bash

# 设置变量
PROGRAM_NAME="dameng_exporter"
VERSION=$1
CONFIG_FILE="dameng_exporter.config custom_metrics.toml"

if [[ $1 == "" ]];then
  echo "need version args"
  exit 1
fi

# 编译 Linux 64 位版本
export GOOS=linux
export GOARCH=amd64
go build -ldflags "-s -w" -o ${PROGRAM_NAME}_${VERSION}_linux_amd64
if [ $? -ne 0 ]; then
    echo "Error compiling Linux 64-bit version"
    sleep 3
    exit 1
fi
echo "Compiled Linux 64-bit version successfully"

# 打包 Linux 版本为 tar.gz，包括配置文件
tar -czvf ${PROGRAM_NAME}_${VERSION}_linux_amd64.tar.gz ${PROGRAM_NAME}_${VERSION}_linux_amd64 $CONFIG_FILE
if [ $? -ne 0 ]; then
    echo "Error packaging Linux 64-bit version"
    sleep 3
    exit 1
fi

# 清理编译生成的可执行文件
rm -f ${PROGRAM_NAME}_${VERSION}_linux_amd64


# 编译 Linux ARM 版本
export GOOS=linux
export GOARCH=arm64
go build -ldflags "-s -w" -o ${PROGRAM_NAME}_${VERSION}_linux_arm64
if [ $? -ne 0 ]; then
    echo "Error compiling Linux ARM version"
    sleep 3
    exit 1
fi
echo "Compiled Linux ARM version successfully"

# 打包 Linux ARM 版本为 tar.gz，包括配置文件
tar -czvf ${PROGRAM_NAME}_${VERSION}_linux_arm64.tar.gz ${PROGRAM_NAME}_${VERSION}_linux_arm64 $CONFIG_FILE
if [ $? -ne 0 ]; then
    echo "Error packaging Linux ARM version"
    sleep 3
    exit 1
fi

# 清理编译生成的可执行文件
rm -f ${PROGRAM_NAME}_${VERSION}_linux_arm64


echo "All versions compiled successfully"
exit 0
