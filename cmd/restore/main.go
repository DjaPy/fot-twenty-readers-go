package main

import (
	"bufio"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type options struct {
	BackupFile string
	DBPath     string
	Force      bool
}

func main() {
	opts := options{}
	flag.StringVar(&opts.BackupFile, "backup", "", "Path to backup file (required)")
	flag.StringVar(&opts.DBPath, "db", "for-twenty-readers.db", "Path to database file")
	flag.BoolVar(&opts.Force, "force", false, "Skip confirmation prompt")
	flag.Parse()

	if opts.BackupFile == "" {
		listBackups()
		fmt.Println("\nUsage: restore -backup <file> [-db <path>] [-force]")
		os.Exit(1)
	}

	if err := runRestore(opts); err != nil {
		slog.Error("restore failed", "error", err)
		os.Exit(1)
	}
}

func runRestore(opts options) error {
	slog.Info("Starting restore", "backup", opts.BackupFile, "db", opts.DBPath)

	if _, err := os.Stat(opts.BackupFile); os.IsNotExist(err) {
		return fmt.Errorf("backup file not found: %s", opts.BackupFile)
	}

	if !opts.Force {
		fmt.Printf("\n⚠️  WARNING: This will replace the current database!\n")
		fmt.Printf("Backup file: %s\n", opts.BackupFile)
		fmt.Printf("Target DB:   %s\n\n", opts.DBPath)
		fmt.Print("Continue? (yes/no): ")

		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		answer = strings.TrimSpace(strings.ToLower(answer))

		if answer != "yes" && answer != "y" {
			fmt.Println("Restore canceled")
			return nil
		}
	}

	if _, err := os.Stat(opts.DBPath); err == nil {
		timestamp := time.Now().Format("20060102_150405")
		safetyBackup := fmt.Sprintf("%s.before_restore_%s", opts.DBPath, timestamp)

		if err := copyFile(opts.DBPath, safetyBackup); err != nil {
			return fmt.Errorf("failed to create safety backup: %w", err)
		}
		slog.Info("created safety backup", "path", safetyBackup)
	}

	if err := copyFile(opts.BackupFile, opts.DBPath); err != nil {
		return fmt.Errorf("failed to restore database: %w", err)
	}

	stat, _ := os.Stat(opts.DBPath)
	slog.Info("Restore completed successfully",
		"path", opts.DBPath,
		"size_mb", stat.Size()/1024/1024,
	)

	fmt.Println("\n✅ Database restored successfully!")
	fmt.Println("Please restart the application for changes to take effect")

	return nil
}

func copyFile(src, dst string) error {
	cleanSrc := filepath.Clean(src)
	cleanDst := filepath.Clean(dst)

	sourceFile, err := os.Open(cleanSrc)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(cleanDst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer destFile.Close()

	_, err = destFile.ReadFrom(sourceFile)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}
	return nil
}

func listBackups() {
	fmt.Println("\nAvailable backups:")

	backupDirs := []string{"./backups", "./data/backups", "/app/backups"}
	found := false

	for _, dir := range backupDirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			continue
		}

		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}

			matched, _ := filepath.Match("for-twenty-readers_*.db", entry.Name())
			if matched {
				info, _ := entry.Info()
				path := filepath.Join(dir, entry.Name())
				fmt.Printf("  %s (%d MB, modified: %s)\n",
					path,
					info.Size()/1024/1024,
					info.ModTime().Format("2006-01-02 15:04:05"),
				)
				found = true
			}
		}
	}

	if !found {
		fmt.Println("  No backups found")
	}
}
