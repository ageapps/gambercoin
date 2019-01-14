package logger

import (
	"fmt"
	"log"
	"os"
)

// DebugLevel for logger
type DebugLevel int

const (
	// Verbose mode
	Verbose DebugLevel = 2
	// Info mode
	Info DebugLevel = 1
	// Warning mode
	Warning DebugLevel = 0
)

// Logger struct
type Logger struct {
	address string
	name    string
	level   DebugLevel
	log     *log.Logger
}

var instance = Logger{
	address: "",
	name:    "",
	level:   -1,
	log:     log.New(os.Stdout, "", log.Ltime),
}

// LogRumor func
func LogRumor(origin, from, id, text string) {
	Logw("RUMOR origin %v from %v ID %v contents %v\n", origin, from, id, text)
}

// LogStatus func
func LogStatus(wanted, from string) {
	Logw("STATUS from %v %v\n", from, wanted)
}

// LogSimple func
func LogSimple(origin, relay, content string) {
	Logw("SIMPLE MESSAGE origin %v from %v contents %v\n", origin, relay, content)
}

// LogPeers func
func LogPeers(peers string) {
	Logw("PEERS %v\n", peers)
}

// LogInSync func
func LogInSync(peer string) {
	Logw("IN SYNC WITH %v\n", peer)
}

// LogClient func
func LogClient(text string) {
	Logw("CLIENT MESSAGE %v\n", text)
}

// LogCoin func
func LogCoin(address string) {
	Logw("FLIPPED COIN sending rumor to %v\n", address)
}

// LogMonguer func
func LogMonguer(address string) {
	Logw("MONGERING with %v \n", address)
}

// LogDSDV func
func LogDSDV(origin, address string) {
	Logw("DSDV %v %v\n", origin, address)
}

// LogPrivate func
func LogPrivate(origin, hops, text string) {
	Logw("PRIVATE origin %v hop-limit %v contents %v \n", origin, hops, text)
}

// LogFoundBlock func
func LogFoundBlock(hash string) {
	Logw("FOUND-BLOCK %v\n", hash)
}

// LogForkShort func
func LogForkShort(hash string) {
	Logw("FORK-SHORTER %v\n", hash)
}

// LogForkLong func
func LogForkLong(blocks int) {
	Logw("FORK-LONGER rewind %v blocks\n", blocks)
}

// CreateLogger func
func CreateLogger(name, address string, level DebugLevel) {
	instance.name = name
	instance.address = address
	instance.level = level

	instance.log = log.New(os.Stdout, "["+name+":"+address+"]: ", log.Ltime)
	Logi("*******  Logger Created  ******\n")
	Logi("NAME: %v\n", name)
	Logi("ADRESS: %v\n", address)
	Logi("*******  **************  ******\n")

}

// Logi func
func Logi(format string, v ...interface{}) {
	print(Info, fmt.Sprintf(format, v...))
}

// Logv func
func Logv(format string, v ...interface{}) {
	print(Verbose, fmt.Sprintf(format, v...))
}

// Logw func
func Logw(format string, v ...interface{}) {
	print(Warning, fmt.Sprintf(format, v...))
}

// Logf func
func Logf(format string, v ...interface{}) {
	print(Info, fmt.Sprintf(format, v...))
}

// Log func
func Log(level DebugLevel, format string, v ...interface{}) {
	print(level, fmt.Sprintf(format, v...))
}

// Log func
func print(level DebugLevel, text string) {
	if level <= instance.level {
		instance.log.Print(text)
	}
}
