//go:build debug
// +build debug

package utils

import "log"

func Debug(fmt string, args ...interface{}) {
	log.Printf("[DEBUG] " + fmt, args...)
}

func Log(fmt string, args ...interface{}) {
	log.Printf(fmt, args...)
}