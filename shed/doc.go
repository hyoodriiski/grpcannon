// Package shed provides a lightweight load-shedding primitive that drops
// incoming requests when the number of concurrent in-flight calls exceeds
// a configurable ceiling, protecting downstream services from overload.
package shed
