{
  description = "A basic golang flake";

  inputs.nixpkgs.url = "github:NixOS/nixpkgs/nixos-unstable";
  inputs.flake-utils.url = "github:numtide/flake-utils";

  outputs = { self, nixpkgs, flake-utils }:
    (flake-utils.lib.eachDefaultSystem
      (system:
        let
          pkgs = import nixpkgs {
            inherit system;
          };

          pname = "changeme";
          buildInputs = with pkgs; [ ];
        in
        {
          packages = {
            default = pkgs.buildGoModule  {
              name = pname;
              src = ./.;
              inherit buildInputs;
              vendorHash = null;
            };
          };
          devShells. default = pkgs.mkShell {
            name = pname + "-shell";
            nativeBuildInputs = with pkgs; [
              go
              gopls
            ];
          };
        })
    );
}

