package httputil

import (
	"errors"
	"fmt"
	"net/http"
	"path"
	"strings"
)

var ErrRouteNotFound = errors.New("route not found")

// Resolver resolves a request or path to a route.
//
// Given a registered set of route strings, this will take an *http.Request or
// a verb-path string and resolve it to the best match.
//
// For example, given the registered path "GET /foo/*", the request "GET /foo/bar"
// will resolve to "GET /foo/*", while "GET /foo/bar/baz" will not.
//
// The path matching syntax follows the one in the `path` package of Go.
//
// HTTP Verbs:
// This resolver also allows you to specify verbs at the beginning of a path:
// - "GET /foo" and "POST /foo" are separate (but legal) paths. "* /foo" will allow any verb.
// - There are no constraints on verb name. Thus, verbs like WebDAV's PROPSET are fine, too. Or you can
//   make up your own.
//
// IMPORTANT! When it comes to matching route patterns against paths, ORDER IS
// IMPORTANT. Routes are evaluated in order. So if two rules (GET /a/b* and GET /a/bc) are
// both defined, the incomming request GET /a/bc will match whichever route is
// defined first. See the unit tests for examples.
//
// The `**` and `/**` Wildcards:
// =============================
//
// In addition to the paths described in the `path` package of Go's core, two
// extra wildcard sequences are defined:
//
// - `**`: Match everything.
// - `/**`: a suffix that matches any sub-path.
//
// The `**` wildcard works in ONLY ONE WAY:  If the path is declared as `**`, with nothing else,
// then any path will match.
//
// VALID: `**`, `GET /foo/**`, `GET /**`
// NOT VALID: `GET **`, `**/foo`, `foo/**/bar`
//
// The `/**` suffix can only be added to the end of a path, and says "Match
// any subpath under this".
//
// Examples:
// - URI paths "POST /foo", "GET /a/b/c", and "hello" all match "**". (The ** rule
//   can be very dangerous for this reason.)
// - URI path "POST /assets/images/foo/bar/baz.jpg" matches "POST /assets/**"
//
// The behavior for rules that contain `/**` anywhere other than the end
// have undefined behavior.
//
// The list of paths is not modifed after the resolver is constructed, so this
// can safely be used concurrently.
type Resolver struct {
	paths []string
}

// NewResolver creates a new Resolver.
//
// This creates a resolver that knows about the given paths. Paths are evaluated
// in order.
//
// Paths may contain HTTP verbs and wildcards as described above.
func NewResolver(paths []string) *Resolver {
	return &Resolver{
		paths: paths,
	}
}

// Resolve takes an HTTP request and attempts to resolve it.
//
// When successful, it will return the path that it matched.
//
// When no path is found, an ErrRouteNotFound error is returned.
//
// Errors in path format, regexp compilation, and so on may also be returned.
func (r *Resolver) Resolve(req *http.Request) (string, error) {
	return r.ResolvePath(req.Method + " " + req.URL.Path)
}

// Resolve a path name based using path patterns.
//
// For general usage, this is intended to be used with verbed paths (e.g.
// "GET /foo"). If the verb is omitted, it will only match paths that do not
// have a verb. It should be noted that the Resolve() method will ONLY match
// verbed paths.
//
// The resolver will take a path and attempt to match it against one of the
// paths the Resolver.
//
// This resolver is designed to match path-like strings to path patterns. For example,
// the path `GET /foo/bar/baz` may match routes like `* /foo/*/baz` or `GET /foo/bar/*`
func (r *Resolver) ResolvePath(pathName string) (string, error) {
	// HTTP verb support naturally falls out of the fact that spaces in paths
	// are legal in UNIXy systems, while illegal in URI paths. So presently we
	// do no special handling for verbs. Yay for simplicity.
	for _, pattern := range r.paths {
		if strings.HasSuffix(pattern, "**") {
			if ok, _ := r.subtreeMatch(pathName, pattern); ok {
				return pattern, nil
			}
			// Fall through to the regular pattern matcher, since ** is legal.
		}

		if ok, err := path.Match(pattern, pathName); ok && err == nil {
			return pattern, nil
		} else if err != nil {
			// Bad pattern
			return pathName, err
		}
	}
	return pathName, ErrRouteNotFound
}

func (r *Resolver) subtreeMatch(pathName, pattern string) (bool, error) {
	if pattern == "**" {
		return true, nil
	}

	// Find out how many slashes we have.
	countSlash := strings.Count(pattern, "/")

	if countSlash == 0 {
		return false, fmt.Errorf("illegal ** pattern: %s", pattern)
	}

	// Add 2 for verb plus trailer.
	parts := strings.SplitN(pathName, "/", countSlash+1)
	prefix := strings.Join(parts[0:countSlash], "/")

	subpattern := strings.Replace(pattern, "/**", "", -1)
	if ok, err := path.Match(subpattern, prefix); ok && err == nil {
		return true, nil
	} else if err != nil {
		return false, err
	}
	return false, nil
}
