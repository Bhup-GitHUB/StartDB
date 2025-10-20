package cli

import "fmt"

// Color constants for terminal output
const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorBlue   = "\033[34m"
	ColorPurple = "\033[35m"
	ColorCyan   = "\033[36m"
	ColorWhite  = "\033[37m"
	ColorGray   = "\033[90m"
	
	// Bright colors
	ColorBrightRed    = "\033[91m"
	ColorBrightGreen  = "\033[92m"
	ColorBrightYellow = "\033[93m"
	ColorBrightBlue   = "\033[94m"
	ColorBrightPurple = "\033[95m"
	ColorBrightCyan   = "\033[96m"
	ColorBrightWhite  = "\033[97m"
)

// Color functions for different types of output
func ColorSuccess(format string, args ...interface{}) string {
	return fmt.Sprintf("%s%s%s", ColorGreen, fmt.Sprintf(format, args...), ColorReset)
}

func ColorError(format string, args ...interface{}) string {
	return fmt.Sprintf("%s%s%s", ColorRed, fmt.Sprintf(format, args...), ColorReset)
}

func ColorWarning(format string, args ...interface{}) string {
	return fmt.Sprintf("%s%s%s", ColorYellow, fmt.Sprintf(format, args...), ColorReset)
}

func ColorInfo(format string, args ...interface{}) string {
	return fmt.Sprintf("%s%s%s", ColorCyan, fmt.Sprintf(format, args...), ColorReset)
}

func ColorHeader(format string, args ...interface{}) string {
	return fmt.Sprintf("%s%s%s", ColorBrightBlue, fmt.Sprintf(format, args...), ColorReset)
}

func ColorPrompt(format string, args ...interface{}) string {
	return fmt.Sprintf("%s%s%s", ColorBrightGreen, fmt.Sprintf(format, args...), ColorReset)
}

func ColorData(format string, args ...interface{}) string {
	return fmt.Sprintf("%s%s%s", ColorWhite, fmt.Sprintf(format, args...), ColorReset)
}

func ColorMuted(format string, args ...interface{}) string {
	return fmt.Sprintf("%s%s%s", ColorGray, fmt.Sprintf(format, args...), ColorReset)
}

func ColorSQL(format string, args ...interface{}) string {
	return fmt.Sprintf("%s%s%s", ColorBrightPurple, fmt.Sprintf(format, args...), ColorReset)
}

func ColorTransaction(format string, args ...interface{}) string {
	return fmt.Sprintf("%s%s%s", ColorBrightYellow, fmt.Sprintf(format, args...), ColorReset)
}

// Print functions with colors
func PrintSuccess(format string, args ...interface{}) {
	fmt.Print(ColorSuccess(format, args...))
}

func PrintError(format string, args ...interface{}) {
	fmt.Print(ColorError(format, args...))
}

func PrintWarning(format string, args ...interface{}) {
	fmt.Print(ColorWarning(format, args...))
}

func PrintInfo(format string, args ...interface{}) {
	fmt.Print(ColorInfo(format, args...))
}

func PrintHeader(format string, args ...interface{}) {
	fmt.Print(ColorHeader(format, args...))
}

func PrintPrompt(format string, args ...interface{}) {
	fmt.Print(ColorPrompt(format, args...))
}

func PrintData(format string, args ...interface{}) {
	fmt.Print(ColorData(format, args...))
}

func PrintMuted(format string, args ...interface{}) {
	fmt.Print(ColorMuted(format, args...))
}

func PrintSQL(format string, args ...interface{}) {
	fmt.Print(ColorSQL(format, args...))
}

func PrintTransaction(format string, args ...interface{}) {
	fmt.Print(ColorTransaction(format, args...))
}
