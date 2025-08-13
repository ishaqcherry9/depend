package replacer

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"time"

	"github.com/ishaqcherry9/depend/pkg/gofile"
)

var _ Replacer = (*replacerInfo)(nil)

type Replacer interface {
	SetReplacementFields(fields []Field)
	SetSubDirsAndFiles(subDirs []string, subFiles ...string)
	SetIgnoreSubDirs(dirs ...string)
	SetIgnoreSubFiles(filenames ...string)
	SetOutputDir(absDir string, name ...string) error
	GetOutputDir() string
	GetSourcePath() string
	SaveFiles() error
	ReadFile(filename string) ([]byte, error)
	GetFiles() []string
	SaveTemplateFiles(m map[string]interface{}, parentDir ...string) error
}

type replacerInfo struct {
	path              string
	fs                embed.FS
	isActual          bool
	files             []string
	ignoreFiles       []string
	ignoreDirs        []string
	replacementFields []Field
	outPath           string
}

func New(path string) (Replacer, error) {
	files, err := gofile.ListFiles(path)
	if err != nil {
		return nil, err
	}

	path, _ = filepath.Abs(path)
	return &replacerInfo{
		path:              path,
		isActual:          true,
		files:             files,
		replacementFields: []Field{},
	}, nil
}

func NewFS(path string, fs embed.FS) (Replacer, error) {
	files, err := listFiles(path, fs)
	if err != nil {
		return nil, err
	}

	return &replacerInfo{
		path:              path,
		fs:                fs,
		isActual:          false,
		files:             files,
		replacementFields: []Field{},
	}, nil
}

type Field struct {
	Old             string
	New             string
	IsCaseSensitive bool
}

func (r *replacerInfo) SetReplacementFields(fields []Field) {
	var newFields []Field
	for _, field := range fields {
		if field.IsCaseSensitive && isFirstAlphabet(field.Old) {
			if field.New == "" {
				continue
			}
			newFields = append(newFields,
				Field{
					Old: strings.ToUpper(field.Old[:1]) + field.Old[1:],
					New: strings.ToUpper(field.New[:1]) + field.New[1:],
				},
				Field{
					Old: strings.ToLower(field.Old[:1]) + field.Old[1:],
					New: strings.ToLower(field.New[:1]) + field.New[1:],
				},
			)
		} else {
			newFields = append(newFields, field)
		}
	}
	r.replacementFields = newFields
}

func (r *replacerInfo) GetFiles() []string {
	return r.files
}

func (r *replacerInfo) SetSubDirsAndFiles(subDirs []string, subFiles ...string) {
	subDirs = r.convertPathsDelimiter(subDirs...)
	subFiles = r.convertPathsDelimiter(subFiles...)

	var files []string
	isExistFile := make(map[string]struct{})
	for _, file := range r.files {
		for _, dir := range subDirs {
			if isSubPath(file, dir) {
				if _, ok := isExistFile[file]; ok {
					continue
				}
				isExistFile[file] = struct{}{}
				files = append(files, file)
			}
		}
		for _, sf := range subFiles {
			if isMatchFile(file, sf) {
				if _, ok := isExistFile[file]; ok {
					continue
				}
				isExistFile[file] = struct{}{}
				files = append(files, file)
			}
		}
	}

	if len(files) == 0 {
		return
	}
	r.files = files
}

func (r *replacerInfo) SetIgnoreSubFiles(filenames ...string) {
	r.ignoreFiles = append(r.ignoreFiles, filenames...)
}

func (r *replacerInfo) SetIgnoreSubDirs(dirs ...string) {
	dirs = r.convertPathsDelimiter(dirs...)
	r.ignoreDirs = append(r.ignoreDirs, dirs...)
}

func (r *replacerInfo) SetOutputDir(absPath string, name ...string) error {
	if absPath != "" {
		abs, err := filepath.Abs(absPath)
		if err != nil {
			return err
		}

		r.outPath = abs
		return nil
	}

	subPath := strings.Join(name, "_")
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}
	r.outPath = pwd + gofile.GetPathDelimiter() + subPath + "_" + time.Now().Format("150405")
	return nil
}

func (r *replacerInfo) GetOutputDir() string {
	return r.outPath
}

func (r *replacerInfo) GetSourcePath() string {
	return r.path
}

func (r *replacerInfo) ReadFile(filename string) ([]byte, error) {
	filename = r.convertPathDelimiter(filename)

	foundFile := []string{}
	for _, file := range r.files {
		if strings.Contains(file, filename) && gofile.GetFilename(file) == gofile.GetFilename(filename) {
			foundFile = append(foundFile, file)
		}
	}
	if len(foundFile) != 1 {
		return nil, fmt.Errorf("total %d file named '%s', files=%+v", len(foundFile), filename, foundFile)
	}

	if r.isActual {
		return os.ReadFile(foundFile[0])
	}
	return r.fs.ReadFile(foundFile[0])
}

func (r *replacerInfo) SaveFiles() error {
	if r.outPath == "" {
		r.outPath = gofile.GetRunPath() + gofile.GetPathDelimiter() + "generate_" + time.Now().Format("150405")
	}

	var existFiles []string
	var writeData = make(map[string][]byte)

	for _, file := range r.files {
		if r.isInIgnoreDir(file) || r.isIgnoreFile(file) {
			continue
		}

		var data []byte
		var err error

		if r.isActual {
			data, err = os.ReadFile(file)
		} else {
			data, err = r.fs.ReadFile(file)
		}
		if err != nil {
			return err
		}

		for _, field := range r.replacementFields {
			data = bytes.ReplaceAll(data, []byte(field.Old), []byte(field.New))
		}

		newFilePath := r.getNewFilePath(file)
		dir, filename := filepath.Split(newFilePath)
		for _, field := range r.replacementFields {
			if strings.Contains(dir, field.Old) {
				dir = strings.ReplaceAll(dir, field.Old, field.New)
			}
			if strings.Contains(filename, field.Old) {
				filename = strings.ReplaceAll(filename, field.Old, field.New)
			}

			if newFilePath != dir+filename {
				newFilePath = dir + filename
			}
		}

		if gofile.IsExists(newFilePath) {
			existFiles = append(existFiles, newFilePath)
		}
		writeData[newFilePath] = data
	}

	if len(existFiles) > 0 {

		return fmt.Errorf("existing files detected\n    %s\nCode generation has been cancelled\n",
			strings.Join(existFiles, "\n    "))
	}

	for file, data := range writeData {
		if isForbiddenFile(file, r.path) {
			return fmt.Errorf("disable writing file(%s) to directory(%s), file size=%d", file, r.path, len(data))
		}
	}

	for file, data := range writeData {
		err := saveToNewFile(file, data)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *replacerInfo) SaveTemplateFiles(m map[string]interface{}, parentDir ...string) error {
	refDir := ""
	if len(parentDir) > 0 {
		refDir = strings.Join(parentDir, gofile.GetPathDelimiter())
	}

	writeData := make(map[string][]byte, len(r.files))
	for _, file := range r.files {
		data, err := replaceTemplateData(file, m)
		if err != nil {
			return err
		}
		newFilePath := r.getNewFilePath2(file, refDir)
		newFilePath = trimExt(newFilePath)
		if gofile.IsExists(newFilePath) {
			return fmt.Errorf("file %s already exists, cancel code generation", newFilePath)
		}
		newFilePath, err = replaceTemplateFilePath(newFilePath, m)
		if err != nil {
			return err
		}
		writeData[newFilePath] = data
	}

	for file, data := range writeData {
		err := saveToNewFile(file, data)
		if err != nil {
			return err
		}
	}

	return nil
}

func replaceTemplateData(file string, m map[string]interface{}) ([]byte, error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("read file failed, err=%s", err)
	}
	if !bytes.Contains(data, []byte("{{")) {
		return data, nil
	}

	builder := bytes.Buffer{}
	tmpl, err := template.New(file).Parse(string(data))
	if err != nil {
		return nil, fmt.Errorf("parse data failed, err=%s", err)
	}
	err = tmpl.Execute(&builder, m)
	if err != nil {
		return nil, fmt.Errorf("execute data failed, err=%s", err)
	}
	return builder.Bytes(), nil
}

func replaceTemplateFilePath(file string, m map[string]interface{}) (string, error) {
	if !strings.Contains(file, "{{") {
		return file, nil
	}

	builder := strings.Builder{}
	tmpl, err := template.New("file: " + file).Parse(file)
	if err != nil {
		return file, fmt.Errorf("parse file failed, err=%s", err)
	}
	err = tmpl.Execute(&builder, m)
	if err != nil {
		return file, fmt.Errorf("execute file failed, err=%s", err)
	}
	return builder.String(), nil
}

func trimExt(file string) string {
	file = strings.TrimSuffix(file, ".tmpl")
	file = strings.TrimSuffix(file, ".tpl")
	file = strings.TrimSuffix(file, ".template")
	return file
}

func (r *replacerInfo) isIgnoreFile(file string) bool {
	isIgnore := false
	for _, v := range r.ignoreFiles {
		if isMatchFile(file, v) {
			isIgnore = true
			break
		}
	}
	return isIgnore
}

func (r *replacerInfo) isInIgnoreDir(file string) bool {
	isIgnore := false
	dir, _ := filepath.Split(file)
	for _, v := range r.ignoreDirs {
		if strings.Contains(dir, v) {
			isIgnore = true
			break
		}
	}
	return isIgnore
}

func isForbiddenFile(file string, path string) bool {
	if gofile.IsWindows() {
		path = strings.ReplaceAll(path, "/", "\\")
		file = strings.ReplaceAll(file, "/", "\\")
	}
	return strings.Contains(file, path)
}

func (r *replacerInfo) getNewFilePath(file string) string {
	newFilePath := r.outPath + strings.Replace(file, r.path, "", 1)

	if gofile.IsWindows() {
		newFilePath = strings.ReplaceAll(newFilePath, "/", "\\")
	}

	return newFilePath
}

func (r *replacerInfo) getNewFilePath2(file string, refDir string) string {
	if refDir == "" {
		return r.getNewFilePath(file)
	}

	newFilePath := r.outPath + gofile.GetPathDelimiter() + refDir + gofile.GetPathDelimiter() + strings.Replace(file, r.path, "", 1)
	if gofile.IsWindows() {
		newFilePath = strings.ReplaceAll(newFilePath, "/", "\\")
	}
	return newFilePath
}

func (r *replacerInfo) convertPathDelimiter(filePath string) string {
	if r.isActual && gofile.IsWindows() {
		filePath = strings.ReplaceAll(filePath, "/", "\\")
	}
	return filePath
}

func (r *replacerInfo) convertPathsDelimiter(filePaths ...string) []string {
	if r.isActual && gofile.IsWindows() {
		filePathsTmp := []string{}
		for _, dir := range filePaths {
			filePathsTmp = append(filePathsTmp, strings.ReplaceAll(dir, "/", "\\"))
		}
		return filePathsTmp
	}
	return filePaths
}

func saveToNewFile(filePath string, data []byte) error {

	dir, _ := filepath.Split(filePath)
	err := os.MkdirAll(dir, 0766)
	if err != nil {
		return err
	}

	err = os.WriteFile(filePath, data, 0666)
	if err != nil {
		return err
	}

	return nil
}

func listFiles(path string, fs embed.FS) ([]string, error) {
	files := []string{}
	err := walkDir(path, &files, fs)
	return files, err
}

func walkDir(dirPath string, allFiles *[]string, fs embed.FS) error {
	files, err := fs.ReadDir(dirPath)
	if err != nil {
		return err
	}

	for _, file := range files {
		deepFile := dirPath + "/" + file.Name()
		if file.IsDir() {
			_ = walkDir(deepFile, allFiles, fs)
			continue
		}
		*allFiles = append(*allFiles, deepFile)
	}

	return nil
}

func isFirstAlphabet(str string) bool {
	if len(str) == 0 {
		return false
	}

	if (str[0] >= 'A' && str[0] <= 'Z') || (str[0] >= 'a' && str[0] <= 'z') {
		return true
	}

	return false
}

func isSubPath(filePath string, subPath string) bool {
	dir, _ := filepath.Split(filePath)
	return strings.Contains(dir, subPath)
}

func isMatchFile(filePath string, sf string) bool {
	dir1, file1 := filepath.Split(filePath)
	dir2, file2 := filepath.Split(sf)
	if file1 != file2 {
		return false
	}

	if gofile.IsWindows() {
		dir1 = strings.ReplaceAll(dir1, "/", "\\")
		dir2 = strings.ReplaceAll(dir2, "/", "\\")
	} else {
		dir1 = strings.ReplaceAll(dir1, "\\", "/")
		dir2 = strings.ReplaceAll(dir2, "\\", "/")
	}

	l1, l2 := len(dir1), len(dir2)
	if l1 >= l2 && dir1[l1-l2:] == dir2 {
		return true
	}
	return false
}
