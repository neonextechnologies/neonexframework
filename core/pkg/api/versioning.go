package api

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// Version represents an API version
type Version struct {
	Major      int
	Minor      int
	Patch      int
	PreRelease string
}

// String returns the version as a string (e.g., "v1.2.3")
func (v Version) String() string {
	version := fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
	if v.PreRelease != "" {
		version += "-" + v.PreRelease
	}
	return version
}

// ShortString returns the short version (e.g., "v1")
func (v Version) ShortString() string {
	return fmt.Sprintf("v%d", v.Major)
}

// ParseVersion parses a version string
func ParseVersion(version string) (*Version, error) {
	version = strings.TrimPrefix(version, "v")
	re := regexp.MustCompile(`^(\d+)(?:\.(\d+))?(?:\.(\d+))?(?:-([a-zA-Z0-9\-\.]+))?$`)
	matches := re.FindStringSubmatch(version)

	if matches == nil {
		return nil, fmt.Errorf("invalid version format: %s", version)
	}

	v := &Version{}
	v.Major, _ = strconv.Atoi(matches[1])
	if matches[2] != "" {
		v.Minor, _ = strconv.Atoi(matches[2])
	}
	if matches[3] != "" {
		v.Patch, _ = strconv.Atoi(matches[3])
	}
	if matches[4] != "" {
		v.PreRelease = matches[4]
	}

	return v, nil
}

// VersionManager manages API versions
type VersionManager struct {
	versions        map[string]*Version
	defaultVersion  *Version
	deprecatedAfter map[string]*Version // Version -> Deprecated after version
}

// NewVersionManager creates a new version manager
func NewVersionManager(defaultVersion string) (*VersionManager, error) {
	v, err := ParseVersion(defaultVersion)
	if err != nil {
		return nil, err
	}

	return &VersionManager{
		versions:        make(map[string]*Version),
		defaultVersion:  v,
		deprecatedAfter: make(map[string]*Version),
	}, nil
}

// RegisterVersion registers a new API version
func (vm *VersionManager) RegisterVersion(version string) error {
	v, err := ParseVersion(version)
	if err != nil {
		return err
	}
	vm.versions[v.ShortString()] = v
	return nil
}

// DeprecateVersion marks a version as deprecated
func (vm *VersionManager) DeprecateVersion(version, deprecatedAfter string) error {
	v, err := ParseVersion(version)
	if err != nil {
		return err
	}
	after, err := ParseVersion(deprecatedAfter)
	if err != nil {
		return err
	}
	vm.deprecatedAfter[v.ShortString()] = after
	return nil
}

// IsDeprecated checks if a version is deprecated
func (vm *VersionManager) IsDeprecated(version string) bool {
	v, _ := ParseVersion(version)
	if v == nil {
		return false
	}
	_, deprecated := vm.deprecatedAfter[v.ShortString()]
	return deprecated
}

// GetVersion returns a version by string
func (vm *VersionManager) GetVersion(version string) (*Version, error) {
	v, err := ParseVersion(version)
	if err != nil {
		return nil, err
	}
	if storedV, ok := vm.versions[v.ShortString()]; ok {
		return storedV, nil
	}
	return nil, fmt.Errorf("version %s not found", version)
}

// GetDefaultVersion returns the default version
func (vm *VersionManager) GetDefaultVersion() *Version {
	return vm.defaultVersion
}

// VersionMiddleware creates middleware for API versioning
func (vm *VersionManager) VersionMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get version from path (e.g., /api/v1/users)
		path := c.Path()
		re := regexp.MustCompile(`/api/(v\d+)`)
		matches := re.FindStringSubmatch(path)

		var version *Version
		if len(matches) > 1 {
			version, _ = vm.GetVersion(matches[1])
		}

		// If no version in path, try header
		if version == nil {
			versionHeader := c.Get("API-Version", "")
			if versionHeader != "" {
				version, _ = vm.GetVersion(versionHeader)
			}
		}

		// If still no version, try Accept header
		if version == nil {
			accept := c.Get("Accept", "")
			if strings.Contains(accept, "version=") {
				re := regexp.MustCompile(`version=v?(\d+)`)
				matches := re.FindStringSubmatch(accept)
				if len(matches) > 1 {
					version, _ = vm.GetVersion("v" + matches[1])
				}
			}
		}

		// Use default version if none specified
		if version == nil {
			version = vm.defaultVersion
		}

		// Store version in context
		c.Locals("api_version", version)

		// Add version to response headers
		c.Set("API-Version", version.ShortString())

		// Add deprecation warning if needed
		if vm.IsDeprecated(version.ShortString()) {
			deprecatedAfter := vm.deprecatedAfter[version.ShortString()]
			c.Set("Warning", fmt.Sprintf("299 - \"API version %s is deprecated. Please upgrade to %s\"",
				version.ShortString(), deprecatedAfter.ShortString()))
			c.Set("Sunset", "Thu, 31 Dec 2026 23:59:59 GMT") // TODO: Make configurable
		}

		return c.Next()
	}
}

// GetVersionFromContext retrieves the API version from context
func GetVersionFromContext(c *fiber.Ctx) *Version {
	if v := c.Locals("api_version"); v != nil {
		if version, ok := v.(*Version); ok {
			return version
		}
	}
	return nil
}

// VersionedRouter creates a versioned router group
func VersionedRouter(app *fiber.App, version string) fiber.Router {
	return app.Group(fmt.Sprintf("/api/%s", version))
}

// CreateVersionedRoutes creates route groups for multiple versions
func CreateVersionedRoutes(app *fiber.App, versions []string) map[string]fiber.Router {
	routers := make(map[string]fiber.Router)
	for _, version := range versions {
		routers[version] = VersionedRouter(app, version)
	}
	return routers
}
