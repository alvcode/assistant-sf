package command

import (
	"assistant-sf/internal/config"
	"assistant-sf/internal/dict"
	"assistant-sf/internal/dto"
	"assistant-sf/internal/service"
	"context"
	"errors"
	"fmt"
	"github.com/fatih/color"
	"os"
	"path/filepath"
	"strings"
)

func ToDiskRun(ctx context.Context, isDebug bool) error {
	cnf := config.MustLoad(ctx)

	if !service.PathExists(cnf.FolderPath) {
		return errors.New("folder does not exist. Create a folder and specify the path to it in the configuration")
	}

	if isDebug {
		color.Yellow("folder exists")
	}

	err := toDiskRecursive(cnf.AssistantURL, cnf.FolderPath, nil, cnf.ExcludeFolders, isDebug)
	if err != nil {
		return err
	}

	return nil
}

func toDiskRecursive(domain string, localPath string, parentID *int, excludeFolders []string, isDebug bool) error {
	var convertParent int
	if parentID != nil {
		convertParent = *parentID
	} else {
		convertParent = 0
	}

	localEntries, err := os.ReadDir(localPath)
	if err != nil {
		return fmt.Errorf("error reading directory %s: %v", localPath, err)
	}

	if isDebug {
		color.Yellow("localPath: %s", localPath)
		color.Yellow("Делаем запрос дерева. parentID: %d", convertParent)
	}

	tree, err := service.GetTree(domain, parentID)
	if err != nil {
		return fmt.Errorf("error get tree from server: %v", err)
	}

	cloudNames := make(map[string]*dto.DriveTree)

	for _, node := range tree {
		needContinue := false
		if node.Type == dict.StructTypeFolder {
			for _, excludeFName := range excludeFolders {
				if excludeFName == node.Name {
					needContinue = true
					break
				}
			}
		}
		if !needContinue {
			cloudNames[node.Name] = node
		}
	}

	for _, le := range localEntries {
		if strings.Contains(le.Name(), ":") {
			continue
		}
		if isDebug {
			color.Yellow("=========== localEntry - %s ===========", le.Name())
		}

		dirPath := filepath.Join(localPath, le.Name())

		localInfo, err := os.Stat(dirPath)
		if err != nil {
			return fmt.Errorf("error get stat directory %s: %v", dirPath, err)
		}
		localIsDir := localInfo.IsDir()

		if localIsDir {
			if isDebug {
				color.Yellow("Это папка")
			}

			needContinue := false
			for _, excludeFName := range excludeFolders {
				if excludeFName == le.Name() {
					if isDebug {
						color.Yellow("Директория добавлена в исключения. Пропускаем")
					}
					needContinue = true
					break
				}
			}
			if needContinue {
				continue
			}

			node, ok := cloudNames[le.Name()]

			if !ok || node.Type != dict.StructTypeFolder {
				if ok && node.Type != dict.StructTypeFolder {
					if isDebug {
						color.Yellow("На сервере НЕ ПАПКА. Удаляем")
					}
					err := service.DeleteStruct(domain, node.ID)
					if err != nil {
						return fmt.Errorf("error delete struct: %v", err)
					}
				}
				if isDebug {
					color.Yellow("На сервере папки нет. Создаем. Тут надо получать ID для рекурсивного проваливания. А может быть обновлять по ключу ноду из ответа")
				}
				newTree, err := service.CreateDir(domain, le.Name(), parentID)
				if err != nil {
					return fmt.Errorf("error create dir: %v", err)
				}
				for _, newNode := range newTree {
					if newNode.Name == le.Name() {
						cloudNames[newNode.Name] = newNode
					}
				}

				node, ok = cloudNames[le.Name()]
				if ok {
					if err := toDiskRecursive(domain, dirPath, &node.ID, excludeFolders, isDebug); err != nil {
						return err
					}
				}
			} else {
				if isDebug {
					color.Yellow("На сервере папка есть. Рекурсивное проваливание")
				}
				if err := toDiskRecursive(domain, dirPath, &node.ID, excludeFolders, isDebug); err != nil {
					return err
				}
			}
		} else {
			if isDebug {
				color.Yellow("Это файл")
			}
			node, ok := cloudNames[le.Name()]

			sha256, hashErr := service.FileSHA256(dirPath)
			if hashErr != nil {
				return fmt.Errorf("error calculate sha256 for %s: %v", dirPath, hashErr)
			}

			if !ok || node.Type != dict.StructTypeFile {
				if ok && node.Type != dict.StructTypeFile {
					if isDebug {
						color.Yellow("На сервере НЕ ФАЙЛ. Удаляем")
					}
					err := service.DeleteStruct(domain, node.ID)
					if err != nil {
						return fmt.Errorf("error delete struct: %v", err)
					}
				}

				if isDebug {
					color.Yellow("На сервере файла нет. Загрузка. Обновление хэша. Добавляем ноду в cloudNames")
				}

				newTree := make([]*dto.DriveTree, 0)
				if localInfo.Size() <= dict.ChunkSize {
					if isDebug {
						color.Yellow("Грузим обычным способом")
					}
					newTree, err = service.UploadFile(domain, dirPath, parentID)
					if err != nil {
						return fmt.Errorf("error upload file: %v", err)
					}
				} else {
					if isDebug {
						color.Yellow("Грузим чанками")
					}
					newTree, err = service.UploadFileByChunks(domain, dirPath, parentID)
					if err != nil {
						return fmt.Errorf("error upload file: %v", err)
					}
				}
				for _, newNode := range newTree {
					if newNode.Name == le.Name() {
						errUpdSha256 := service.UpdateHash(domain, newNode.ID, sha256)
						if errUpdSha256 != nil {
							return fmt.Errorf("error update hash: %v", errUpdSha256)
						}
						cloudNames[newNode.Name] = newNode
					}
				}
			} else {
				if isDebug {
					color.Yellow("На сервере файл есть. Сверяем хэш. Если нету или разница - удаление, загрузка, обновление хэша, добавление в cloudNames")
				}
				if node.SHA256 == nil || *node.SHA256 != sha256 {
					err := service.DeleteStruct(domain, node.ID)
					if err != nil {
						return fmt.Errorf("error delete struct: %v", err)
					}

					newTree := make([]*dto.DriveTree, 0)
					if localInfo.Size() <= dict.ChunkSize {
						newTree, err = service.UploadFile(domain, dirPath, parentID)
						if err != nil {
							return fmt.Errorf("error upload file: %v", err)
						}
					} else {
						newTree, err = service.UploadFileByChunks(domain, dirPath, parentID)
						if err != nil {
							return fmt.Errorf("error upload file: %v", err)
						}
					}

					for _, newNode := range newTree {
						if newNode.Name == le.Name() {
							errUpdSha256 := service.UpdateHash(domain, newNode.ID, sha256)
							if errUpdSha256 != nil {
								return fmt.Errorf("error update hash: %v", errUpdSha256)
							}
							cloudNames[newNode.Name] = newNode
						}
					}
				}
			}
		}
	}

	if isDebug {
		color.Yellow("=== Обратное сканирование ===")
	}
	for _, node := range cloudNames {
		dirPath := filepath.Join(localPath, node.Name)
		if isDebug {
			color.Yellow("cloudPath: %s", dirPath)
		}

		needDelete := false
		if node.Type == dict.StructTypeFile {
			if !service.FileExists(dirPath) {
				if isDebug {
					color.Yellow("Локально файла нет. Удаляем на сервере")
				}
				needDelete = true
			}
		} else if node.Type == dict.StructTypeFolder {
			if !service.PathExists(dirPath) {
				if isDebug {
					color.Yellow("Локально папки нет. Удаляем на сервере")
				}
				needDelete = true
			}
		}

		if needDelete {
			err := service.DeleteStruct(domain, node.ID)
			if err != nil {
				return fmt.Errorf("error delete struct: %v", err)
			}
		}
	}

	return nil
}
