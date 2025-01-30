self: super: {
  pkgsStatic = super.pkgsStatic // {
    nodejs_22 = super.pkgsStatic.nodejs_22.overrideAttrs (oldAttrs:
      let

        darwin-cctools-only-libtool-fixed =
          super.runCommand "darwin-cctools-only-libtool-fixed" { cctools = super.lib.getBin super.pkgsStatic.buildPackages.cctools; } ''
            mkdir -p "$out/bin"
            ln -s "$cctools/bin/${super.stdenv.hostPlatform.config}-libtool" "$out/bin/libtool"
          '';

        isDarwinCctoolsLibtool = derivation:
          super.lib.isDerivation derivation && derivation.name == "darwin-cctools-only-libtool";

      in
      {
        doCheck = false;

        nativeBuildInputs = (builtins.filter (x: ! isDarwinCctoolsLibtool x) oldAttrs.nativeBuildInputs)
          ++ super.lib.optionals super.stdenv.hostPlatform.isDarwin [ darwin-cctools-only-libtool-fixed ];
      });
  };
}

