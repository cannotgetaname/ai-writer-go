package store

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidName(t *testing.T) {
	// 合法名称
	validNames := []string{
		"mybook",
		"我的书",
		"book-1_test",
		"TestBook123",
		"玄幻小说",
	}
	for _, name := range validNames {
		if !validName(name) {
			t.Errorf("validName(%s) should return true", name)
		}
	}

	// 非法名称
	invalidNames := []string{
		"",                         // 空名称
		"../etc/passwd",            // 路径遍历
		"book/name",                // 路径分隔符
		"book\\name",               // Windows 路径分隔符
		"book..name",               // 路径遍历变体
		"book/name/sub",            // 多层路径
		strings.Repeat("a", 101),   // 过长名称
		"book@name",                // 特殊字符
		"book!",                    // 特殊字符
		"book#name",                // 特殊字符
	}
	for _, name := range invalidNames {
		if validName(name) {
			t.Errorf("validName(%s) should return false", name)
		}
	}
}

func TestJSONStore_CreateBook(t *testing.T) {
	// 创建临时目录
	tmpDir := t.TempDir()
	store := NewJSONStore(tmpDir)

	// 测试正常创建
	meta, err := store.CreateBook("testbook")
	if err != nil {
		t.Fatalf("CreateBook failed: %v", err)
	}
	if meta.Name != "testbook" {
		t.Errorf("Expected name 'testbook', got '%s'", meta.Name)
	}

	// 检查目录是否创建
	bookPath := filepath.Join(tmpDir, "projects", "testbook")
	if _, err := os.Stat(bookPath); os.IsNotExist(err) {
		t.Error("Book directory not created")
	}

	// 测试重复创建
	_, err = store.CreateBook("testbook")
	if err == nil {
		t.Error("Creating duplicate book should fail")
	}

	// 测试非法名称
	_, err = store.CreateBook("../malicious")
	if err == nil {
		t.Error("Creating book with invalid name should fail")
	}
}

func TestJSONStore_DeleteBook(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewJSONStore(tmpDir)

	// 先创建书籍
	_, err := store.CreateBook("deletetest")
	if err != nil {
		t.Fatalf("CreateBook failed: %v", err)
	}

	// 正常删除
	err = store.DeleteBook("deletetest")
	if err != nil {
		t.Fatalf("DeleteBook failed: %v", err)
	}

	// 检查目录是否删除
	bookPath := filepath.Join(tmpDir, "projects", "deletetest")
	if _, err := os.Stat(bookPath); !os.IsNotExist(err) {
		t.Error("Book directory should be deleted")
	}

	// 测试删除非法名称
	err = store.DeleteBook("../malicious")
	if err == nil {
		t.Error("Deleting book with invalid name should fail")
	}
}

func TestJSONStore_LoadBook(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewJSONStore(tmpDir)

	// 创建书籍
	_, err := store.CreateBook("loadtest")
	if err != nil {
		t.Fatalf("CreateBook failed: %v", err)
	}

	// 正常加载
	book, err := store.LoadBook("loadtest")
	if err != nil {
		t.Fatalf("LoadBook failed: %v", err)
	}
	if book.Name != "loadtest" {
		t.Errorf("Expected name 'loadtest', got '%s'", book.Name)
	}

	// 测试加载非法名称
	_, err = store.LoadBook("../malicious")
	if err == nil {
		t.Error("Loading book with invalid name should fail")
	}
}

func TestJSONStore_ChapterParagraphs(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewJSONStore(tmpDir)

	// 创建书籍
	_, err := store.CreateBook("paratest")
	if err != nil {
		t.Fatalf("CreateBook failed: %v", err)
	}

	// 测试加载空章节
	paragraphs, err := store.LoadChapterParagraphs("paratest", 1)
	if err != nil {
		t.Fatalf("LoadChapterParagraphs failed: %v", err)
	}
	if paragraphs.Metadata.ParagraphCount != 0 {
		t.Errorf("Expected empty paragraphs, got %d", paragraphs.Metadata.ParagraphCount)
	}

	// 测试非法书名
	_, err = store.LoadChapterParagraphs("../malicious", 1)
	if err == nil {
		t.Error("Loading paragraphs with invalid name should fail")
	}
}

func TestJSONStore_SaveLoadMethods(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewJSONStore(tmpDir)

	// 创建书籍
	_, err := store.CreateBook("savetest")
	if err != nil {
		t.Fatalf("CreateBook failed: %v", err)
	}

	// 测试 LoadVolumes
	_, err = store.LoadVolumes("../malicious")
	if err == nil {
		t.Error("LoadVolumes with invalid name should fail")
	}

	// 测试 LoadCharacters
	_, err = store.LoadCharacters("../malicious")
	if err == nil {
		t.Error("LoadCharacters with invalid name should fail")
	}

	// 测试 SaveVolumes
	err = store.SaveVolumes("../malicious", nil)
	if err == nil {
		t.Error("SaveVolumes with invalid name should fail")
	}
}

func TestJSONStore_RenameBook(t *testing.T) {
	tmpDir := t.TempDir()
	store := NewJSONStore(tmpDir)

	// 先创建书籍
	_, err := store.CreateBook("oldname")
	if err != nil {
		t.Fatalf("CreateBook failed: %v", err)
	}

	// 正常重命名
	err = store.RenameBook("oldname", "newname")
	if err != nil {
		t.Fatalf("RenameBook failed: %v", err)
	}

	// 检查新目录是否存在
	newPath := filepath.Join(tmpDir, "projects", "newname")
	if _, err := os.Stat(newPath); os.IsNotExist(err) {
		t.Error("New book directory should exist")
	}

	// 检查旧目录是否已删除
	oldPath := filepath.Join(tmpDir, "projects", "oldname")
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Error("Old book directory should be deleted")
	}

	// 验证元数据更新
	book, err := store.LoadBook("newname")
	if err != nil {
		t.Fatalf("LoadBook after rename failed: %v", err)
	}
	if book.Name != "newname" {
		t.Errorf("Expected name 'newname', got '%s'", book.Name)
	}

	// 测试重命名到已存在的名称
	_, err = store.CreateBook("existing")
	if err != nil {
		t.Fatalf("CreateBook failed: %v", err)
	}
	err = store.RenameBook("newname", "existing")
	if err == nil {
		t.Error("Renaming to existing name should fail")
	}

	// 测试重命名不存在的书籍
	err = store.RenameBook("nonexistent", "anothername")
	if err == nil {
		t.Error("Renaming nonexistent book should fail")
	}

	// 测试非法旧名称
	err = store.RenameBook("../malicious", "goodname")
	if err == nil {
		t.Error("Renaming with invalid old name should fail")
	}

	// 测试非法新名称
	err = store.RenameBook("newname", "../malicious")
	if err == nil {
		t.Error("Renaming with invalid new name should fail")
	}
}