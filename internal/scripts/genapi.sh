#!/bin/bash

if [ "$#" -ne 1 ]; then
	echo "usage: ./genapi.sh /path/to/neofs-api"
	exit 1
fi

SDK_PROTO_PATH=proto
API_PATH=$1

# collect and copy protobuf files where the generated ones will be placed
API_PROTO_FILES=$(find "${API_PATH}" -name '*.proto')
SDK_PROTO_FILES=()
for file in ${API_PROTO_FILES}; do
	SDK_PROTO_FILE=${SDK_PROTO_PATH}/${file#"${API_PATH}/"}
	mkdir -p "$(dirname "${SDK_PROTO_FILE}")"
	cp -r "$file" "${SDK_PROTO_FILE}"
	SDK_PROTO_FILES+=("${SDK_PROTO_FILE}")
done

# fix imports in copied files
for file in "${SDK_PROTO_FILES[@]}"; do
	sed -i "s/import\ \"\(.*\)\/\(.*\)\.proto\";/import\ \"${SDK_PROTO_PATH}\/\1\\/\2.proto\";/" "$file"
	sed -i "s/option go_package = \"\(.*\)\/\(.*\)\/\(.*\)\/\(.*\)\/\(.*\)\/grpc/option go_package = \"\1\\/\2\\/neofs-sdk-go\/${SDK_PROTO_PATH}\/\5/" "$file"
done

# compile protobuf
go list -f '{{.Path}}/...@{{.Version}}' -m  google.golang.org/protobuf | xargs go install -v
for file in "${SDK_PROTO_FILES[@]}"; do
	protoc \
		--proto_path=. \
		--go_out=. --go_opt=paths=source_relative \
		--go-grpc_opt=require_unimplemented_servers=false \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative "$file"; \
done

# remove no longer needed protobuf files
for file in "${SDK_PROTO_FILES[@]}"; do
	rm "$file"
done
