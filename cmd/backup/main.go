package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/asdine/storm/v3"
	bolt "go.etcd.io/bbolt"
)

type options struct {
	DBPath        string
	BackupDir     string
	RetentionDays int
}

func main() {
	opts := options{}
	flag.StringVar(&opts.DBPath, "db", "for-twenty-readers.db", "Path to database file")
	flag.StringVar(&opts.BackupDir, "backup-dir", "./backups", "Directory for backups")
	flag.IntVar(&opts.RetentionDays, "retention", 30, "Days to keep backups")
	flag.Parse()

	if err := runBackup(opts); err != nil {
		slog.Error("backup failed", "error", err)
		os.Exit(1)
	}
}

func runBackup(opts options) error {
	slog.Info("Starting backup", "db", opts.DBPath, "backup_dir", opts.BackupDir)

	if _, err := os.Stat(opts.DBPath); os.IsNotExist(err) {
		return fmt.Errorf("database file not found: %s", opts.DBPath)
	}

	if err := os.MkdirAll(opts.BackupDir, 0o750); err != nil {
		return fmt.Errorf("failed to create backup directory: %w", err)
	}

	db, err := storm.Open(opts.DBPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer db.Close()

	timestamp := time.Now().Format("20060102_150405")
	backupPath := filepath.Join(opts.BackupDir, fmt.Sprintf("for-twenty-readers_%s.db", timestamp))

	backupFile, err := os.Create(filepath.Clean(backupPath))
	if err != nil {
		return fmt.Errorf("failed to create backup file: %w", err)
	}
	defer func() { _ = backupFile.Close() }()

	err = db.Bolt.View(func(tx *bolt.Tx) error {
		_, errWrite := tx.WriteTo(backupFile)
		if errWrite != nil {
			return fmt.Errorf("failed to write backup file: %w", errWrite)
		}
		return nil
	})
	if err != nil {
		errRemove := os.Remove(backupPath)
		if errRemove != nil {
			return fmt.Errorf("failed to remove file path %s: %w", backupPath, errRemove)
		}
		return fmt.Errorf("failed to write backup: %w", err)
	}

	stat, _ := backupFile.Stat()
	slog.Info("Backup created successfully",
		"path", backupPath,
		"size_mb", stat.Size()/1024/1024,
	)

	if errClOld := cleanOldBackups(opts.BackupDir, opts.RetentionDays); errClOld != nil {
		slog.Warn("failed to clean old backups", "error", errClOld)
	}

	return nil
}

func cleanOldBackups(backupDir string, retentionDays int) error {
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)

	entries, err := os.ReadDir(backupDir)
	if err != nil {
		return fmt.Errorf("failed to read backup directory: %w", err)
	}

	deleted := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		matched, err := filepath.Match("for-twenty-readers_*.db", entry.Name())
		if err != nil || !matched {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().Before(cutoffTime) {
			path := filepath.Join(backupDir, entry.Name())
			if err := os.Remove(path); err != nil {
				slog.Warn("failed to delete old backup", "path", path, "error", err)
			} else {
				deleted++
				slog.Info("deleted old backup", "path", path)
			}
		}
	}

	slog.Info("cleanup completed", "deleted", deleted)
	return nil
}
