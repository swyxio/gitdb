package db

import (
	"errors"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"time"
	"os"
)

func gitInit() {
	absDbPath, err := filepath.Abs(config.DbPath)
	if err != nil {
		panic(err)
	}

	cmd := exec.Command("git", "-C", absDbPath, "init")
	//log.PutInfo(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		panic(string(out))
	}

	cmd = exec.Command("git", "-C", absDbPath, "remote", "add", "online", config.OnlineRemote)
	//log.PutInfo(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		panic(string(out))
	}

	cmd = exec.Command("git", "-C", absDbPath, "remote", "add", "offline", config.OnlineRemote)
	//log.PutInfo(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		panic(string(out))
	}

	//this function is only called once. I.e when a initializing the database for the
	//very first time. In this case we must pull from the online repo
	if len(config.OnlineRemote) > 0 {
		cmd = exec.Command("git", "-C", absDbPath, "pull", "online", "master")
		//log.PutInfo(utils.CmdToString(cmd))
		if out, err := cmd.CombinedOutput(); err != nil {
			panic(string(out))
		}
	}

	//create .gitignore file
	gitIgnore := filepath.Join(internalDir, "Index")
	gitIgnore = gitIgnore+"\n.id"
	gitIgnoreFile := filepath.Join(absDbPath, ".gitignore")
	if _, err := os.Stat(gitIgnoreFile); err != nil {
		ioutil.WriteFile(gitIgnoreFile, []byte(gitIgnore), 0744)
	}
}

//first attempt to pull from offline DB repo followed by online DB repo
//fails silently, logs error message and determine if we need to put the
//application in an error state
func gitPull() error {
	errorCount := 0
	//do a pull every time we read from the db
	cmd := exec.Command("git", "-C", absDbPath, "pull", "online", "master")
	//log.PutInfo(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		//log.PutError("Failed to pull data from online remote.")
		//log.PutError(string(out))
		println(string(out))
		errorCount++
	}

	if errorCount > 0 {
		//todo put application in error state
		//log.PutInfo("Putting application in error state")
		return errors.New("Failed to pull data from online remote.")
	}

	return nil
}

/**
always push to offline remote, only servers can push to online remote
*/
func gitPush() error {
	errorCount := 0

	cmd := exec.Command("git", "-C", absDbPath, "push", "online", "master")
	//log.PutInfo(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		//log.PutError("Failed to push data to online remotes.")
		//log.PutError(string(out))
		println(string(out))
		errorCount++
	}

	if errorCount > 0 {
		//todo put application in error state
		//log.PutInfo("Putting application in error state")
		return errors.New("Failed to push data to online remotes.")
	}

	return nil
}

func gitCheckout() {
	//remove any changes that might have been introduced directly
	cmd := exec.Command("git", "-C", absDbPath, "checkout", ".")
	//log.PutInfo(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		panic(string(out))
	}
}

func gitCommit(msg string, user *DbUser) {
	//log.PutInfo("**** COMMIT BEGIN ****")
	errorCount := 0

	cmd := exec.Command("git", "-C", absDbPath, "config", "user.email", user.Email)
	//log.PutInfo(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		//log.PutError(string(out))
		println(string(out))
	}

	cmd = exec.Command("git", "-C", absDbPath, "config", "user.name", user.Name)
	//log.PutInfo(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		//log.PutError(string(out))
		println(string(out))
	}

	cmd = exec.Command("git", "-C", absDbPath, "add", ".")
	//log.PutInfo(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		errorCount++
		//log.PutError(string(out))
		println(string(out))
	}

	cmd = exec.Command("git", "-C", absDbPath, "commit", "-am", msg)
	//log.PutInfo(utils.CmdToString(cmd))
	if out, err := cmd.CombinedOutput(); err != nil {
		errorCount++
		//log.PutError(string(out))
		println(string(out))
	}
	//log.PutInfo("**** COMMIT END ****")
	if errorCount > 0 {
		//log.PutInfo("Putting application in error state")
	}
}

func gitLastCommitTime() (time.Time, error) {
	var t time.Time
	cmd := exec.Command("git", "-C", absDbPath, "log", "-1", "--remotes=online", "--format=%cd", "--date=iso")
	//log.PutInfo(utils.CmdToString(cmd))
	out, err := cmd.CombinedOutput()
	if err != nil {
		//log.PutError("gitLastCommit Failed")
		return t, err
	}

	timeString := string(out)
	return time.Parse("2006-01-02 15:04:05", timeString[:19])
}
