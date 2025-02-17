package distribution

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/osbuild/image-builder/internal/common"
)

func TestDistributionFile_Architecture(t *testing.T) {
	adr, err := LoadDistroRegistry("../../distributions")
	require.NoError(t, err)
	d, err := adr.Available(false).Get("centos-8")
	require.NoError(t, err)

	arch, err := d.Architecture("x86_64")
	require.NoError(t, err)

	// don't test packages, they are huge
	arch.Packages = nil

	require.Equal(t, &Architecture{
		ImageTypes: []string{"aws", "gcp", "azure", "ami", "vhd", "guest-image", "image-installer", "oci", "vsphere", "vsphere-ova", "wsl"},
		Repositories: []Repository{
			{
				Id:      "baseos",
				Baseurl: common.ToPtr("http://mirror.centos.org/centos/8-stream/BaseOS/x86_64/os/"),
				Rhsm:    false,
			},
			{
				Id:      "appstream",
				Baseurl: common.ToPtr("http://mirror.centos.org/centos/8-stream/AppStream/x86_64/os/"),
				Rhsm:    false,
			},
			{
				Id:      "extras",
				Baseurl: common.ToPtr("http://mirror.centos.org/centos/8-stream/extras/x86_64/os/"),
				Rhsm:    false,
			},
			{
				Id:            "google-compute-engine",
				Baseurl:       common.ToPtr("https://packages.cloud.google.com/yum/repos/google-compute-engine-el8-x86_64-stable"),
				Rhsm:          false,
				ImageTypeTags: []string{"gcp"},
			},
			{
				Id:            "google-cloud-sdk",
				Baseurl:       common.ToPtr("https://packages.cloud.google.com/yum/repos/cloud-sdk-el8-x86_64"),
				Rhsm:          false,
				ImageTypeTags: []string{"gcp"},
			},
		},
	}, arch,
	)

	arch, err = d.Architecture("unsupported")
	require.Nil(t, arch)
	require.Error(t, err, "Architecture not supported")
}

func TestArchitecture_FindPackages(t *testing.T) {
	adr, err := LoadDistroRegistry("../../distributions")
	require.NoError(t, err)
	d, err := adr.Available(false).Get("centos-8")
	require.NoError(t, err)

	arch, err := d.Architecture("x86_64")
	require.NoError(t, err)

	pkgs := arch.FindPackages("vim")
	require.ElementsMatch(t, []Package{
		{
			Name:    "vim-minimal",
			Summary: "A minimal version of the VIM editor",
		},
		{
			Name:    "vim-common",
			Summary: "The common files needed by any version of the VIM editor",
		},
		{
			Name:    "vim-enhanced",
			Summary: "A version of the VIM editor which includes recent enhancements",
		},
		{
			Name:    "vim-X11",
			Summary: "The VIM version of the vi editor for the X Window System - GVim",
		},
		{
			Name:    "vim-filesystem",
			Summary: "VIM filesystem layout",
		},
	}, pkgs)

	d, err = adr.Available(true).Get("rhel-84")
	require.NoError(t, err)

	arch, err = d.Architecture("x86_64")
	require.NoError(t, err)

	pkgs = arch.FindPackages("vim")
	require.ElementsMatch(t, []Package{
		{
			Name:    "vim-minimal",
			Summary: "A minimal version of the VIM editor",
		},
		{
			Name:    "vim-common",
			Summary: "The common files needed by any version of the VIM editor",
		},
		{
			Name:    "vim-enhanced",
			Summary: "A version of the VIM editor which includes recent enhancements",
		},
		{
			Name:    "vim-X11",
			Summary: "The VIM version of the vi editor for the X Window System - GVim",
		},
		{
			Name:    "vim-filesystem",
			Summary: "VIM filesystem layout",
		},
	}, pkgs)

	// load the test distributions and check that a distro with no_package_list == true works
	adr, err = LoadDistroRegistry("testdata/distributions")
	require.NoError(t, err)

	d, err = adr.Available(true).Get("no-packages-distro")
	require.NoError(t, err)

	arch, err = d.Architecture("x86_64")
	require.NoError(t, err)

	pkgs = arch.FindPackages("vim")
	require.Nil(t, pkgs)

}

func TestInvalidDistribution(t *testing.T) {
	_, err := readDistribution("../../distributions", "none")
	require.Error(t, err, DistributionNotFound)
}

func TestDistributionFileIsRestricted(t *testing.T) {
	distsDir := "testdata/distributions"

	t.Run("distro is not restricted, has no restricted_access field", func(t *testing.T) {
		d, err := readDistribution(distsDir, "rhel-90")
		require.NoError(t, err)
		actual := d.IsRestricted()
		expected := false
		require.Equal(t, expected, actual)
	})

	t.Run("distro is not restricted, restricted_access field is false", func(t *testing.T) {
		d, err := readDistribution(distsDir, "centos-9")
		require.NoError(t, err)
		actual := d.IsRestricted()
		expected := false
		require.Equal(t, expected, actual)
	})

	t.Run("distro is restricted, restricted_access field is true", func(t *testing.T) {
		d, err := readDistribution(distsDir, "centos-8")
		require.NoError(t, err)
		actual := d.IsRestricted()
		expected := true
		require.Equal(t, expected, actual)
	})
}

func TestArchitecture_validate(t *testing.T) {
	tests := []struct {
		name string
		arch Architecture
		err  error
	}{
		{
			"good",
			Architecture{
				ImageTypes: nil,
				Repositories: []Repository{
					{Baseurl: common.ToPtr("http://example.com/repo1")},
					{Metalink: common.ToPtr("http://example.com/repo2")},
				},
				Packages: nil,
			},
			nil,
		},
		{
			"multiple-sources",
			Architecture{
				ImageTypes: nil,
				Repositories: []Repository{
					{
						Baseurl:  common.ToPtr("http://example.com/repo1"),
						Metalink: common.ToPtr("http://example.com/repo2"),
					},
				},
				Packages: nil,
			},
			RepoSourceError,
		},
		{
			"no-source",
			Architecture{
				ImageTypes: nil,
				Repositories: []Repository{
					{},
				},
				Packages: nil,
			},
			RepoSourceError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.arch.validate()
			require.Equal(t, tt.err, err)
		})
	}
}
