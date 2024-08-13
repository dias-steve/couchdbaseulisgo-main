package couchdbUtils

import "log"

func printSuccess(methodName string, msg ...interface{}) {
	log.Println("_________[SUCCESS]_________", methodName, " => ", msg)
}
func printStart(methodName string) {
	log.Println("_________[START]_________", methodName, "_______")
}
func printInfo(methodName string, msg ...interface{}) {
	log.Println("*****[INFO]***** - func : ", methodName, " :: ", msg)
}
func printError(methodName string, msg ...interface{}) {
	log.Println("xxxx[ERROR]xxxx - func : ", methodName, " :: ", msg)
}
func printDebug(methodName string, printDebugStatus bool, msg ...interface{}) {
	if printDebugStatus {
		log.Println("-----[DEBUG]----- - func : ", methodName, " :: ", msg)

	}
}
func printStatus(index int, length int) {
	log.Println(" --- STATUS --- ", index+1, " / ", length)
}
