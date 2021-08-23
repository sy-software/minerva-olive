// Package adapters will contain all driver adapters
// These adapters can be an HTTP REST interface, a CLI command, graphQL, etc
// and should contain the logic to invoke a service using parameters from
// the desired external interface.
// The service should not be called directly but by using the corresponding
// port interface
package handlers
