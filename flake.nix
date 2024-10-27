{
  description = "Boot.dev flake";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-24.05";
  };

  outputs = {
    self,
    nixpkgs,
  }: let
    pkgs = nixpkgs.legacyPackages.x86_64-linux;
  in {
    formatter.x86_64-linux = pkgs.alejandra;
    packages.x86_64-linux.bootdev = with pkgs; callPackage ./bootdev.nix {};
    packages.x86_64-linux.default = self.packages.x86_64-linux.bootdev;
    devShells.x86_64-linux.default = pkgs.mkShellNoCC {
      nativeBuildInputs = with pkgs; [go];
      packages = with pkgs; [sqlc goose postgresql jq gh gopls self.packages.x86_64-linux.bootdev];
      shellHook = ''echo "Welcome to the Boot.dev environment"'';
    };
  };
}
