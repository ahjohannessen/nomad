// Copyright IBM Corp. 2015, 2026
// SPDX-License-Identifier: BUSL-1.1

//go:build !linux

package users

import "net"

// ListenTaskSocket creates a unix socket that tasks connect to.
func ListenTaskSocket(path string) (net.Listener, error) {
	return net.Listen("unix", path)
}
