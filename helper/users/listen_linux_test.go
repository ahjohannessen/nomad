// Copyright IBM Corp. 2015, 2026
// SPDX-License-Identifier: BUSL-1.1

//go:build linux

package users

import (
	"net"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/opencontainers/selinux/go-selinux"
	"github.com/shoenig/test/must"
	"golang.org/x/sys/unix"
)

func TestListenTaskSocket_SELinuxLabel(t *testing.T) {
	expected, err := TaskSocketLabel()
	if err != nil || expected == "" {
		t.Skip("SELinux is disabled or has no container policy")
	}

	ln, err := ListenTaskSocket(filepath.Join(t.TempDir(), "test.sock"))
	must.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })

	// The listening socket carries the container type, not the type of the
	// process that created it.
	f, err := ln.(*net.UnixListener).File()
	must.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })
	buf := make([]byte, 256)
	n, err := unix.Fgetxattr(int(f.Fd()), "security.selinux", buf)
	must.NoError(t, err)
	must.Eq(t, expected, strings.TrimRight(string(buf[:n]), "\x00"))
}

func TestListenWithSocketLabel_ResetsThread(t *testing.T) {
	socketLabel, err := TaskSocketLabel()
	if err != nil || socketLabel == "" {
		t.Skip("SELinux is disabled or has no container policy")
	}

	// Stay on one thread so the check below sees the thread the socket was
	// created on.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	ln, reset, err := listenWithSocketLabel(filepath.Join(t.TempDir(), "test.sock"), socketLabel)
	must.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })
	must.True(t, reset)

	current, err := selinux.SocketLabel()
	must.NoError(t, err)
	must.Eq(t, "", current)
}
