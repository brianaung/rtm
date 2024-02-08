{
	description = "rtm build environment";
	inputs = {
		templ.url = "github:a-h/templ";
		nixpkgs.url = "github:nixos/nixpkgs/nixos-unstable";
	};
	outputs = inputs@{ self, nixpkgs, ... }:

	let
		system = "x86_64-linux";
		templ = inputs.templ.packages.${system}.templ;
		pkgs = nixpkgs.legacyPackages.${system}.pkgs;
	in {
		# packages = with pkgs; {
		# 	myNewPackage = pkgs.buildGoModule rec {
		# 		pname = "myNewPackage";
		# 		src = ./.;
		# 		goPackagePath = "example.com/${pname}";
		# 		checkInputs = [ pkgs.go ];
		# 		meta = with pkgs.stdenv.lib; {
		# 			description = "My new package";
		# 			license = licenses.mit;
		# 		};
		# 		preBuild = ''
		# 			${templ}/bin/templ generate
		# 		'';
		# 	};
		# };

		devShells.${system}.default = pkgs.mkShell {
			buildInputs = with pkgs; [
				nodejs_18
				go
				goose
				templ
			];
		};
	};
}
