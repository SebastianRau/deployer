package ostools

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/sebastianRau/deployer/pkg/verbose"
)

// Mkdir creates all given directries
//
// path could contain any number of folders
// e.g. "./temp/bla/bla/bla/" all four foldes will be created if not existing
// it returns any error
func Mkdir(path string, v io.Writer) error {
	verbose.Fprintf(v, "mkdir %s", path)
	return ensureDir(path, v)
}

//	Copy  makes a file copy from src to dest
//
// src can contain * wildcard for any matching file  in the folder
// e.g. "temp/*.go" for all files with ".go" ending in temp folder
// --> /temp/test.go
// src can contain ** wildcard for any matching file in the folder and all subfolders
// e.g. "temp/*.go" for all files with ".go" ending in temp folder and all subfolders
// --> temp/test.go
// --> test/bla/test.go
// dest can be a file or a folder, if dest folder does not exist, it will be created
func Copy(src string, dest string, ignoreDotFiles bool, v io.Writer) error {

	sourceHasWildcard := strings.Contains(src, "*")

	sourceIsFolder, err := isFolder(src)
	if err != nil {
		sourceIsFolder = false
	}

	destIsFolder, err := isFolder(dest)
	if err != nil {
		destIsFolder = false
		err = ensureDir(filepath.Dir(dest), v)
		if err != nil {
			return err
		}
	}

	if sourceHasWildcard || sourceIsFolder {
		files, err := listFiles(src, ignoreDotFiles)
		if err != nil {
			return err
		}

		for _, file := range files {
			if destIsFolder {
				_, postPath, _ := strings.Cut(filepath.Dir(file), filepath.Dir(src))
				target := filepath.Clean(dest + postPath + "/" + filepath.Base(file))
				err := Copy(file, target, ignoreDotFiles, v)
				if err != nil {
					return err
				}
			}
		}
	} else {
		err = ensureDir(filepath.Dir(dest), v)
		if err != nil {
			return err
		}

		err := CopyFile(src, dest)
		if err != nil {
			return err
		}
		verbose.Fprintf(v, "Copy from: %s to: %s\n", src, dest)

		return err
	}
	return nil
}

func CopyFile(source string, dest string) error {
	src, err := os.Open(source)
	if err != nil {
		return err
	}
	defer src.Close()

	err = os.MkdirAll(filepath.Dir(dest), 0666)
	if err != nil {
		return err
	}

	dst, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}

	info, err := os.Stat(source)
	if err != nil {
		err = os.Chmod(dest, info.Mode())
		if err != nil {
			return err
		}
	}

	return nil
}

// Deletes all given files oder folders
//
// dir can contain a * wildcard
// e.g Delete("./temp/") will delete the folder "temp" with all content
// e.g Delete("./temp/*") will delete all the content, but not the folder itself
func Delete(dir string, v io.Writer) error {
	usingWildcard := strings.Contains(dir, "*")
	isFolder, _ := isFolder(dir)

	if usingWildcard {
		files, err := listFiles(dir, false)
		if err != nil {
			return err
		}

		for _, f := range files {
			verbose.Fprintf(v, "rm %s\n", f)
			err := os.Remove(f)
			if err != nil {
				return err
			}
		}

	} else {
		if isFolder {
			verbose.Fprintf(v, "rm -R %s\n", dir)
			err := os.RemoveAll(dir)
			return err
		} else {
			verbose.Fprintf(v, "rm %s\n", dir)
			err := os.Remove(dir)
			return err
		}
	}
	return nil
}

func isFolder(path string) (bool, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return fileInfo.IsDir(), nil
}

func ensureDir(dir string, v io.Writer) error {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		verbose.Fprintf(v, "mkdir: %s\n", dir)
		err = os.MkdirAll(dir, 0755)
		return err
	}
	return nil
}

func evaluatePathPattern(pattern string) (path string, prefix string, suffix string, wildcard bool) {
	path = filepath.Dir(pattern)
	file := filepath.Base(pattern)
	wildcard = false
	usingWildcard := strings.Contains(file, "*")

	if usingWildcard {
		s := strings.SplitN(file, "*", 2)
		prefix = s[0]

		if len(s) > 1 {
			suffix = s[1]
			wildcard = strings.Contains(suffix, "*")
			suffix = strings.ReplaceAll(suffix, "*", "")
		}
	}

	return path, prefix, suffix, wildcard
}

func listFiles(pattern string, ignoreDotFiles bool) ([]string, error) {
	var files []string
	startpath, prefix, suffix, wildcard := evaluatePathPattern(pattern)

	err := filepath.Walk(startpath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		filename := filepath.Base(path)
		dirName := filepath.Dir(path)
		appendFile := true
		_, postPath, _ := strings.Cut(dirName, startpath)

		// no Dirs will be add
		appendFile = appendFile && !info.IsDir()

		// check file prefix, but add only if path is exact equal
		if prefix != "" {
			appendFile = appendFile && postPath == ""
			appendFile = appendFile && strings.HasPrefix(filename, prefix)
		}

		if suffix != "" {
			if !wildcard {
				appendFile = appendFile && postPath == ""
			}
			appendFile = appendFile && strings.HasSuffix(path, suffix)
		}

		if ignoreDotFiles {
			appendFile = appendFile && !strings.HasPrefix(filename, ".")
			s := strings.Split(dirName, "/")
			for _, dirPart := range s {
				appendFile = appendFile && !strings.HasPrefix(dirPart, ".")
			}

		}

		if appendFile {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}
