#!/bin/bash

#Build linux-version first and determine the build-version
GOOS=linux GOARCH=amd64 go build -ldflags '-s -w -extldflags "-static"' .
BUILD_VERSION=$(space-cloud -v | cut -f3 -d ' ')

mkdir linux && mkdir windows && mkdir darwin
zip space-cloud_v$BUILD_VERSION.zip space-cloud
mv ./space-cloud_v$BUILD_VERSION.zip ./linux/
rm space-cloud
GOOS=darwin GOARCH=amd64 go build -ldflags '-s -w -extldflags "-static"' .
zip space-cloud_v$BUILD_VERSION.zip space-cloud
mv ./space-cloud_v$BUILD_VERSION.zip ./darwin/
rm space-cloud
GOOS=windows GOARCH=amd64 go build -ldflags '-s -w -extldflags "-static"' .
zip space-cloud_v$BUILD_VERSION.zip space-cloud.exe
mv ./space-cloud_v$BUILD_VERSION.zip ./windows/
rm space-cloud.exe

# echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin
# docker push spaceuptech/space-cloud:latest

#Upload as versioned build
curl -H "Authorization: Bearer $JWT_TOKEN" -F "file=@./darwin/space-cloud_v$BUILD_VERSION.zip" -F 'fileType=file' -F 'makeAll=false' -F 'path=/darwin' https://spaceuptech.com/v1/api/downloads/files
curl -H "Authorization: Bearer $JWT_TOKEN" -F "file=@./windows/space-cloud_v$BUILD_VERSION.zip" -F 'fileType=file' -F 'makeAll=false' -F 'path=/windows' https://spaceuptech.com/v1/api/downloads/files
curl -H "Authorization: Bearer $JWT_TOKEN" -F "file=@./linux/space-cloud_v$BUILD_VERSION.zip" -F 'fileType=file' -F 'makeAll=false' -F 'path=/linux' https://spaceuptech.com/v1/api/downloads/files

#Upload as latest build
curl -H "Authorization: Bearer $JWT_TOKEN" -F 'file=@./darwin/space-cloud.zip' -F 'fileType=file' -F 'makeAll=false' -F 'path=/darwin' https://spaceuptech.com/v1/api/downloads/files
curl -H "Authorization: Bearer $JWT_TOKEN" -F 'file=@./windows/space-cloud.zip' -F 'fileType=file' -F 'makeAll=false' -F 'path=/windows' https://spaceuptech.com/v1/api/downloads/files
curl -H "Authorization: Bearer $JWT_TOKEN" -F 'file=@./linux/space-cloud.zip' -F 'fileType=file' -F 'makeAll=false' -F 'path=/linux' https://spaceuptech.com/v1/api/downloads/files
