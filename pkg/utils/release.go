//go:build !debug
// +build !debug

package utils

import "log"

func Debug(_ string, _ ...interface{}) {}
func Log(fmt string, args ...interface{}) {
	log.Printf(fmt, args...)
}
