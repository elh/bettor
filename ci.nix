{ pkgs ? import <nixpkgs> { } }:

with pkgs;

mkShell {
  buildInputs = [
    go_1_26
    # CI dependencies
    golangci-lint
    golint
    buf
  ];
}
