package chezmoi

import (
	"errors"
	"io/fs"
	"os"
	"os/exec"
	"runtime"

	"github.com/bmatcuk/doublestar/v4"
	vfs "github.com/twpayne/go-vfs/v4"
	"go.uber.org/multierr"

	"github.com/twpayne/chezmoi/v2/internal/chezmoilog"
)

// A RealSystemOption sets an option on a RealSystem.
type RealSystemOption func(*RealSystem)

// Glob implements System.Glob.
func (s *RealSystem) Glob(pattern string) ([]string, error) {
	return doublestar.Glob(s.UnderlyingFS(), pattern)
}

// IdempotentCmdCombinedOutput implements System.IdempotentCmdCombinedOutput.
func (s *RealSystem) IdempotentCmdCombinedOutput(cmd *exec.Cmd) ([]byte, error) {
	return chezmoilog.LogCmdCombinedOutput(cmd)
}

// IdempotentCmdOutput implements System.IdempotentCmdOutput.
func (s *RealSystem) IdempotentCmdOutput(cmd *exec.Cmd) ([]byte, error) {
	return chezmoilog.LogCmdOutput(cmd)
}

// Link implements System.Link.
func (s *RealSystem) Link(oldname, newname AbsPath) error {
	return s.fileSystem.Link(oldname.String(), newname.String())
}

// Lstat implements System.Lstat.
func (s *RealSystem) Lstat(filename AbsPath) (fs.FileInfo, error) {
	return s.fileSystem.Lstat(filename.String())
}

// Mkdir implements System.Mkdir.
func (s *RealSystem) Mkdir(name AbsPath, perm fs.FileMode) error {
	return s.fileSystem.Mkdir(name.String(), perm)
}

// RawPath implements System.RawPath.
func (s *RealSystem) RawPath(absPath AbsPath) (AbsPath, error) {
	rawAbsPath, err := s.fileSystem.RawPath(absPath.String())
	if err != nil {
		return EmptyAbsPath, err
	}
	return NewAbsPath(rawAbsPath), nil
}

// ReadDir implements System.ReadDir.
func (s *RealSystem) ReadDir(name AbsPath) ([]fs.DirEntry, error) {
	return s.fileSystem.ReadDir(name.String())
}

// ReadFile implements System.ReadFile.
func (s *RealSystem) ReadFile(name AbsPath) ([]byte, error) {
	return s.fileSystem.ReadFile(name.String())
}

// RemoveAll implements System.RemoveAll.
func (s *RealSystem) RemoveAll(name AbsPath) error {
	return s.fileSystem.RemoveAll(name.String())
}

// Rename implements System.Rename.
func (s *RealSystem) Rename(oldpath, newpath AbsPath) error {
	return s.fileSystem.Rename(oldpath.String(), newpath.String())
}

// RunCmd implements System.RunCmd.
func (s *RealSystem) RunCmd(cmd *exec.Cmd) error {
	return chezmoilog.LogCmdRun(cmd)
}

// RunIdempotentCmd implements System.RunIdempotentCmd.
func (s *RealSystem) RunIdempotentCmd(cmd *exec.Cmd) error {
	return chezmoilog.LogCmdRun(cmd)
}

// RunScript implements System.RunScript.
func (s *RealSystem) RunScript(scriptname RelPath, dir AbsPath, data []byte, interpreter *Interpreter) (err error) {
	// Write the temporary script file. Put the randomness at the front of the
	// filename to preserve any file extension for Windows scripts.
	f, err := os.CreateTemp("", "*."+scriptname.Base())
	if err != nil {
		return
	}
	defer func() {
		err = multierr.Append(err, os.RemoveAll(f.Name()))
	}()

	// Make the script private before writing it in case it contains any
	// secrets.
	if runtime.GOOS != "windows" {
		if err = f.Chmod(0o700); err != nil {
			return
		}
	}
	_, err = f.Write(data)
	err = multierr.Append(err, f.Close())
	if err != nil {
		return
	}

	cmd := interpreter.ExecCommand(f.Name())
	cmd.Dir, err = s.getScriptWorkingDir(dir)
	if err != nil {
		return err
	}
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return s.RunCmd(cmd)
}

// Stat implements System.Stat.
func (s *RealSystem) Stat(name AbsPath) (fs.FileInfo, error) {
	return s.fileSystem.Stat(name.String())
}

// UnderlyingFS implements System.UnderlyingFS.
func (s *RealSystem) UnderlyingFS() vfs.FS {
	return s.fileSystem
}

// getScriptWorkingDir returns the script's working directory.
//
// If this is a before_ script then the requested working directory may not
// actually exist yet, so search through the parent directory hierarchy till
// we find a suitable working directory.
func (s *RealSystem) getScriptWorkingDir(dir AbsPath) (string, error) {
	// This should always terminate because dir will eventually become ".", i.e.
	// the current directory.
	for {
		switch info, err := s.Stat(dir); {
		case err == nil && info.IsDir():
			// dir exists and is a directory. Use it.
			dirRawAbsPath, err := s.RawPath(dir)
			if err != nil {
				return "", err
			}
			return dirRawAbsPath.String(), nil
		case err == nil || errors.Is(err, fs.ErrNotExist):
			// Either dir does not exist, or it exists and is not a directory.
			dir = dir.Dir()
		default:
			// Some other error occurred.
			return "", err
		}
	}
}
