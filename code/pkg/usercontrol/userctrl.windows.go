//go:build !linux
// +build !linux

// The usercontrol package allows a process to switch from running as root to
// running as another user.  This is often done for security reasons by programs
// such as web servers.  A web server may need root privilege to set up but can
// then run as an ordinaru user, posing less of a security risk.
//
// A process can only use this package if it's running as root and on a UNIX-style
// system such as Linux because the necessary functionality only exists on those
// systems and only the root user can switch to another user without specifying a
// password.  In the systems that this solution is aimed at, users are typically
// authenticated using certificates and don't have a password, so only root can
// become such a user.
//
// The solution uses Seteuid(targetID) to switch to the user with the given target
// ID.  The syscall package (now deprecated) and its successor golang.org/x/sys both
// contain different functions on different systems.  Under Linux they contain
// Setuid but under Windows they don't.  If this package contained a direct call to
// Setuid in one of those packages, it and anything that imported it wouldn't compile
// under Windows, which is a nuisance.  I haven't found a good way around this
// problem so I use the horrible solution of two source files, one for Linux and
// another for Windows.  The user has to copy the appropriate file or set a link to
// it before compiling the solution.
//
// Tthe Windows version of the package contains functions that return a "not
// implemented" error if called.  They are only there to satisfy the compiler.
package usercontrol

import (
	"errors"
)

// Getuid gets the current user ID.  The Windows version always returns -1
// which is not a valid user ID.
func Getuid() int {
	return -1
}

// Setuid switches the effective user to the user with the given user ID.  The
// Windows version always returns a "not impleented" error.
func Setuid(targetID int) error {
	return errors.New("not implemented")
}
