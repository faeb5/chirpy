{
  lib,
  buildGoModule,
  fetchFromGitHub,
}:
buildGoModule rec {
  pname = "bootdev";
  version = "1.10.1";

  src = fetchFromGitHub {
    owner = "bootdotdev";
    repo = "bootdev";
    rev = "v${version}";
    hash = "sha256-mgicbuN/CrqqEcORMmFLHYtxYPxHvLVt9UnP1ksDdkE=";
  };

  vendorHash = "sha256-jhRoPXgfntDauInD+F7koCaJlX4XDj+jQSe/uEEYIMM=";

  ldflags = ["-s"];

  CGO_ENABLED = 0;

  meta = {
    description = ''
      The official command line tool for Boot.dev.
      It allows you to submit lessons and do other such nonsense.
    '';
    homepage = "https://github.com/bootdotdev/bootdev";
    license = lib.licenses.mit;
  };
}
