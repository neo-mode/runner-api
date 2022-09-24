package runner

import (
	"bytes"
	"io"
	"os"
	"os/exec"
)

type GitError string

const origin = "origin/"

func UpdateRefs(dir, targetName, sourceName, refSHA, repoURL string) (bool, error) {

	var err = os.Mkdir(dir, 0755)
	if err == nil {
		return false, git(dir, nil, "clone", repoURL, ".")
	}

	if !os.IsExist(err) {
		return false, err
	}

	if targetName == "" || sourceName == "" {
		return false, git(dir, nil, "fetch", repoURL, refSHA)
	}

	var data bytes.Buffer
	if err = git(dir, &data, "fetch", repoURL, "+refs/heads/"+targetName+":refs/remotes/"+origin+targetName, "+refs/heads/"+sourceName+":refs/remotes/"+origin+sourceName); err != nil {
		return false, err
	}
	return findData(data.String(), targetName), nil
}

func Checkout(dir, targetRef, sourceRef string) (bool, error) {

	var err = git(dir, nil, "checkout", targetRef)
	if sourceRef == "" || err != nil {
		return false, err
	}

	var data bytes.Buffer
	if err = git(dir, &data, "merge", "--no-ff", sourceRef); err != nil {
		git(dir, nil, "merge", "--reset")
		return false, err
	}

	var _data = data.String()
	return _data == "Already up to date.\n" || _data == "Merge made by the 'recursive' strategy.\n", nil
}

func GetRef(dir, refName string) string {

	var f, _ = os.Open(dir + "/.git/" + refName)
	if f == nil {
		return ""
	}

	var data = make([]byte, 40)
	f.Read(data)
	f.Close()

	return string(data)
}

func SetRef(dir, targetDir, targetName, sourceName string) {

	var src, _ = os.Open(dir + "/.git/" + sourceName)
	if src == nil {
		return
	}

	targetDir = dir + "/.git/" + targetDir
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		src.Close()
		return
	}

	var dst, _ = os.OpenFile(targetDir+"/"+targetName, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
	if dst == nil {
		src.Close()
		return
	}

	io.Copy(dst, src)
	src.Close()
	dst.Close()
}

func findData(text, targetName string) bool {

	if text == "" {
		return false
	}

	var tln = len(targetName) + len(origin) + 3
	var ln = len(text) - tln

	for i := 0; i < ln; i++ {
		if text[i:i+tln] == "-> "+origin+targetName {
			return true
		}
	}

	return false
}

func git(dir string, output io.Writer, args ...string) error {

	var cmd = exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Stdout = output
	cmd.Stderr = output

	if err := cmd.Run(); err != nil {
		return GitError("")
	}

	return nil
}

func (git GitError) Error() string {
	return string(git)
}
