package runtimeutil

import "runtime/debug"

type BuildInfoSetting string

const (
	// name of the version control system
	BuildInfoVersionControl BuildInfoSetting = "vcs"

	// commit id
	BuildInfoRevision BuildInfoSetting = "vcs.revision"

	// commit time
	BuildInfoTime BuildInfoSetting = "vcs.time"
)

type BuildInfo struct {
	*debug.BuildInfo
}

func NewBuildInfo() *BuildInfo {
	info, ok := debug.ReadBuildInfo()

	if !ok {
		info = &debug.BuildInfo{
			Settings: []debug.BuildSetting{},
			Deps:     []*debug.Module{},
		}
	}

	return &BuildInfo{BuildInfo: info}
}

func (b *BuildInfo) GetStringSetting(key BuildInfoSetting) (string, bool) {
	for _, s := range b.Settings {
		if s.Key == string(key) {
			return s.Value, true
		}
	}

	return "", false
}
