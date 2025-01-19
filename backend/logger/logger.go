package logger

import (
	"log"
	"os"
)

// Print logs the provided arguments using the default logger.
// It accepts a variadic number of arguments of any type and prints them
// to the standard logger output.
//
// Example usage:
//     logger.Print("This is a log message")
//     logger.Print("Multiple", "arguments", 123, true)
//
// Parameters:
//     v ...any: A variadic number of arguments to be logged.
func Print(v ...any) {
	log.Println(v...)
}

// Printf formats according to a format specifier and writes to the log.
// It accepts a format string and a variadic number of arguments.
// The format string follows the same rules as fmt.Printf.
func Printf(format string, v ...any) {
	log.Printf(format, v...)
}
// Debug logs the provided arguments if the application is in debug mode.
// It accepts a variadic number of arguments of any type.
//
// Usage:
//     Debug("This is a debug message")
//     Debug("Value of x:", x)
//
// The function will only log the message if the isDebugMode function returns true.
func Debug(v ...any) {
	if isDebugMode() {
		log.Println(v...)
	}
}

// Debugf logs a formatted debug message if the application is in debug mode.
// The message is formatted according to the specified format and arguments.
//
// Parameters:
//   - format: A format string as described in the fmt package.
//   - v: A variadic list of arguments to be formatted according to the format string.
func Debugf(format string, v ...any) {
	if isDebugMode() {
		log.Printf(format, v...)
	}
}

// isDebugMode checks if the application is running in debug mode.
// It returns true if the environment variable "DEBUG" is set to "true",
// otherwise, it returns false.
func isDebugMode() bool {
	return os.Getenv("DEBUG") == "true"
}
