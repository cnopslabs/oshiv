package utils

func init() {
	// Control the order in which the utilities initialize
	loggerInit()
	configInit()
}
