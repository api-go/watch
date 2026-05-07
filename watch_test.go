package watch_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"apigo.cc/go/watch"
	"apigo.cc/go/file"
)

func TestWatch(t *testing.T) {
	testDir := "test_watch"
	_ = os.RemoveAll(testDir)
	_ = os.MkdirAll(testDir, 0755)
	defer os.RemoveAll(testDir)

	events := make(chan *watch.Event, 10)
	config := watch.Config{
		Paths: []string{testDir},
		Types: []string{".txt"},
	}

	w, err := watch.Start(config, func(e *watch.Event) {
		events <- e
	})
	if err != nil {
		t.Fatal(err)
	}
	defer w.Stop()

	// 1. 测试创建文件
	txtFile, _ := filepath.Abs(filepath.Join(testDir, "test.txt"))
	_ = file.Write(txtFile, "hello")
	
	select {
	case e := <-events:
		if e.Type != watch.Create || e.Path != txtFile {
			t.Errorf("expected create event for %s, got %v", txtFile, e)
		}
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for create event")
	}

	// 2. 测试排除类型
	config2 := watch.Config{
		Paths:        []string{testDir},
		ExcludeTypes: []string{".tmp"},
	}
	events2 := make(chan *watch.Event, 10)
	w2, _ := watch.Start(config2, func(e *watch.Event) {
		events2 <- e
	})
	defer w2.Stop()

	tmpFile := filepath.Join(testDir, "test.tmp")
	_ = file.Write(tmpFile, "temp")

	select {
	case e := <-events2:
		t.Errorf("should not receive event for .tmp file, got %v", e)
	case <-time.After(50 * time.Millisecond):
		// OK
	}

	// 3. 测试模糊排除
	config3 := watch.Config{
		Paths:    []string{testDir},
		Excludes: []string{"node_modules", "*.log"},
	}
	events3 := make(chan *watch.Event, 10)
	w3, _ := watch.Start(config3, func(e *watch.Event) {
		events3 <- e
	})
	defer w3.Stop()

	logFile := filepath.Join(testDir, "test.log")
	_ = file.Write(logFile, "log")
	
	nodeDir := filepath.Join(testDir, "node_modules")
	_ = os.MkdirAll(nodeDir, 0755)
	_ = file.Write(filepath.Join(nodeDir, "index.js"), "js")

	select {
	case e := <-events3:
		t.Errorf("should not receive event for excluded path/file, got %v", e)
	case <-time.After(50 * time.Millisecond):
		// OK
	}
}

func TestEasyStart(t *testing.T) {
	testDir := "test_easy"
	_ = os.RemoveAll(testDir)
	_ = os.MkdirAll(testDir, 0755)
	defer os.RemoveAll(testDir)

	done := make(chan bool)
	_, _ = watch.EasyStart(testDir, func(path string, et watch.EventType) {
		if et == watch.Create {
			done <- true
		}
	})

	_ = file.Write(filepath.Join(testDir, "easy.txt"), "easy")
	
	select {
	case <-done:
		// OK
	case <-time.After(100 * time.Millisecond):
		t.Fatal("timeout waiting for easy event")
	}
}
