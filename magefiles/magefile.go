package main

import (
	"fmt"
	"strings"

	"github.com/magefile/mage/mg"
	"github.com/magefile/mage/sh"
)

// var Default = All

func All() {
	mg.Deps(BuildFSEventImageDocker)
}

func getGitShortHash() (string, error) {
	return sh.Output("git", "rev-parse", "--short", "HEAD")
}

func gitGetLatestTag() (string, error) {
	return sh.Output("git", "describe", "--tags", "--abbrev=0")
}

func getAllTags() (string, error) {
	// collect the tags
	tags := []string{"latest"}

	t, err := getGitShortHash()
	if err != nil {
		return "", err
	}
	if t != "" {
		tags = append(tags, t)
	}

	t, err = gitGetLatestTag()
	_ = err
	// if err != nil {
	// 	return "", err
	// }
	if t != "" {
		tags = append(tags, t)
	}

	return strings.Join(tags, ","), nil
}

// BuildFSEventImageBD builds a container image and pushes it to cir-cn.chp.belastingdienst.nl/zandp06
func BuildFSEventImageDocker() error {
	env := map[string]string{
		"KO_DOCKER_REPO":      "docker.io/peterzandbergen",
		"KO_DEFAULTBASEIMAGE": "cgr.dev/chainguard/static",
	}

	imageTags, err := getAllTags()
	if err != nil {
		return err
	}

	if err := sh.RunWith(env, "ko", "build", "--tags", imageTags, "./cmd/fsevent"); err != nil {
		return fmt.Errorf("ko build failed: %w", err)
	}
	return nil
}

// BuildFSEventImageLocal builds a container image and pushes it to the local docker daemon
func BuildFSEventImageLocal() error {
	imageTags, err := getAllTags()
	if err != nil {
		return err
	}

	err = sh.Run("ko", "build", "--local", "--tags", imageTags, "./cmd/fsevent")
	if err != nil {
		return fmt.Errorf("ko build failed: %w", err)
	}
	return nil
}

// BuildFSEventServerLocal build an exe in bin
func BuildFSEventServerLocal() error {
	err := sh.Run("go", "build", "-o", "./bin/fsevent", "github.com/myhops/FSEventserver/cmd/FSEventserver")
	if err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}
	return nil
}

func RunFSEventServer() error {
	err := sh.Run("go", "run", "./cmd/FSEventserver/", "github.com/myhops/FSEventserver/cmd/FSEventserver")
	if err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}
	return nil
}

func RunGodoc() error {
	err := sh.RunV("godoc", "-v", "-http", "localhost:6060", "-index")
	if err != nil {
		return fmt.Errorf("go build failed: %w", err)
	}
	return nil
}
