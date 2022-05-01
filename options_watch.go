/*
 * Copyright © 2019 Hedzr Yeh.
 */

package cmdr

import (
	"bufio"
	"bytes"
	"encoding/json"
	"github.com/BurntSushi/toml"
	"github.com/hedzr/log/dir"
	"gopkg.in/hedzr/errors.v3"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

// CurrentOptions returns the global options instance (rxxtOptions),
// i.e. cmdr Options Store
func CurrentOptions() *Options {
	return currentOptions()
}

func currentOptions() *Options {
	return internalGetWorker().rxxtOptions
}

// GetUsedAlterConfigFile returns the alternative config filename
func GetUsedAlterConfigFile() string {
	return currentOptions().usedAlterConfigFile
}

// GetUsedSecondaryConfigFile returns the secondary config filename
func GetUsedSecondaryConfigFile() string {
	return currentOptions().usedSecondaryConfigFile
}

// GetUsedSecondaryConfigSubDir returns the subdirectory `conf.d` of secondary config files.
func GetUsedSecondaryConfigSubDir() string {
	return currentOptions().usedSecondaryConfigSubDir
}

// GetUsedConfigFile returns the main config filename (generally
// it's `<appname>.yml`)
func GetUsedConfigFile() string {
	return currentOptions().usedConfigFile
}

// GetUsedConfigSubDir returns the sub-directory `conf.d` of config files.
// Note that it be always normalized now.
// Sometimes it might be empty string ("") if `conf.d` have not been found.
func GetUsedConfigSubDir() string {
	return currentOptions().usedConfigSubDir
}

// GetUsingConfigFiles returns all loaded config files, includes
// the main config file and children in sub-directory `conf.d`.
func GetUsingConfigFiles() []string {
	return currentOptions().configFiles
}

// GetWatchingConfigFiles returns all config files being watched.
func GetWatchingConfigFiles() []string {
	return currentOptions().filesWatching
}

// var rwlCfgReload = new(sync.RWMutex)

// AddOnConfigLoadedListener adds an functor on config loaded
// and merged
func AddOnConfigLoadedListener(c ConfigReloaded) {
	opts := currentOptions()
	defer opts.rwlCfgReload.Unlock()
	opts.rwlCfgReload.Lock()

	// rwlCfgReload.RLock()
	if _, ok := opts.onConfigReloadedFunctions[c]; ok {
		// rwlCfgReload.RUnlock()
		return
	}

	// rwlCfgReload.RUnlock()
	// rwlCfgReload.Lock()

	// defer rwlCfgReload.Unlock()

	opts.onConfigReloadedFunctions[c] = true
}

// RemoveOnConfigLoadedListener remove an functor on config
// loaded and merged
func RemoveOnConfigLoadedListener(c ConfigReloaded) {
	w := internalGetWorker()
	opts := w.rxxtOptions
	defer opts.rwlCfgReload.Unlock()
	opts.rwlCfgReload.Lock()
	delete(opts.onConfigReloadedFunctions, c)
}

// SetOnConfigLoadedListener enable/disable an functor on config
// loaded and merged
func SetOnConfigLoadedListener(c ConfigReloaded, enabled bool) {
	w := internalGetWorker()
	opts := w.rxxtOptions
	defer opts.rwlCfgReload.Unlock()
	opts.rwlCfgReload.Lock()
	opts.onConfigReloadedFunctions[c] = enabled
}

// LoadConfigFile loads a yaml config file and merge the settings
// into `rxxtOptions`
// and load files in the `conf.d` child directory too.
func LoadConfigFile(file string) (mainFile, subDir string, err error) {
	return currentOptions().LoadConfigFile(file, mainConfigFiles)
}

type configFileType int

const (
	mainConfigFiles configFileType = iota
	secondaryConfigFiles
	alterConfigFile

	jsonSuffixString = ".json"
	tomlSuffixString = ".toml"
	confSuffixString = ".conf"
	iniSuffixString  = ".ini"
)

// LoadConfigFile loads a yaml config file and merge the settings
// into `rxxtOptions`
// and load files in the `conf.d` child directory too.
func (s *Options) LoadConfigFile(file string, cft configFileType) (mainFile, subDir string, err error) {
	defer func() {
		s.batchMerging = false
		s.mapOrphans()
	}()

	s.rw.Lock()
	s.batchMerging = true
	s.rw.Unlock()

	if !dir.FileExists(file) {
		// log.Warnf("%v NOT EXISTS. PWD=%v", file, GetCurrentDir())
		return // not error, just ignore loading
	}

	if err = s.loadConfigFile(file); err != nil {
		return
	}

	w := internalGetWorker()
	mainFile = file
	dirWatch := path.Dir(mainFile)
	enableWatching := w.watchMainConfigFileToo
	confDFolderName := w.confDFolderName
	var filesWatching []string
	if cft != alterConfigFile {
		dirname := dirWatch
		if cft == mainConfigFiles {
			s.usedConfigFile = mainFile
			_ = os.Setenv("CFG_DIR", dirname)
		} else if cft == secondaryConfigFiles {
			s.usedSecondaryConfigFile = mainFile
			_ = os.Setenv("CFG_DIR_2NDRY", dirname)
		}

		if w.watchMainConfigFileToo {
			filesWatching = append(filesWatching, mainFile)
		}

		subDir = path.Join(dirname, confDFolderName)
		if !dir.FileExists(subDir) {
			subDir = ""
			if len(filesWatching) == 0 {
				return
			}
		}

		subDir, err = filepath.Abs(subDir)
		if err == nil {
			err = filepath.Walk(subDir, s.visit)
			if err == nil {
				if !w.watchMainConfigFileToo {
					dirWatch = subDir
				}
				filesWatching = append(filesWatching, s.configFiles...)
				enableWatching = true
			}
			// don't bring the minor error for sub-dir walking back to main caller
			err = nil
			// log.Fatalf("ERROR: filepath.Walk() returned %v\n", err)
		}

		if cft == mainConfigFiles {
			s.usedConfigSubDir = subDir
		} else if cft == secondaryConfigFiles {
			s.usedSecondaryConfigSubDir = subDir
		}
	} else {
		s.usedAlterConfigFile = mainFile
	}
	err = s.doWatchConfigFile(enableWatching, confDFolderName, dirWatch, filesWatching)
	return
}

func (s *Options) doWatchConfigFile(enableWatching bool, confDFolderName, dirWatch string, filesWatching []string) (err error) { //nolint:staticcheck //likw it
	if internalGetWorker().watchChildConfigFiles {
		var dirname string
		confDFolderName = os.ExpandEnv(".$APPNAME") //nolint:staticcheck //like it
		dirname, err = filepath.Abs(confDFolderName)
		if err == nil && dir.FileExists(dirname) {
			err = filepath.Walk(dirname, s.visit)
			if err == nil {
				filesWatching = append(filesWatching, s.configFiles...)
				enableWatching = true
			}
			// don't bring the minor error for sub-dir walking back to main caller
			err = nil
		}
	}

	if enableWatching {
		s.watchConfigDir(dirWatch, filesWatching)
	}
	flog("the watching config files: %v", s.filesWatching)
	flog("the loaded config files: %v", s.configFiles)
	return
}

// Load a yaml config file and merge the settings into `Options`
func (s *Options) loadConfigFile(file string) (err error) {
	var m map[string]interface{}
	m, err = s.loadConfigFileAsMap(file)
	if err == nil {
		err = s.loopMap("", m)
	}
	// if err != nil {
	//	return
	// }
	return
}

func (s *Options) loadConfigFileAsMap(file string) (m map[string]interface{}, err error) {
	var (
		b  []byte
		mm map[string]map[string]interface{}
	)

	b, _ = dir.ReadFile(file)

	m = make(map[string]interface{})
	switch path.Ext(file) {
	case tomlSuffixString, iniSuffixString, confSuffixString, "toml":
		mm = make(map[string]map[string]interface{})
		err = toml.Unmarshal(b, &mm)
		if err == nil {
			err = s.loopMapMap("", mm)
		}
		if err != nil {
			return
		}
		return

	case jsonSuffixString, "json":
		err = json.Unmarshal(b, &m)
	default:
		err = yaml.Unmarshal(b, &m)
	}
	return
}

func (s *Options) mergeConfigFile(fr io.Reader, src, ext string) (err error) {
	var (
		m   map[string]interface{}
		buf *bytes.Buffer
	)

	buf = new(bytes.Buffer)
	if _, err = buf.ReadFrom(fr); err == nil {
		m = make(map[string]interface{})
		switch ext {
		case tomlSuffixString, iniSuffixString, confSuffixString, "toml":
			err = toml.Unmarshal(buf.Bytes(), &m)
		case jsonSuffixString, "json":
			err = json.Unmarshal(buf.Bytes(), &m)
		default:
			err = yaml.Unmarshal(buf.Bytes(), &m)
		}
	}

	if err == nil {
		err = s.loopMap("", m)
	}

	if err != nil {
		ferr("unsatisfied config file `%s` while importing: %v", src, err)
		return
	}

	return
}

func (s *Options) visit(pathname string, f os.FileInfo, e error) (err error) {
	// fmt.Printf("Visited: %s, e: %v\n", pathname, e)
	flog("    visiting: %v, e: %v", pathname, e)
	err = e
	if f != nil && !f.IsDir() && e == nil {
		// log.Infof("    path: %v, ext: %v", path, filepath.Ext(path))
		ext := filepath.Ext(pathname)
		switch ext {
		case ".yml", ".yaml", jsonSuffixString, tomlSuffixString, iniSuffixString, confSuffixString: // , "yml", "yaml":
			var file *os.File
			file, err = os.Open(pathname)
			// if err != nil {
			// log.Warnf("ERROR: os.Open() returned %v", err)
			// } else {
			if err == nil {
				defer file.Close()
				flog("    visited and merging: %v", file.Name())
				if err = s.mergeConfigFile(bufio.NewReader(file), file.Name(), ext); err != nil {
					err = errors.New("error in merging config file '%s': %v", pathname, err)
					return
				}
				s.configFiles = uniAddStr(s.configFiles, pathname)
			} else {
				err = errors.New("error in merging config file '%s': %v", pathname, err)
			}
		}
	}
	return
}

func (s *Options) reloadConfig() {
	// log.Debugf("\n\nConfig file changed: %s\n", e.String())

	defer s.rwlCfgReload.RUnlock()
	s.rwlCfgReload.RLock()

	for x, ok := range s.onConfigReloadedFunctions {
		if ok {
			x.OnConfigReloaded()
		}
	}
}

func (s *Options) watchConfigDir(configDir string, filesWatching []string) {
	if internalGetWorker().doNotWatchingConfigFiles || GetBoolR("no-watch-conf-dir") {
		return
	}

	if configDir == "" || len(filesWatching) == 0 {
		return
	}

	initWG := &sync.WaitGroup{}
	initWG.Add(1)
	// initExitingChannelForFsWatcher()
	s.filesWatching = filesWatching
	go fsWatcherRoutine(s, configDir, filesWatching, initWG)
	initWG.Wait() // make sure that the go routine above fully ended before returning
	s.SetNx("watching", true)
}

func testCfgSuffix(name string) bool {
	for _, suf := range []string{".yaml", ".yml", jsonSuffixString, tomlSuffixString, iniSuffixString, confSuffixString} {
		if strings.HasSuffix(name, suf) {
			return true
		}
	}
	return false
}

func testArrayContains(s string, container []string) (contained bool) {
	for _, ss := range container {
		if ss == s {
			contained = true
			break
		}
	}
	return
}
