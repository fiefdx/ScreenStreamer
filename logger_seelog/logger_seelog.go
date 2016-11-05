// Created on 2015-07-31
// summary: logger_seelog
// author: YangHaitao

package logger_seelog

import (
    "os"
    "fmt"
    "errors"
    "strings"
    "path/filepath"
    "seelog"
)

// <rollingfile type="size" filename="FILE_PATH" maxsize="MAX_SIZE" maxrolls="MAX_ROLLS"/>
// <rollingfile type="date" filename="FILE_PATH" datepattern="02.01.2006" maxrolls="MAX_ROLLS"/>
// type="asynctimer" asyncinterval="5000000"
// <console/>

var ConfigTemp string = `
<seelog type="sync" minlevel="LOG_LEVEL">
    <outputs formatid="main">
        CONSOLE
        ROLLING_FILE
    </outputs>
    <formats>
        <format id="main" format="%Date(2006-01-02 15:04:05.999) [%Level] [%RelFile %Func] %Msg%n"/>
    </formats>
</seelog>
`
var Log seelog.LoggerInterface
var Logger *seelog.LoggerInterface
var Loggers map[string]*seelog.LoggerInterface
var Level map[string]string

func init() {
    Level = make(map[string]string)
    Level["TRACE"] = "trace"
    Level["DEBUG"] = "debug"
    Level["INFO"] = "info"
    Level["WARN"] = "warn"
    Level["ERROR"] = "error"
    Level["CRITICAL"] = "critical"
    DisableLog()
}

func DisableLog() {
    Logger = &seelog.Disabled
}

func UseLogger(newLogger *seelog.LoggerInterface) {
    Logger = newLogger
}

func setLevel(config, level string) string {
    newConfig := strings.Replace(config, "LOG_LEVEL", Level[level], 1)
    return newConfig
}

func enableConsole(config string, b bool) string {
    var newConfig string
    if b {
        newConfig = strings.Replace(config, "CONSOLE", "<console/>", 1)
    } else {
        newConfig = strings.Replace(config, "CONSOLE", "", 1)
    }
    return newConfig
}

func enableRolling(config, rolling_type, file_path, max_size, max_rolls string) string {
    var newConfig string
    if rolling_type == "size" {
        newConfig = strings.Replace(config, "ROLLING_FILE", fmt.Sprintf("<rollingfile type=\"size\" filename=\"%s\" maxsize=\"%s\" maxrolls=\"%s\"/>", file_path, max_size, max_rolls), 1)
    } else if rolling_type == "date" {
        newConfig = strings.Replace(config, "ROLLING_FILE", fmt.Sprintf("<rollingfile type=\"date\" filename=\"%s\" datepattern=\"02.01.2006\" maxrolls=\"%s\"/>", file_path, max_rolls), 1)
    } else {
        newConfig = strings.Replace(config, "ROLLING_FILE", "", 1)
    }
    return newConfig
}

func NewLogger(log_name, log_path, log_file, log_level, rolling_type, max_size, max_rolls string, console bool) (*seelog.LoggerInterface, error) {
    if Loggers == nil {
        Loggers = make(map[string]*seelog.LoggerInterface)
        fmt.Printf("make Loggers success\n")
    }

    if logger, ok := Loggers[log_name]; ok {
        // seelog.UseLogger(*logger)
        Log = *logger
        Log.Info(fmt.Sprintf("Logger(\"%s\") have been exist, so return it!", log_name))
        return logger, nil
    }

    if _, err := os.Stat(log_path); os.IsNotExist(err) {
        fmt.Printf("Init logger: log path: (%s) does not exist!", log_path)
        return nil, err
    }

    file_path := filepath.Join(log_path, log_file)
    Config := setLevel(ConfigTemp, log_level)
    if console {
        Config = enableConsole(Config, console)
    }
    Config = enableRolling(Config, rolling_type, file_path, max_size, max_rolls)

    logger, _ := seelog.LoggerFromConfigAsBytes([]byte(Config))
    logger.Info("Init logger success")
    fmt.Printf("Init logger success\n")
    // seelog.ReplaceLogger(logger)
    Loggers[log_name] = &logger
    return &logger, nil
}

func GetLogger(log_name string) (*seelog.LoggerInterface, error) {
    if logger, ok := Loggers[log_name]; ok {
        // seelog.UseLogger(*logger)
        (*logger).Info(fmt.Sprintf("Logger(\"%s\") have been exist, so return it!", log_name))
        return logger, nil
    }
    return nil, errors.New(fmt.Sprintf("Logger(\"%s\") does not exist!", log_name))
}

func SetLogger(log_name string) error {
    if logger, ok := Loggers[log_name]; ok {
        seelog.UseLogger(*logger)
        Log.Info(fmt.Sprintf("Set Logger(\"%s\") success", log_name))
        return nil
    }
    return errors.New(fmt.Sprintf("Logger(\"%s\") does not exist!", log_name))
}

func CloseAll() {
    for name, logger := range Loggers {
        // seelog.UseLogger(*logger)
        (*logger).Info(fmt.Sprintf("Close logger: %s\n", name))
        defer (*logger).Flush()
    }
}

