package gittools

import (
	"io"
	"os"

	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/transport/ssh"
	"github.com/sebastianRau/deployer/pkg/verbose"
)

func UpdateKeyBytes(url string, directory string, pemBytes []byte, password string, v io.Writer) (string, bool, error) {

	publicKeys, err := ssh.NewPublicKeys("git", pemBytes, os.ModeAppend.String())
	if err != nil {
		verbose.Fprintf(v, "generate publickeys failed: %s\n", err.Error())
		return "", false, err
	}

	return updateGitRepo(url, directory, publicKeys, v)

}

func UpdateKeyFilename(url string, directory string, privateKeyFile string, password string, v io.Writer) (string, bool, error) {

	_, err := os.Stat(privateKeyFile)
	if err != nil {
		verbose.Fprintf(v, "read file %s failed %s\n", privateKeyFile, err.Error())
		return "", false, err
	}

	publicKeys, err := ssh.NewPublicKeysFromFile("git", privateKeyFile, password)
	if err != nil {
		verbose.Fprintf(v, "generate publickeys failed: %s\n", err.Error())
		return "", false, err
	}

	return updateGitRepo(url, directory, publicKeys, v)

}

func UpdateNoKeyurl(url string, directory string, v io.Writer) (string, bool, error) {
	return updateGitRepo(url, directory, nil, v)
}

func updateGitRepo(url string, directory string, publicKeys *ssh.PublicKeys, v io.Writer) (string, bool, error) {

	gitCloneOptions := &git.CloneOptions{
		URL:      url,
		Progress: v,
	}

	gitPullOptions := &git.PullOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
	}

	if publicKeys != nil {
		gitCloneOptions.Auth = publicKeys
		gitPullOptions.Auth = publicKeys
	}

	// open git repo
	verbose.Fprintf(v, "open git repo %s\n", directory)
	r, err := git.PlainOpen(directory)
	if err != nil {

		verbose.Fprintf(v, "can't open repo %s\n", directory)

		// Clone the given repository to the given directory
		verbose.Fprintf(v, "git clone %s into %s\n ", url, directory)
		r, err = git.PlainClone(directory, false, gitCloneOptions)
		if err != nil {
			return "", false, err
		}

		ref, _ := r.Head()

		// repo was new cloned, no pull needed. in call cases it's a new hash
		return ref.Hash().String(), true, nil
	}

	// Pull the latest changes from the origin remote and merge into the current branch
	verbose.Fprintf(v, "git pull origin\n")
	w, err := r.Worktree()
	if err != nil {
		return "", false, err
	}

	err = w.Pull(gitPullOptions)
	ref, _ := r.Head()
	if err != nil {
		if err.Error() == "already up-to-date" {
			verbose.Fprintf(v, "%s\n", err.Error())
			return ref.Hash().String(), false, nil
		} else {
			return "", false, nil
		}
	}

	return ref.Hash().String(), true, nil
}
