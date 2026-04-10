// Package app contains legacy lifecycle wiring retained for compatibility tests.
//
// Deprecated in v1.1 and scheduled for removal in v1.3:
// runtime compile/dispatch forwarding that previously lived here now routes
// directly to internal/runtime. New code must not import internal/app except
// approved bridge points that exercise historical lifecycle behavior.
package app
