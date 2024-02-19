package main

import "os"

// gets list of directories in path
func GetDirectoriesInPath(path string) (sites []string) {
	dirs, _ := os.ReadDir(path)
	for _, entry := range dirs {
		if entry.IsDir() && entry.Name()[0] != '.' {
			sites = append(sites, entry.Name())
		}
	}
	return
}
