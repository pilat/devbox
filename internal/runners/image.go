package runners

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pilat/devbox/internal/config"
	"github.com/pilat/devbox/internal/docker"
	"github.com/pilat/devbox/internal/pkg/utils"
)

type imageRunner struct {
	cli docker.Service
	log *slog.Logger

	cfg       *config.Config
	container *config.ContainerConfig // if nil, then it's an external image
	dependsOn []string
}

var _ Runner = (*imageRunner)(nil)

func NewImageRunner(cli docker.Service, log *slog.Logger, cfg *config.Config, container *config.ContainerConfig, dependsOn []string) Runner {
	return &imageRunner{
		cli: cli,
		log: log,

		cfg:       cfg,
		container: container,
		dependsOn: dependsOn,
	}
}

func (s *imageRunner) Ref() string {
	return s.container.Image
}

func (s *imageRunner) DependsOn() []string {
	return s.dependsOn
}

func (s *imageRunner) Start(ctx context.Context) error {
	err := s.start(ctx)
	if err != nil {
		s.log.Error("Failed to build image", "error", err)
		return err
	}

	return nil
}

func (s *imageRunner) Stop(ctx context.Context) error {
	return nil
}

func (s *imageRunner) Destroy(ctx context.Context) error {
	// TODO: implement image removal
	return nil
}

func (s *imageRunner) start(ctx context.Context) error {
	s.log.Info("Creating Dockerfile", "image", s.container.Image)

	// Create tar archive containing Dockerfile and context directory
	contextBuffer := new(bytes.Buffer)
	tarWriter := tar.NewWriter(contextBuffer)
	defer tarWriter.Close()

	s.log.Debug("Getting home dir")
	homedir, err := utils.GetHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home dir: %v", err)
	}

	projectPath := fmt.Sprintf("%s/.devbox/%s", homedir, s.cfg.Name)
	dockerfilePath := filepath.Join(projectPath, s.container.Dockerfile)

	s.log.Debug("Reading Dockerfile", "dockerfile", dockerfilePath)
	dockerfile, err := os.ReadFile(dockerfilePath)
	if err != nil {
		return fmt.Errorf("failed to read Dockerfile: %v", err)
	}

	s.log.Debug("Writing Dockerfile to tar")
	dockerfileHeader := &tar.Header{
		Name: "Dockerfile",
		Size: int64(len(dockerfile)),
		Mode: 0600,
	}
	err = tarWriter.WriteHeader(dockerfileHeader)
	if err != nil {
		return fmt.Errorf("failed to write Dockerfile header to tar: %v", err)
	}
	_, err = tarWriter.Write([]byte(dockerfile))
	if err != nil {
		return fmt.Errorf("failed to write Dockerfile content to tar: %v", err)
	}

	contextDir := filepath.Join(projectPath, s.container.Context)
	s.log.Debug("Reading context directory", "context", contextDir)
	if _, err := os.Stat(contextDir); os.IsNotExist(err) {
		return fmt.Errorf("context directory not found: %v", err)
	}

	// TODO: implement .dockerignore support

	s.log.Debug("Scanning context directory", "context", contextDir)
	var files []string
	err = filepath.Walk(contextDir, func(file string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}

		// ignore "".git" and "source" directories
		if strings.HasPrefix(file, projectPath+"/.git") || strings.HasPrefix(file, projectPath+"/source") {
			return nil
		}

		files = append(files, file)

		return nil
	})
	if err != nil {
		return fmt.Errorf("failed to create tar archive: %v", err)
	}

	sort.Strings(files)

	s.log.Debug("Writing context directory to tar")
	for _, file := range files {
		fi, err := os.Stat(file)
		if err != nil {
			return err
		}

		header, err := tar.FileInfoHeader(fi, file)
		if err != nil {
			return err
		}

		header.Name = strings.TrimPrefix(file, contextDir+"/")

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if fi.Mode().IsRegular() {
			f, err := os.Open(file)
			if err != nil {
				return err
			}
			defer f.Close()

			if _, err := io.Copy(tarWriter, f); err != nil {
				return err
			}
		}
	}

	// Close the tar archive
	if err := tarWriter.Close(); err != nil {
		return fmt.Errorf("failed to close tar archive: %v", err)
	}

	// Set up context and build options
	buildOptions := docker.ImageBuildOptions{
		Tags:       []string{s.container.Image},
		Dockerfile: "Dockerfile",
		Version:    "2", // important for docker, not for podman
	}

	s.log.Debug("Starting Docker image build", "image", s.container.Image)
	imageBuildResponse, err := s.cli.ImageBuild(ctx, contextBuffer, buildOptions)
	if err != nil {
		return fmt.Errorf("failed to build Docker image: %v", err)
	}
	defer imageBuildResponse.Body.Close()

	s.log.Debug("Waiting for Docker image build to complete", "image", s.container.Image)
	var stdout bytes.Buffer
	for {
		temp := make([]byte, 1024) // Temporary buffer to read one byte
		n, err := imageBuildResponse.Body.Read(temp)
		if err != nil {
			break
		}

		if n > 0 {
			_, _ = stdout.Write(temp[:n]) // Write the read byte to stdout
		}
	}

	s.log.Debug("Docker image build completed.", "image", s.container.Image, "output", stdout.String())
	return nil
}
