# Chat service

Chat service for vreco that leverages GRPC and buf.


## Install Linux

Install buf
```bash
GO111MODULE=on go install \
  github.com/bufbuild/buf/cmd/buf@v1.4.0 && \
sudo apt-get install jq git libnss3-tools && \
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2 && \
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28 && \
go install github.com/twitchtv/twirp/protoc-gen-twirp@v8.1.2
```

Install mkcert
```bash
curl -JLO "https://dl.filippo.io/mkcert/latest?for=linux/amd64" && \
chmod +x mkcert-v*-linux-amd64 && \
sudo cp mkcert-v*-linux-amd64 /usr/local/bin/mkcert
```