package config

import (
	"bufio"
	"errors"
	"os"
	"strings"

	"github.com/misha-ssh/kernel/configs/envconst"
	"github.com/misha-ssh/kernel/internal/logger"
	"github.com/misha-ssh/kernel/internal/storage"
)

const (
	Filename      = envconst.FilenameConfig
	FileFlagWrite = os.O_WRONLY | os.O_APPEND | os.O_CREATE
	FileFlagRead  = os.O_RDWR

	CharNewLine = "\n"
	EmptyValue  = ""
	Separator   = "="
)

var (
	DirectionApp = storage.GetAppDir()

	ErrWriteDataToOpenFile = errors.New("write data to open file error")
	ErrGetKeyValueData     = errors.New("get value data error")
	ErrKeyOfNonLetters     = errors.New("key of non letters error")
	ErrValueIsInvalid      = errors.New("dont valid value at set data")
	ErrGetOpenFile         = errors.New("get open file error")
)

// Set writes or updates a configuration value:
// - Validates key (letters only) and value
// - Creates new entry or updates existing one
// - Automatically converts keys to uppercase
// Returns error if validation or write fails
func Set(key, value string) error {
	err := validateKey(key)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	err = validateValue(value)
	if err != nil {
		logger.Error(err.Error())
		return err
	}

	if Exists(key) {
		err = rewrite(key, value)
		if err != nil {
			logger.Error(err.Error())
			return err
		}

		return nil
	}

	openConfigFile, err := storage.GetOpenFile(DirectionApp, Filename, FileFlagWrite)
	defer func(openConfigFile *os.File) {
		err = openConfigFile.Close()
	}(openConfigFile)
	if err != nil {
		logger.Error(ErrGetOpenFile.Error())
		return ErrGetOpenFile
	}

	param := strings.ToUpper(key) + Separator + value + CharNewLine

	_, err = openConfigFile.WriteString(param)
	if err != nil {
		logger.Error(ErrWriteDataToOpenFile.Error())
		return ErrWriteDataToOpenFile
	}

	return nil
}

// Get retrieves a configuration value:
// - Case-insensitive key lookup (converts to uppercase)
// - Returns empty string if key not found
// - Validates key before lookup
func Get(key string) string {
	err := validateKey(key)
	if err != nil {
		logger.Error(ErrGetOpenFile.Error())
		return EmptyValue
	}

	openConfigFile, err := storage.GetOpenFile(DirectionApp, Filename, FileFlagRead)
	defer func(openConfigFile *os.File) {
		err = openConfigFile.Close()
	}(openConfigFile)
	if err != nil {
		logger.Error(err.Error())
		return EmptyValue
	}

	sc := bufio.NewScanner(openConfigFile)

	for sc.Scan() {
		data := strings.Split(sc.Text(), Separator)

		if len(data) != 2 {
			logger.Error(ErrGetKeyValueData.Error())
			return EmptyValue
		}

		keyConfig := data[0]
		valueConfig := data[1]

		if keyConfig == strings.ToUpper(key) {
			return valueConfig
		}
	}

	if err = sc.Err(); err != nil {
		logger.Error(err.Error())
		return EmptyValue
	}

	return EmptyValue
}

// Exists checks if a configuration key is present:
// - Case-insensitive check
// - Returns false for invalid keys
// - Uses simple string contains check
func Exists(key string) bool {
	err := validateKey(key)
	if err != nil {
		logger.Error(err.Error())
		return false
	}

	got, err := storage.Get(DirectionApp, Filename)
	if err != nil {
		logger.Error(err.Error())
		return false
	}

	return strings.Contains(got, strings.ToUpper(key))
}

func rewrite(key, value string) error {
	openConfigFile, err := storage.GetOpenFile(DirectionApp, Filename, FileFlagRead)
	defer func(openConfigFile *os.File) {
		err = openConfigFile.Close()
	}(openConfigFile)
	if err != nil {
		logger.Error(ErrGetOpenFile.Error())
		return ErrGetOpenFile
	}

	sc := bufio.NewScanner(openConfigFile)
	var lines []string

	for sc.Scan() {
		line := sc.Text()
		data := strings.Split(line, Separator)

		if len(data) != 2 {
			logger.Error(ErrGetKeyValueData.Error())
			return ErrGetKeyValueData
		}

		keyConfig := data[0]
		UpperKey := strings.ToUpper(key)

		if keyConfig == UpperKey {
			newValue := UpperKey + Separator + value + CharNewLine
			lines = append(lines, newValue)
		} else {
			lines = append(lines, line+CharNewLine)
		}
	}

	if err = sc.Err(); err != nil {
		logger.Error(err.Error())
		return err
	}

	if _, err = openConfigFile.Seek(0, 0); err != nil {
		logger.Error(err.Error())
		return err
	}
	if err = openConfigFile.Truncate(0); err != nil {
		logger.Error(err.Error())
		return err
	}

	writer := bufio.NewWriter(openConfigFile)
	for _, line := range lines {
		if _, err = writer.WriteString(line); err != nil {
			logger.Error(err.Error())
			return err
		}
	}

	return writer.Flush()
}
