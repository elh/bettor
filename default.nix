{ pkgs ? import <nixpkgs> { } }:

with pkgs;

mkShell {
  buildInputs = [
    go_1_19
    protoc-gen-validate
    # CI dependencies
    golangci-lint
    golint
    buf
    # Development dependencies
    protoc-gen-go
    protoc-gen-connect-go
    protoc-gen-doc
    jq # for buf generate
  ];
}
