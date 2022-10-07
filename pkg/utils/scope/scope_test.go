package scope

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestManagerValidate(t *testing.T) {
	t.Run("url", func(t *testing.T) {
		manager, err := NewManager([]string{`example`}, []string{`logout\.php`}, "dn", false)
		require.NoError(t, err, "could not create scope manager")

		parsed, _ := url.Parse("https://test.com/index.php/example")
		validated, err := manager.Validate(parsed, "test.com")
		require.NoError(t, err, "could not validate url")
		require.True(t, validated, "could not get correct in-scope validation")

		parsed, _ = url.Parse("https://test.com/logout.php")
		validated, err = manager.Validate(parsed, "another.com")
		require.NoError(t, err, "could not validate url")
		require.False(t, validated, "could not get correct out-scope validation")
	})
	t.Run("host", func(t *testing.T) {
		t.Run("dn", func(t *testing.T) {
			manager, err := NewManager(nil, nil, "dn", false)
			require.NoError(t, err, "could not create scope manager")

			parsed, _ := url.Parse("https://testanother.com/index.php")
			validated, err := manager.Validate(parsed, "test.com")
			require.NoError(t, err, "could not validate host")
			require.True(t, validated, "could not get correct in-scope validation")
		})
		t.Run("rdn", func(t *testing.T) {
			manager, err := NewManager(nil, nil, "rdn", false)
			require.NoError(t, err, "could not create scope manager")

			parsed, _ := url.Parse("https://subdomain.example.com/logout.php")
			validated, err := manager.Validate(parsed, "example.com")
			require.NoError(t, err, "could not validate host")
			require.True(t, validated, "could not get correct in-scope validation")
		})
		t.Run("fqdn", func(t *testing.T) {
			manager, err := NewManager(nil, nil, "fqdn", false)
			require.NoError(t, err, "could not create scope manager")

			parsed, _ := url.Parse("https://test.com/index.php")
			validated, err := manager.Validate(parsed, "test.com")
			require.NoError(t, err, "could not validate host")
			require.True(t, validated, "could not get correct in-scope validation")

			parsed, _ = url.Parse("https://subdomain.example.com/logout.php")
			validated, err = manager.Validate(parsed, "example.com")
			require.NoError(t, err, "could not validate host")
			require.False(t, validated, "could not get correct out-scope validation")

			parsed, _ = url.Parse("https://example.com/logout.php")
			validated, err = manager.Validate(parsed, "another.com")
			require.NoError(t, err, "could not validate host")
			require.False(t, validated, "could not get correct out-scope validation")
		})
	})
}

func TestGetDomainRDNandDN(t *testing.T) {
	rdn, dn, err := getDomainRDNandRDN("test.projectdiscovery.io")
	require.Nil(t, err, "could not get domain rdn and dn")
	require.Equal(t, "projectdiscovery.io", rdn, "could not get correct rdn")
	require.Equal(t, "projectdiscovery", dn, "could not get correct dn")
}
