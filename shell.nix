{ pkgs ? import <nixpkgs> {} }:

pkgs.mkShell {
  packages = [
    pkgs.go
    pkgs.sqlite
  ];

  shellHook = ''
    export PATH=$PWD/result/bin:$PATH
  '';
}
