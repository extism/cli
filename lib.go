package cli

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/ebitengine/purego"
	"github.com/google/go-github/v55/github"
	"github.com/spf13/cobra"
)

type libArgs struct {
	args       []string
	prefix     string
	libDir     string
	includeDir string
}

type libInstallArgs struct {
	libArgs
	version string
	os      string
	arch    string
}

type libUninstallArgs struct {
	libArgs
}

func (a *libArgs) SetArgs(args []string) {
	a.args = args
}

func getReleases(ctx context.Context) (releases []*github.RepositoryRelease, err error) {
	Log("Fetching releases from Github")
	client := github.NewClient(nil)
	if GithubToken != "" {
		client = client.WithAuthToken(GithubToken)
	}
	releases, _, err = client.Repositories.ListReleases(ctx, "extism", "extism", nil)
	if err != nil {
		return releases, err
	}
	Log("Found", len(releases), "releases")
	sort.Slice(releases, func(i, j int) bool {
		return releases[i].CreatedAt.After(releases[j].CreatedAt.Time)
	})
	return releases, nil
}

func findRelease(ctx context.Context, tag string) (release *github.RepositoryRelease, err error) {
	releases, err := getReleases(ctx)
	if err != nil {
		return release, err
	}

	if tag == "" {
		Log("Getting most recent release")
	} else {
		Log("Searching for releases tagged with version:", tag)
	}

	if tag == "" {
		rel := releases[0]
		if rel.GetTagName() == "latest" {
			rel = releases[1]
		}
		Log("Found", rel.URL, "published at", rel.PublishedAt)
		return rel, nil
	}

	for _, rel := range releases {
		if strings.HasPrefix(rel.GetTagName(), tag) {
			Log("Found", rel.URL, "published at", rel.PublishedAt)
			return rel, nil
		}
	}

	return nil, errors.New("unable to find release " + tag)
}

func assetPrefix(os, arch string) (string, error) {
	s := "libextism-"
	if arch == "amd64" {
		s += "x86_64"
	} else if arch == "arm64" {
		s += "aarch64"
	} else {
		s += arch
	}
	if os == "linux" {
		return s + "-unknown-linux-gnu", nil
	} else if os == "windows" {
		return s + "-pc-windows-msvc", nil
	} else if os == "windows-gnu" {
		return s + "-pc-windows-gnu", nil
	} else if os == "darwin" || os == "macos" {
		return s + "-apple-darwin", nil
	}

	return "", errors.New("unsupported " + os + " " + arch)
}

func sharedLibraryName(os string) string {
	switch os {
	case "darwin":
		fallthrough
	case "macos":
		return "libextism.dylib"
	case "windows":
		fallthrough
	case "windows-gnu":
		return "extism.dll"
	default:
		return "libextism.so"
	}
}

func runLibInstall(cmd *cobra.Command, installArgs *libInstallArgs) error {
	Log("Searching for release matching", installArgs.version)
	if installArgs.version == "git" {
		Log("Converting version from `git` to `latest` ")
		installArgs.version = "latest"
	}

	rel, err := findRelease(cmd.Context(), installArgs.version)
	if err != nil {
		return err
	}

	assetName, err := assetPrefix(installArgs.os, installArgs.arch)
	if err != nil {
		return err
	}

	Log("Searching for asset matching:", assetName)
	for _, asset := range rel.Assets {
		if strings.HasPrefix(asset.GetName(), assetName) && strings.HasSuffix(asset.GetName(), ".tar.gz") {
			Print("Installing", rel.GetTagName(), "to", installArgs.prefix)
			url := asset.GetBrowserDownloadURL()
			Print("Fetching", url)
			res, err := http.Get(url)
			if err != nil {
				return err
			}
			defer res.Body.Close()

			Log("Creating gzip reader")
			r, err := gzip.NewReader(res.Body)
			if err != nil {
				return err
			}

			Log("Reading tar file")
			tarReader := tar.NewReader(r)

			for {
				item, err := tarReader.Next()
				if err == io.EOF {
					break
				}

				if strings.HasSuffix(item.Name, getSharedObjectExt(installArgs.os)) {
					Log("Found shared object file in tarball")
					lib := filepath.Join(installArgs.prefix, installArgs.libDir)
					Log("Creating directory for lib:", lib)
					os.MkdirAll(lib, 0o755)
					out, err := os.Create(filepath.Join(lib, item.Name))
					if err != nil {
						return err
					}

					Print("Copying", item.Name, "to", out.Name())
					io.Copy(out, tarReader)
					out.Close()
				} else if strings.HasSuffix(item.Name, ".h") {
					Log("Found header file in tarball")
					include := filepath.Join(installArgs.prefix, installArgs.includeDir)
					Log("Creating directory for header file:", include)
					os.MkdirAll(include, 0o755)
					out, err := os.Create(filepath.Join(include, item.Name))
					if err != nil {
						return err
					}

					Print("Copying", item.Name, "to", out.Name())
					io.Copy(out, tarReader)
					out.Close()
				} else if strings.HasSuffix(item.Name, getStaticLibFileName(installArgs.os)) {
					Log("Found static library in tarball")
					lib := filepath.Join(installArgs.prefix, installArgs.libDir)
					Log("Creating directory for lib:", lib)
					os.MkdirAll(lib, 0o755)
					out, err := os.Create(filepath.Join(lib, item.Name))
					if err != nil {
						return err
					}

					Print("Copying", item.Name, "to", out.Name())
					io.Copy(out, tarReader)
					out.Close()

				} else {
					Log("File:", item.Name)
				}
			}
			return nil
		} else {
			Log("Invalid asset:", asset.GetName())
		}

	}

	return errors.New("No release asset found matching " + assetName)
}

func runLibUninstall(cmd *cobra.Command, uninstallArgs *libUninstallArgs) error {
	Log("Uninstalling files from prefix:", uninstallArgs.prefix)
	soFile := filepath.Join(uninstallArgs.prefix, uninstallArgs.libDir, getSharedObjectFileName(runtime.GOOS))
	Print("Removing", soFile)
	err := os.Remove(soFile)
	if err != nil {
		Print(err)
	}

	staticLibFile := filepath.Join(uninstallArgs.prefix, uninstallArgs.libDir, getStaticLibFileName(runtime.GOOS))
	Print("Removing", staticLibFile)
	err = os.Remove(staticLibFile)
	if err != nil {
		Print(err)
	}

	headerFile := filepath.Join(uninstallArgs.prefix, uninstallArgs.includeDir, "extism.h")
	Print("Removing", headerFile)
	err = os.Remove(headerFile)
	if err != nil {
		Print(err)
	}

	return nil
}

func runLibVersions(cmd *cobra.Command, args []string) error {
	releases, err := getReleases(cmd.Context())
	if err != nil {
		return err
	}

	Log("Found", len(releases))

	for _, rel := range releases {
		name := rel.GetTagName()
		if name == "latest" {
			continue
		}

		Print(name)
	}

	return nil
}

func runLibCheck(cmd *cobra.Command, args []string) error {
	soName := sharedLibraryName(runtime.GOOS)
	Log("dlopen", soName)
	ptr, err := dlopen(soName)
	if err != nil {
		return errors.New("unable to open libextism, no installation detected")
	}

	Log("Registering extism_version func")
	var version func() string
	purego.RegisterLibFunc(&version, ptr, "extism_version")
	Print(version())
	return nil
}

func LibCmd() *cobra.Command {
	lib := &cobra.Command{
		Use:   "lib",
		Short: "Manage libextism",
	}

	// Install
	installArgs := &libInstallArgs{}
	libInstall := &cobra.Command{
		Use:          "install",
		Short:        "Install libextism",
		SilenceUsage: true,
		RunE:         RunArgs(runLibInstall, installArgs),
	}
	libInstall.Flags().StringVar(&installArgs.version, "version", "",
		"Install a specified Extism version, `git` or `latest` can be used to specify the latest from git and no version will default to the most recent release")
	libInstall.Flags().StringVar(&installArgs.os, "os", runtime.GOOS, "The target OS: linux, darwin, windows")
	libInstall.Flags().StringVar(&installArgs.arch, "arch", runtime.GOARCH, "The target architecture: x86_64, aarch64")
	libInstall.Flags().StringVar(&installArgs.prefix, "prefix", "/usr/local",
		"Prefix for libextism installation. libextism will be copied to $prefix/$libdir and extism.h will be copied to $prefix/$includedir")
	libInstall.Flags().StringVar(&installArgs.libDir, "libdir", "lib", "The shared object will be installed to $prefix/$libdir")
	libInstall.Flags().StringVar(&installArgs.includeDir, "includedir", "include", "The header file will be installed to $prefix/$includedir")
	lib.AddCommand(libInstall)

	// Uninstall
	uninstallArgs := &libUninstallArgs{}
	libUninstall := &cobra.Command{
		Use:          "uninstall",
		Short:        "Uninstall libextism",
		SilenceUsage: true,
		RunE:         RunArgs(runLibUninstall, uninstallArgs),
	}
	libUninstall.Flags().StringVar(&uninstallArgs.prefix, "prefix", "/usr/local",
		"Prefix for existing libextism installation")
	libUninstall.Flags().StringVar(&uninstallArgs.libDir, "libdir", "lib", "The shared object will be removed from $prefix/$libdir")
	libUninstall.Flags().StringVar(&uninstallArgs.includeDir, "includedir", "include", "The header file will be removed from $prefix/$includedir")
	lib.AddCommand(libUninstall)

	// Versions
	libVersions := &cobra.Command{
		Use:          "versions",
		Short:        "List available Extism versions",
		SilenceUsage: true,
		RunE:         runLibVersions,
	}
	lib.AddCommand(libVersions)

	// Check
	libCheck := &cobra.Command{
		Use:          "check",
		Short:        "Check for libextism installation",
		SilenceUsage: true,
		RunE:         runLibCheck,
	}
	lib.AddCommand(libCheck)

	return lib
}
