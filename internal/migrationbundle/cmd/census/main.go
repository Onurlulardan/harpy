package main

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

func main() {
	root := "."
	if len(os.Args) == 2 {
		root = os.Args[1]
	}
	release := filepath.Join(root, "core", "internal", "platform", "migration", "release")
	entries, err := os.ReadDir(release)
	if err != nil {
		fail(err)
	}
	sizes := make([]int, 0, len(entries))
	var rows, total int
	unique := map[string][]byte{}
	phases := map[string]int{}
	generations := map[[2]int64]int{}
	sameSHA, diffSHA, samePath := 0, 0, 0
	sets := make([]map[string]struct{}, 0, len(entries))
	counts := make([]int, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		path := filepath.Join(release, entry.Name())
		body, err := os.ReadFile(path)
		if err != nil {
			fail(err)
		}
		sizes = append(sizes, len(body))
		total += len(body)
		var manifest struct {
			Entries []map[string]any `json:"entries"`
		}
		if err = json.Unmarshal(body, &manifest); err != nil {
			fail(err)
		}
		rows += len(manifest.Entries)
		counts = append(counts, len(manifest.Entries))
		set := make(map[string]struct{}, len(manifest.Entries))
		for _, item := range manifest.Entries {
			encoded, err := json.Marshal(item)
			if err != nil {
				fail(err)
			}
			sum := sha256.Sum256(encoded)
			fp := fmt.Sprintf("%x", sum[:])
			set[fp] = struct{}{}
			if _, exists := unique[fp]; exists {
				continue
			}
			unique[fp] = encoded
			phase, _ := item["phase"].(string)
			phases[phase]++
			generations[[2]int64{number(item["target_database_generation"]), number(item["next_minimum_runtime_generation"])}]++
			fresh, adopted := variant(item, "fresh"), variant(item, "v1_adopted")
			if fresh["sha256"] == adopted["sha256"] {
				sameSHA++
			} else {
				diffSHA++
			}
			if fresh["path"] == adopted["path"] {
				samePath++
			}
		}
		sets = append(sets, set)
	}
	overlap, larger := largestOverlap(sets, counts)
	sort.Ints(sizes)
	fmt.Printf("go %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
	fmt.Printf("manifest_files %d manifest_bytes %d\n", len(sizes), total)
	fmt.Printf("size_min %d median %d p95 %d max %d\n", sizes[0], percentile(sizes, 0.5), percentile(sizes, 0.95), sizes[len(sizes)-1])
	sort.Ints(counts)
	fmt.Printf("entries_min %d median %d max %d entry_rows %d unique_entries %d\n", counts[0], counts[len(counts)/2], counts[len(counts)-1], rows, len(unique))
	fmt.Printf("phases %v\n", phases)
	fmt.Printf("generation_1_1 %d generation_pairs %d\n", generations[[2]int64{1, 1}], len(generations))
	fmt.Printf("variant_sha_same %d variant_sha_diff %d variant_path_same %d\n", sameSHA, diffSHA, samePath)
	fmt.Printf("largest_overlap %d/%d\n", overlap, larger)
	fmt.Printf("unique_entry_json_bytes %d\n", uniqueBytes(unique))
	sqlFiles, sqlBytes := walkSQL(root, false)
	migFiles, migBytes := walkSQL(root, true)
	fmt.Printf("sql_files %d sql_bytes %d\n", sqlFiles, sqlBytes)
	fmt.Printf("migrations_sql %d migrations_sql_bytes %d\n", migFiles, migBytes)
}

func variant(entry map[string]any, name string) map[string]any {
	variants, _ := entry["variants"].(map[string]any)
	value, _ := variants[name].(map[string]any)
	return value
}

func number(value any) int64 {
	switch typed := value.(type) {
	case float64:
		return int64(typed)
	default:
		return 0
	}
}

func percentile(sorted []int, p float64) int {
	position := float64(len(sorted)-1) * p
	low := int(position)
	high := low + 1
	if high >= len(sorted) {
		return sorted[low]
	}
	weight := position - float64(low)
	return int(float64(sorted[low]) + (float64(sorted[high]-sorted[low]) * weight))
}

func largestOverlap(sets []map[string]struct{}, counts []int) (int, int) {
	first, second := 0, 1
	if counts[1] > counts[0] {
		first, second = 1, 0
	}
	for index := 2; index < len(counts); index++ {
		if counts[index] >= counts[first] {
			second, first = first, index
			continue
		}
		if counts[index] > counts[second] {
			second = index
		}
	}
	overlap := 0
	for fp := range sets[first] {
		if _, ok := sets[second][fp]; ok {
			overlap++
		}
	}
	return overlap, counts[first]
}

func uniqueBytes(unique map[string][]byte) int {
	total := 0
	for _, body := range unique {
		total += len(body)
	}
	return total
}

func walkSQL(root string, migrationsOnly bool) (int, int) {
	files, bytes := 0, 0
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			if entry != nil && entry.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if filepath.Ext(path) != ".sql" {
			return nil
		}
		if migrationsOnly && !containsMigrations(path) {
			return nil
		}
		info, statErr := entry.Info()
		if statErr != nil {
			return nil
		}
		files++
		bytes += int(info.Size())
		return nil
	})
	return files, bytes
}

func containsMigrations(path string) bool {
	return strings.Contains(filepath.ToSlash(path), "/migrations/")
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
