package adr

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

// FileRepository implements Repository by reading ADR markdown files from a directory.
type FileRepository struct {
	dir string
}

// NewFileRepository creates a FileRepository rooted at dir.
func NewFileRepository(dir string) *FileRepository {
	return &FileRepository{dir: dir}
}

func (r *FileRepository) List(_ context.Context) ([]ADR, error) {
	entries, err := os.ReadDir(r.dir)
	if err != nil {
		return nil, fmt.Errorf("reading directory %q: %w", r.dir, err)
	}

	var adrs []ADR
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		matches := adrFilePattern.FindStringSubmatch(entry.Name())
		if matches == nil {
			continue
		}
		fileNumber, _ := strconv.Atoi(matches[1])

		content, err := os.ReadFile(filepath.Join(r.dir, entry.Name()))
		if err != nil {
			continue
		}

		meta := ExtractMetadata(string(content))
		record, err := MetadataToADR(meta, fileNumber)
		if err != nil {
			continue
		}

		adrs = append(adrs, record)
	}

	sort.Slice(adrs, func(i, j int) bool {
		return adrs[i].Number < adrs[j].Number
	})

	return adrs, nil
}

func (r *FileRepository) Get(_ context.Context, number int) (*ADR, error) {
	filename, err := FindADRFile(r.dir, number)
	if err != nil {
		return nil, err
	}

	content, err := os.ReadFile(filepath.Join(r.dir, filename))
	if err != nil {
		return nil, fmt.Errorf("reading %q: %w", filename, err)
	}

	meta := ExtractMetadata(string(content))
	record, err := MetadataToADR(meta, number)
	if err != nil {
		return nil, err
	}

	record.Content = string(content)
	return &record, nil
}

func (r *FileRepository) NextNumber(_ context.Context) (int, error) {
	return NextNumber(r.dir)
}

func (r *FileRepository) Save(_ context.Context, _ *ADR) error {
	return fmt.Errorf("FileRepository.Save not implemented")
}

// UpdateStatus changes the status of the ADR with the given number and returns the updated record.
func (r *FileRepository) UpdateStatus(_ context.Context, number int, newStatus string) (*ADR, error) {
	if _, ok := ParseStatus(newStatus); !ok {
		return nil, fmt.Errorf("invalid status %q", newStatus)
	}

	filename, err := FindADRFile(r.dir, number)
	if err != nil {
		return nil, err
	}

	filePath := filepath.Join(r.dir, filename)
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("reading %q: %w", filename, err)
	}

	updated, err := UpdateStatus(string(content), newStatus)
	if err != nil {
		return nil, err
	}

	if err := os.WriteFile(filePath, []byte(updated), 0o644); err != nil {
		return nil, fmt.Errorf("writing %q: %w", filename, err)
	}

	meta := ExtractMetadata(updated)
	record, err := MetadataToADR(meta, number)
	if err != nil {
		return nil, err
	}

	record.Content = updated
	return &record, nil
}
