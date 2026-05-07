package watch

import (
	"fmt"
	"testing"
	"time"
)

func BenchmarkIsMatch(b *testing.B) {
	config := Config{
		Paths:        []string{"."},
		Types:        []string{".go", ".js", ".ts", ".md"},
		ExcludeTypes: []string{".log", ".tmp", ".exe"},
		Excludes:     []string{"**/node_modules/**", "vendor/**", "dist/**"},
		Events:       []EventType{Create, Change},
	}
	watcher, _ := Start(config, func(e *Event) {})
	defer watcher.Stop()

	testPath := "src/main.go"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		watcher.isMatch(testPath, Change, false)
	}
}

func BenchmarkIsExcluded(b *testing.B) {
	config := Config{
		Excludes: []string{"**/node_modules/**", "vendor/**", "dist/**", "*.log"},
	}
	watcher, _ := Start(config, func(e *Event) {})
	defer watcher.Stop()

	testPath := "/abs/path/to/project/node_modules/pkg/index.js"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		watcher.isExcluded(testPath, false)
	}
}

func BenchmarkDebounce(b *testing.B) {
	config := Config{
		Paths:    []string{"."},
		Debounce: 10 * time.Millisecond,
	}
	watcher, _ := Start(config, func(e *Event) {})
	defer watcher.Stop()

	event := &Event{
		Path: "test.txt",
		Type: Change,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		watcher.debounce(event)
	}
}

func BenchmarkKeyGeneration(b *testing.B) {
	event := &Event{
		Path: "/very/long/path/to/some/file/in/the/deep/directory/structure/test.txt",
		Type: Change,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = event.Path + ":" + string(event.Type)
	}
}

func BenchmarkKeyGenerationWithFmt(b *testing.B) {
	event := &Event{
		Path: "/very/long/path/to/some/file/in/the/deep/directory/structure/test.txt",
		Type: Change,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = fmt.Sprintf("%s:%s", event.Path, event.Type)
	}
}
