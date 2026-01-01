package settings

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const configFile = "config.json"

var defaultSettings = Settings{
	ExportDir: "/tmp/cms-export",
}

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Get() (*Settings, error) {
	data, err := os.ReadFile(configFile)
	if err != nil {
		if os.IsNotExist(err) {
			// ファイルがなければデフォルト設定を返す
			return &defaultSettings, nil
		}
		return nil, err
	}

	var settings Settings
	if err := json.Unmarshal(data, &settings); err != nil {
		return nil, err
	}

	return &settings, nil
}

func (s *Service) Update(settings *Settings) error {
	// パスの検証
	if err := s.validateExportDir(settings.ExportDir); err != nil {
		return err
	}

	// JSONに変換
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}

	// ファイルに書き込み
	return os.WriteFile(configFile, data, 0644)
}

func (s *Service) validateExportDir(dir string) error {
	if dir == "" {
		return errors.New("export_dir is required")
	}

	// 絶対パスに変換
	absPath, err := filepath.Abs(dir)
	if err != nil {
		return errors.New("invalid path: " + err.Error())
	}

	// ディレクトリが存在しない場合は作成を試みる
	if err := os.MkdirAll(absPath, 0755); err != nil {
		return errors.New("cannot create directory: " + err.Error())
	}

	// 書き込み可能かテスト
	testFile := filepath.Join(absPath, ".cms-write-test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return errors.New("directory is not writable: " + err.Error())
	}
	os.Remove(testFile)

	return nil
}
