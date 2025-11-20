package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"time"
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

type Sessions []Session

type Session struct {
	ID        string   `json:"id"`
	Version   string   `json:"version"`
	ProjectID string   `json:"projectID"`
	Directory string   `json:"directory"`
	Title     string   `json:"title"`
	Time      TimeInfo `json:"time"`
	Summary   Summary  `json:"summary"`
}

func getTime(created interface{}) time.Time {
	switch v := created.(type) {
	case float64:
		return time.UnixMilli(int64(v))
	case string:
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return time.UnixMilli(i)
		}
	}
	return time.Time{}
}

func NewSession(filePath string) (Session, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return Session{}, err
	}
	var s Session
	err = json.Unmarshal(data, &s)
	return s, err
}

func (s *Sessions) Scan(dir string) error {
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".json" {
			session, err := NewSession(path)
			if err != nil {
				log.Printf("Error parsing session file %s: %v", path, err)
				return nil // continue
			}
			*s = append(*s, session)
		}
		return nil
	})
	return err
}

func (s Sessions) Sort() {
	sort.Slice(s, func(i, j int) bool {
		ti := getTime(s[i].Time.Created)
		tj := getTime(s[j].Time.Created)
		return ti.After(tj)
	})
}
