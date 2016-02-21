# httputil: Make better web apps without frameworks

This library provides utilities for building better web tools, but
without buying into a framework.

## Resolver

The Resolver is a tool to assist HTTP applications resolve from a
request to a route.

For example, you may wish to define a route like this:

```
GET /foo/*/bar
```

That route will handle any requests that come in and match the given
pattern. When a request comes in for `GET /foo/123/bar`, it should be
matched to the route above.

The `httputil.Resolver` provides this capability.

## Notes

Parts of this library were extracted from Cookoo and refactored.
