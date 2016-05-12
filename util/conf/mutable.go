package conf

import (
	"bytes"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/gochenzl/chess/util/log"
)

const (
	ConfigTypeJson = 1
	ConfigTypeCsv  = 2
	ConfigTypeIni  = 3
)

type mutableConfig struct {
	mutex       sync.RWMutex
	lastModTime time.Time
	confFile    string
	configType  int
	configData  interface{}
}

var mutex sync.Mutex
var once sync.Once
var mutableConfigs []*mutableConfig

func NewMutableConfig(confFile string, configType int, configData interface{}) *sync.RWMutex {
	if configType != ConfigTypeCsv &&
		configType != ConfigTypeIni &&
		configType != ConfigTypeJson {
		return nil
	}

	var mc mutableConfig
	mc.confFile = confFile
	mc.configType = configType
	mc.configData = configData

	if !mc.load() {
		return nil
	}

	mutex.Lock()
	mutableConfigs = append(mutableConfigs, &mc)
	mutex.Unlock()

	once.Do(func() { go refreshConfigs() })

	return &(mc.mutex)
}

func refreshConfigs() {
	for {
		time.Sleep(time.Second * 10)

		mutex.Lock()
		for i := 0; i < len(mutableConfigs); i++ {
			mutableConfigs[i].load()
		}
		mutex.Unlock()
	}
}

func getModTime(confFile string) (time.Time, error) {
	f, err := os.Open(confFile)
	if err != nil {
		return time.Now(), err
	}

	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return time.Now(), err
	}

	return fi.ModTime(), nil
}

func (config *mutableConfig) load() bool {
	lastModTime, err := getModTime(config.confFile)
	if err != nil {
		log.Error("load %s fail: %s", config.confFile, err.Error())
		return false
	}

	if lastModTime == config.lastModTime {
		return true
	}

	config.lastModTime = lastModTime

	log.Info("load %s", config.confFile)

	dat, err := ioutil.ReadFile(config.confFile)
	if err != nil {
		log.Error("load %s fail: %s", config.confFile, err.Error())
		return false
	}

	reader := bytes.NewReader(dat)

	config.mutex.Lock()
	defer config.mutex.Unlock()

	if config.configType == ConfigTypeJson {
		if err := LoadJson(reader, config.configData); err != nil {
			log.Error("load %s fail: %s", config.confFile, err.Error())
			return false
		}
	} else if config.configType == ConfigTypeCsv {
		if err := LoadCsv(reader, config.configData, true); err != nil {
			log.Error("load %s fail: %s", config.confFile, err.Error())
			return false
		}
	} else if config.configType == ConfigTypeIni {
		if err := LoadIni(reader, config.configData); err != nil {
			log.Error("load %s fail: %s", config.confFile, err.Error())
			return false
		}
	} else {
		return false
	}

	return true
}
