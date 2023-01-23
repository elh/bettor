{ pkgs ? import <nixpkgs> { } }:

with pkgs;

mkShell {
  buildInputs = [
    go_1_19
    # CI dependencies
    golangci-lint
    golint
    buf
    # Development dependencies
    protoc-gen-go
    protoc-gen-connect-go
    protoc-gen-validate
    protoc-gen-doc
    jq # for buf generate
    # Deployment dependencies
    # flyctl # NOTE: maybe this is too active to be running off of nix registry
  ];
}
