# This file is auto-updated by GoReleaser on each release.
# Do not edit manually — changes will be overwritten.
{ lib, stdenvNoCC, fetchurl }:
let
  version = "0.0.1";
  tarballs = {
    "aarch64-darwin" = fetchurl {
      url = "https://github.com/flaticols/gitmd/releases/download/v${version}/gitmd_Darwin_arm64.tar.gz";
      hash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="; # arm — updated by GoReleaser
    };
    "x86_64-darwin" = fetchurl {
      url = "https://github.com/flaticols/gitmd/releases/download/v${version}/gitmd_Darwin_x86_64.tar.gz";
      hash = "sha256-AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="; # x86 — updated by GoReleaser
    };
  };
in
stdenvNoCC.mkDerivation {
  pname = "gitmd";
  inherit version;

  src = tarballs.${stdenvNoCC.hostPlatform.system}
    or (throw "Unsupported system: ${stdenvNoCC.hostPlatform.system}. gitmd ships macOS binaries only.");

  unpackPhase = "tar xzf $src";

  installPhase = ''
    runHook preInstall
    install -Dm755 gitmd $out/bin/gitmd
    runHook postInstall
  '';

  meta = {
    description = "Manage git commit trailers (Reason, Ticket, Assisted, …) for git 2.54+";
    homepage = "https://github.com/flaticols/gitmd";
    license = lib.licenses.mit;
    mainProgram = "gitmd";
    platforms = [ "aarch64-darwin" "x86_64-darwin" ];
  };
}
