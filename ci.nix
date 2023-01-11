{ pkgs ? import <nixpkgs> { } }:

with pkgs;

mkShell {
  buildInputs = [
    go_1_19
    # CI dependencies
    golangci-lint
    golint
    buf
  ];
}
