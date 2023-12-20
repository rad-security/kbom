package model

import (
	"testing"
)

func TestImagePkgID(t *testing.T) {
	testCases := []struct {
		name        string
		image       Image
		expectedID  string
		expectedErr error
	}{
		{
			name: "VersionAndDigest",
			image: Image{
				FullName: "full_name",
				Name:     "repo/name",
				Version:  "version",
				Digest:   "sha256:digest",
			},
			expectedID: "pkg:oci/name@sha256%3Adigest?repository_url=repo%2Fname&tag=version",
		},
		{
			name: "VersionOnly",
			image: Image{
				FullName: "full_name",
				Name:     "repo/name",
				Version:  "version",
			},
			expectedID: "pkg:oci/name?repository_url=repo%2Fname&tag=version",
		},
		{
			name: "DigestOnly",
			image: Image{
				FullName: "full_name",
				Name:     "repo/subrepo/name",
				Digest:   "sha256:digest",
			},
			expectedID: "pkg:oci/name@sha256%3Adigest?repository_url=repo%2Fsubrepo%2Fname",
		},
		{
			name: "NoVersionOrDigest",
			image: Image{
				FullName: "full_name",
				Name:     "repo/name",
			},
			expectedID: "pkg:oci/name?repository_url=repo%2Fname",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.image.PkgID()
			if result != tc.expectedID {
				t.Errorf("Expected %s, but got %s", tc.expectedID, result)
			}
		})
	}
}
