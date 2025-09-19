package utils

import (
	"fmt"
	"math/rand"
	"net"
	"regexp"
	"strconv"
	"strings"
)

func CreateRandConref() string {
	return fmt.Sprintf("%08x-%04x-%04x-%08x",
		rand.Intn(1<<32), // Generates a random 32-bit integer for the first 8 characters
		rand.Intn(1<<16), // Generates a random 16-bit integer for the next 4 characters
		rand.Intn(1<<16), // Generates a random 16-bit integer for the next 4 characters
		rand.Intn(1<<32)) // Generates a random 32-bit integer for the last 8 characters
}

func CreateRand4Digits() string {
	return fmt.Sprintf("%04d", rand.Intn(1e4))
}

func ExtractIpAddr(str string) string {
	ipPattern := regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
	match := ipPattern.FindString(str)
	if match == "" || net.ParseIP(match) == nil {
		return ""
	}
	return match
}

func ExtractPort(str string) string {
	portPattern := regexp.MustCompile(`\d+`)
	match := portPattern.FindString(str)
	if match == "" {
		return ""
	}
	portInt, err := strconv.Atoi(match)
	if err != nil || portInt <= 0 || portInt >= (1<<16) {
		return ""
	}
	return match
}

func RemoveComments(str string) string {
	var result strings.Builder

	inBlockComment := false
	for i := 0; i < len(str); i++ {
		if !inBlockComment && strings.HasPrefix(str[i:], "/*") {
			inBlockComment = true
			i++ // Skip the second character of /*
		} else if inBlockComment && strings.HasPrefix(str[i:], "*/") {
			inBlockComment = false
			i++ // Skip the second character of */
		} else if !inBlockComment && strings.HasPrefix(str[i:], "//") {
			for i < len(str) && str[i] != '\n' {
				i++
			}
		} else if !inBlockComment {
			result.WriteByte(str[i])
		}
	}

	return result.String()
}

func IsRTPPacket(buf []byte) bool {
	if len(buf) < 12 {
		return false
	}
	version := buf[0] >> 6
	return version == 2
}

func CreateListenServer(sListenAddr string) (*net.UDPConn, error) {
	listenAddr, err := net.ResolveUDPAddr("udp", sListenAddr)
	if err != nil {
		return nil, err
	}
	conn, err := net.ListenUDP("udp", listenAddr)
	if err != nil {
		return nil, err
	}
	return conn, err

}
