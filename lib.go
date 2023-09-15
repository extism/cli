package cli

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/ebitengine/purego"
	"github.com/google/go-github/github"
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
}

type libUninstallArgs struct {
	libArgs
}

func (a *libInstallArgs) SetArgs(args []string) {
	a.args = args
}

func (a *libUninstallArgs) SetArgs(args []string) {
	a.args = args
}

func getReleases(ctx context.Context) (releases []*github.RepositoryRelease, err error) {
	client := github.NewClient(http.DefaultClient)
	releases, _, err = client.Repositories.ListReleases(ctx, "extism", "extism", nil)
	if err != nil {
		return releases, err
	}
	sort.Slice(releases, func(i, j int) bool {
		return releases[i].CreatedAt.Before(releases[j].CreatedAt.Time)
	})
	return releases, nil
}

func findRelease(ctx context.Context, name string) (release *github.RepositoryRelease, err error) {
	releases, err := getReleases(ctx)
	if err != nil {
		return release, err
	}

	for i, rel := range releases {
		if i == len(releases)-1 && name == "" {
			return rel, nil
		} else if rel.TagName != nil && *rel.TagName == name {
			return rel, nil
		}
	}

	return release, errors.New("unable to find release " + name)
}

func assetPrefix() string {

	s := "libextism-"
	if runtime.GOARCH == "amd64" {
		s += "x86_64"
	} else {
		s += runtime.GOARCH
	}
	if runtime.GOOS == "linux" {
		return s + "-unknown-linux-gnu"
	} else if runtime.GOOS == "windows" {
		return s + "-pc-windows-mvsc"
	} else if runtime.GOOS == "darwin" {
		return s + "-apple-darwin"
	}

	return s
}

func sharedLibraryName() string {
	switch runtime.GOOS {
	case "darwin":
		return "libextism.dylib"
	case "windows":
		return "extism.dll"
	default:
		return "libextism.so"
	}
}

func runLibInstall(cmd *cobra.Command, installArgs *libInstallArgs) error {
	if installArgs.version == "git" {
		installArgs.version = "latest"
	}

	rel, err := findRelease(cmd.Context(), installArgs.version)
	if err != nil {
		return err
	}

	for _, asset := range rel.Assets {
		if strings.HasPrefix(asset.GetName(), assetPrefix()) && strings.HasSuffix(asset.GetName(), ".tar.gz") {
			url := asset.GetBrowserDownloadURL()
			fmt.Println("Fetching", url)
			res, err := http.Get(url)
			if err != nil {
				return err
			}
			defer res.Body.Close()
			r, err := gzip.NewReader(res.Body)
			if err != nil {
				return err
			}
			tarReader := tar.NewReader(r)

			for {
				item, err := tarReader.Next()
				if err == io.EOF {
					break
				}

				if strings.HasSuffix(item.Name, getSharedObjectExt()) {
					os.MkdirAll(filepath.Join(installArgs.prefix, installArgs.libDir), 0o755)
					out, err := os.Create(filepath.Join(installArgs.prefix, installArgs.libDir, item.Name))
					if err != nil {
						return err
					}

					fmt.Println("Copying", item.Name, "to", out.Name())
					io.Copy(out, tarReader)
					out.Close()
				} else if strings.HasSuffix(item.Name, ".h") {
					os.MkdirAll(filepath.Join(installArgs.prefix, installArgs.includeDir), 0o755)
					out, err := os.Create(filepath.Join(installArgs.prefix, installArgs.includeDir, item.Name))
					if err != nil {
						return err
					}

					fmt.Println("Copying", item.Name, "to", out.Name())
					io.Copy(out, tarReader)
					out.Close()
				}
			}

		}

	}

	return nil
}

func runLibUninstall(cmd *cobra.Command, uninstallArgs *libUninstallArgs) error {
	soFile := filepath.Join(uninstallArgs.prefix, uninstallArgs.libDir, getSharedObjectFileName())

	fmt.Println("Removing", soFile)
	err := os.Remove(soFile)
	if err != nil {
		return err
	}

	headerFile := filepath.Join(uninstallArgs.prefix, uninstallArgs.includeDir, "extism.h")
	fmt.Println("Removing", headerFile)
	err = os.Remove(headerFile)
	if err != nil {
		return err
	}

	return nil
}

func runLibVersions(cmd *cobra.Command, args []string) error {
	releases, err := getReleases(cmd.Context())
	if err != nil {
		return err
	}

	for _, rel := range releases {
		name := rel.GetTagName()
		if name == "latest" {
			continue
		}

		fmt.Println(name)
	}

	return nil
}

func runLibCheck(cmd *cobra.Command, args []string) error {
	ptr, err := dlopen(sharedLibraryName())
	if err != nil {
		return errors.New("unable to open libextism, no installation detected")
	}

	var version func() string
	purego.RegisterLibFunc(&version, ptr, "extism_version")
	fmt.Println(version())
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
		RunE:         runArgs(runLibInstall, installArgs),
	}
	libInstall.Flags().StringVar(&installArgs.version, "version", "",
		"Install a specified Extism version, `git` can be used to specify the latest from git")
	libInstall.Flags().StringVar(&installArgs.prefix, "prefix", "/usr/local",
		"Prefix to install libextism and extism.h into, the shared object will be copied to $PREFIX/lib and the header will be copied to $PREFIX/include")
	libInstall.Flags().StringVar(&installArgs.libDir, "libdir", "lib", "The shared object will be installed to $prefix/$libdir")
	libInstall.Flags().StringVar(&installArgs.includeDir, "includedir", "include", "The header file will be installed to $prefix/$includedir")
	lib.AddCommand(libInstall)

	// Uninstall
	uninstallArgs := &libUninstallArgs{}
	libUninstall := &cobra.Command{
		Use:          "uninstall",
		Short:        "Uninstall libextism",
		SilenceUsage: true,
		RunE:         runArgs(runLibUninstall, uninstallArgs),
	}
	libUninstall.Flags().StringVar(&uninstallArgs.prefix, "prefix", "/usr/local", "Prefix previously to used to install libextism")
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
