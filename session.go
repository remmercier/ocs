package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
)

type TimeInfo struct {
	Created interface{} `json:"created"`
	Updated interface{} `json:"updated"`
}

type Summary struct {
	Additions int `json:"additions"`
	Deletions int `json:"deletions"`
	Files     int `json:"files"`
}

type Session struct {
	ID        string   `json:"id"`
	Version   string   `json:"version"`
	ProjectID string   `json:"projectID"`
	Directory string   `json:"directory"`
	Title     string   `json:"title"`
	Time      TimeInfo `json:"time"`
	Summary   Summary  `json:"summary"`
}

func parseSession(filePath string) (Session, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return Session{}, err
	}
	var s Session
	err = json.Unmarshal(data, &s)
	return s, err
}

func scanSessions(dir string) ([]Session, error) {
	var sessions []Session
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".json" {
			s, err := parseSession(path)
			if err != nil {
				log.Printf("Error parsing session file %s: %v", path, err)
				return nil // continue
			}
			sessions = append(sessions, s)
		}
		return nil
	})
	return sessions, err
}
