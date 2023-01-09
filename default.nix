{ pkgs ? import <nixpkgs> { } }:

with pkgs;

mkShell {
  buildInputs = [
    # Go
    go_1_19
    golangci-lint
    golint
    # Protocol buffers
    buf
    protoc-gen-go
    protoc-gen-go-grpc
    jq
  ];
}
