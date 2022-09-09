#!/usr/bin/env python3
import argparse
import os
import sys
import subprocess
import shutil
import tempfile
import json
import tarfile
from urllib import request

from typing import Optional, List

# Paths

home_path = os.path.expanduser("~")
home_local = os.path.join(home_path, ".local")
extism_path = os.getenv("EXTISM_PATH", os.path.join(home_path, ".extism"))
tmp_path = tempfile.gettempdir()

remote_ssh = "git@github.com:extism/extism"
remote_http = "https://github.com/extism/extism"

# Utils


def cp(src: str, dest: str, sudo: bool = False):
    if dest.startswith("/usr/"):
        sudo = True
    """Copy a file, optionally using sudo"""
    if sudo:
        subprocess.run(["sudo", "cp", src, dest])
    else:
        shutil.copy2(src, dest)


def quit(*args, code=1):
    """Exit with message"""
    print(*args)
    sys.exit(code)


class System:
    """System specific code"""

    def __init__(self):
        self.uname = os.uname()

        if self.uname.sysname == "Darwin":
            self.soext = "dylib"
        elif self.uname.sysname == "Windows":
            self.soext = "dll"
        else:
            self.soext = "so"

    def lib(self):
        """Get the extism library name with the proper extension"""
        return f"libextism.{self.soext}"

    def asset_prefix(self, libc: str = "gnu"):
        """Get the prefix of the uploaded release asset for the current system"""

        if self.uname.sysname == "Darwin":
            return f"libextism-{self.uname.machine}-apple-darwin"
        elif self.uname.sysname == "Linux":
            return f"libextism-{self.uname.machine}-unknown-linux-{libc}"
        quit("Invalid OS, try installing from source")


# Parse arguments

parser = argparse.ArgumentParser()
subparsers = parser.add_subparsers(title="command",
                                   dest="command",
                                   required=True)
parser.add_argument("--quiet",
                    default=False,
                    action="store_true",
                    help="Limit output to errors")
parser.add_argument("--prefix", default=None, help="Installation prefix")
parser.add_argument("--github-token",
                    default=os.getenv("GITHUB_TOKEN"),
                    help="Github token")
parser.add_argument("--sudo",
                    default=False,
                    action="store_true",
                    help="Use sudo to install files")

# Build arguments
build = subparsers.add_parser("build")
build.add_argument("--features",
                   default=["default"],
                   nargs="*",
                   help="Enable/disable features")
build.add_argument("--mode",
                   default="release",
                   choices=["debug", "release"],
                   help="Cargo build mode")
build.add_argument(
    "--no-default-features",
    default=False,
    action="store_true",
    help="Disable default features",
)

# Install command
install = subparsers.add_parser("install")
install.add_argument(
    "--no-update",
    default=False,
    action="store_true",
    help="Update if an existing installation is present",
)
install.add_argument("version",
                     nargs="?",
                     default="latest",
                     help="Version to install")
install.add_argument(
    "--list-available",
    default=False,
    action="store_true",
    help="List available versions",
)
install.add_argument("--branch", default="main", help="Git branch or tag")

# Fetch command
fetch = subparsers.add_parser("fetch")
fetch.add_argument("version",
                   nargs="?",
                   default="latest",
                   help="Version to install")
fetch.add_argument("--branch", default="main", help="Git branch or tag")
fetch.add_argument("--libc", default="gnu", help="Linux libc")

# Uninstall command
uninstall = subparsers.add_parser("uninstall")

# Link command
link = subparsers.add_parser("link")
link.add_argument("version",
                  nargs="?",
                  default="latest",
                  help="Version to install")

link.add_argument("--mode",
                  default="release",
                  choices=["debug", "release"],
                  help="Cargo build mode")

# Info command
info = subparsers.add_parser("info")

info.add_argument("--cflags",
                  default=False,
                  action="store_true",
                  help="Print include path")
info.add_argument("--libs",
                  default=False,
                  action="store_true",
                  help="Print link flags")

# Call subcommand
call = subparsers.add_parser("call")
call.add_argument("wasm", help='WASM file')
call.add_argument("--input", default=None, help='Plugin input')
call.add_argument("function", help='Function name')
call.add_argument("--log-level", default="error", help="Set log level")
call.add_argument("--wasi", action='store_true', help="Enables WASI")
call.add_argument("--set-config", default=None, help="Set config JSON for plug-in")


class ExtismBuilder:
    """Builds and installs extism from source or Github releases"""

    def __init__(self, prefix: Optional[str] = None, source=None, quiet=False):
        self.load_config()

        if prefix is not None or not hasattr(self, "install_prefix"):
            self.install_prefix = prefix or home_local

        if source is not None or not hasattr(self, "source_path"):
            self.source_path = source or os.path.join(extism_path, "extism")

        self._init()

        self.quiet = quiet
        self.system = System()

    def _init(self):
        if hasattr(self, "source_path"):
            self.runtime_path = os.path.join(self.source_path, "runtime")
            self.workspace_path = self.source_path

    def save_config(self, version: Optional[str] = None):
        """Save config to disk"""
        os.makedirs(extism_path, exist_ok=True)
        with open(os.path.join(extism_path, "config.json"), "w") as f:
            j = {
                "source_path": self.source_path,
                "install_prefix": self.install_prefix,
            }
            if version is not None:
                j["version"] = version
            f.write(json.dumps(j, indent=True))

    def load_config(self):
        """Load config from disk"""
        path = os.path.join(extism_path, "config.json")
        if not os.path.exists(path):
            return

        with open(path) as f:
            j = json.loads(f.read())
            for k, v in j.items():
                self.__setattr__(k, v)
        self._init()

    def releases(self, token: Optional[str] = None):
        """Get a list of all releases"""
        req = request.Request(
            url="https://api.github.com/repos/extism/extism/releases",
            method="GET",
        )
        req.add_header("Accept", "application/vnd.github+json")
        if token is not None:
            req.add_header("Authorization", f"token {token}")
        res = request.urlopen(req)
        data = json.loads(res.read())
        dest = []
        for release in data:
            dest.append(
                dict(
                    filter(
                        lambda item: item[0] in
                        ["tarball_url", "assets", "tag_name", "name"],
                        release.items(),
                    )))
        return dest

    def find_release(self, name: str, token: Optional[str] = None):
        """Find a specific release"""
        if name == "git":
            return None

        releases = self.releases(token=token)

        if name == "latest":
            return releases[0]

        for release in releases:
            if release["tag_name"] == name:
                return release

        quit(f"Invalid release {name}")

    def download_git(self, branch: str = "main", no_update: bool = True):
        """
        Download or update the git repo, if `no_update` is set then this will only clone
        the repo if it doesn't already exist
        """

        if os.path.exists(os.path.join(self.source_path, ".git")):
            if no_update:
                return
            subprocess.run(["git", "fetch", "origin"], cwd=self.source_path)

            subprocess.run(["git", "checkout", branch], cwd=self.source_path)

            subprocess.run(["git", "pull", "origin", branch],
                           cwd=self.source_path)
        else:
            os.makedirs(os.path.dirname(self.source_path), exist_ok=True)
            subprocess.run([
                "git", "clone", "--branch", branch, remote_http,
                self.source_path
            ])

    def download_release(self, release: dict, libc: str = "gnu"):
        """
        Download a release from Github
        """
        cache = os.path.join(extism_path, "cache")
        cache_path = os.path.join(cache, release["tag_name"])
        if os.path.exists(cache_path):
            return

        os.makedirs(cache, exist_ok=True)

        asset_prefix = self.system.asset_prefix(libc=libc)
        url = None
        found = []
        for asset in release["assets"]:
            if '.txt' not in asset['name']:
                found.append(asset["name"])
            if asset["name"].startswith(asset_prefix):
                url = asset["browser_download_url"]
        if url is None:
            found = ', '.join(found)
            quit(f"Unable to find suitable release: found {found}")
        else:
            req = request.Request(url=url)
        res = request.urlopen(req)
        with open(cache_path, "wb") as f:
            f.write(res.read())
            f.flush()

    def link_release(self, release: dict, sudo: bool = False):
        """Copy lib and header file from a release to the install prefix"""
        cache = os.path.join(extism_path, "cache")
        release_ = os.path.join(extism_path, "release")
        cache_path = os.path.join(cache, release["tag_name"])
        release_path = os.path.join(release_, release["tag_name"])
        tar = tarfile.open(name=cache_path)

        os.makedirs(release_path, exist_ok=True)

        lib_dest = os.path.join(self.install_prefix, "lib", self.system.lib())
        header_dest = os.path.join(self.install_prefix, "include", "extism.h")

        tar.extractall(path=release_path)
        cp(os.path.join(release_path, self.system.lib()), lib_dest, sudo)

        self.print("Installed", lib_dest)

        cp(os.path.join(release_path, "extism.h"), header_dest, sudo)
        self.print("Installed", header_dest)

    def link_git(self, mode: str = "release", sudo: bool = False):
        """Copy lib and header file from git repo to the install prefix"""
        lib_name = self.system.lib()
        lib_dest = os.path.join(self.install_prefix, "lib", lib_name)
        header_dest = os.path.join(self.install_prefix, "include", "extism.h")
        lib_src = os.path.join(self.workspace_path, "target", mode, lib_name)
        header_src = os.path.join(self.runtime_path, "extism.h")
        cp(lib_src, lib_dest, sudo)
        self.print(f"Installed {lib_dest}")

        cp(header_src, header_dest, sudo)
        self.print(f"Installed {header_dest}")

    def link(self,
             release: Optional[dict] = None,
             sudo: bool = False,
             mode: str = "release"):
        if release is None:
            version = "git"
        else:
            version = release["tag_name"]
        self.print(f"Installing to {self.install_prefix} (version {version})")
        try:
            os.makedirs(os.path.join(self.install_prefix, "lib"),
                        exist_ok=True)
            os.makedirs(os.path.join(self.install_prefix, "include"),
                        exist_ok=True)
        except:
            pass
        if release is None:
            self.link_git(sudo=sudo, mode=mode)
            self.save_config(version="git")
        else:
            self.link_release(release)
            self.save_config(version=release["tag_name"])

    def unlink(self, sudo: bool = False):
        lib = os.path.join(self.install_prefix, "lib", self.system.lib())
        header = os.path.join(self.install_prefix, "include", "extism.h")
        try:
            if sudo:
                subprocess.run(["sudo", "rm", "-f", lib])
            else:
                os.remove(lib)
            self.print(f"Removed {lib}")
        except:
            self.print(f"Warning: file does not exist {lib}")

        try:
            if sudo:
                subprocess.run(["sudo", "rm", "-f", header])
            else:
                os.remove(header)
            self.print(f"Removed {header}")
        except:
            self.print(f"Warning: file does not exist {header}")

    def fetch(
        self,
        version: str = "git",
        branch: str = "main",
        libc: str = "gnu",
        token: Optional[str] = None,
        no_update: bool = False,
    ):
        if version == "git":
            self.print("Getting source from git")
            self.download_git(branch=branch, no_update=no_update)
            return None

        self.print(f"Getting release for {version}")
        release = self.find_release(version, token=token)
        self.download_release(release=release, libc=libc)
        return release

    def build(
        self,
        mode: Optional[str] = None,
        features: List[str] = ["default"],
        no_default_features: bool = False,
    ):
        self.print(f"Building from source in {self.runtime_path}")
        cmd = ["cargo", "build"]
        if mode is None:
            mode = "release"

        if mode == "release":
            cmd.append(f"--{mode}")

        if no_default_features:
            cmd.append("--no-default-features")
            if features == ["default"]:
                features = []
        cmd.extend(["--features", ",".join(features)])
        subprocess.run(cmd, cwd=self.runtime_path)

    def install(
        self,
        version: str = "git",
        branch: str = "main",
        sudo: bool = False,
        token: Optional[str] = None,
        no_update: bool = False,
    ):
        release = self.fetch(version=version,
                             branch=branch,
                             token=token,
                             no_update=no_update)
        if version == "git":
            self.build()
        self.link(release=release, sudo=sudo)

    def print(self, *args):
        if not self.quiet:
            print(*args)


def main():
    args = parser.parse_args()
    extism = ExtismBuilder(prefix=args.prefix, quiet=args.quiet)
    if args.command == "install":
        if args.list_available:
            print("git")
            first = True
            for release in extism.releases(token=args.github_token):
                if first:
                    print(release['tag_name'], "(latest)")
                    first = False
                else:
                    print(release["tag_name"])
        else:
            extism.install(
                version=args.version,
                branch=args.branch,
                sudo=args.sudo,
                token=args.github_token,
                no_update=args.no_update,
            )
    elif args.command == "build":
        extism.build(
            features=args.features,
            mode=args.mode,
            no_default_features=args.no_default_features,
        )
    elif args.command == "uninstall":
        extism.unlink(sudo=args.sudo)
    elif args.command == "fetch":
        extism.fetch(
            version=args.version,
            branch=args.branch,
            libc=args.libc,
            token=args.github_token,
        )
    elif args.command == "link":
        release = extism.find_release(
            name=args.version,
            token=args.github_token,
        )
        extism.link(release, sudo=args.sudo, mode=args.mode)
    elif args.command == "info":
        if args.cflags:
            h = os.path.join(extism.install_prefix, "include")
            print(f"-I{h}", end=" ")

        if args.libs:
            l = os.path.join(extism.install_prefix, "lib")
            print(f"-L{l} -lextism")

        if args.cflags and not args.libs:
            print()

        if not args.cflags and not args.libs:
            if hasattr(extism, "version"):
                print(
                    f"Prefix\t{extism.install_prefix}\nVersion\t{extism.version}"
                )
    elif args.command == "call":
        if args.input is None and not sys.stdin.isatty():
            input = sys.stdin.read()
        else:
            input = args.input.encode()

        if not os.path.exists(extism.source_path):
            extism.fetch(version="git")

        if args.set_config is None:
            config = ""
        else:
            config = args.set_config.encode()

        config = json.loads(config)

        try:
            # First try installed Python library
            import extism
        except:
            # If that fails use the Python SDK from the source directory
            sys.path.append(os.path.join(extism.source_path, "python"))
            import extism

        data = open(args.wasm, 'rb').read()
        extism.set_log_file("-", args.log_level)
        plugin = extism.Plugin(data, wasi=args.wasi, config=config)
        r = plugin.call(args.function, input)
        sys.stdout.buffer.write(r)


if __name__ == "__main__":
    main()
