package command

import (
	"assistant-sf/internal/config"
	"assistant-sf/internal/dict"
	"assistant-sf/internal/dto"
	"assistant-sf/internal/service"
	"context"
	"fmt"
	"github.com/fatih/color"
	"os"
	"path/filepath"
	"strings"
)

func ToDiskRun(ctx context.Context) error {
	cnf := config.MustLoad(ctx)

	if !service.PathExists(cnf.FolderPath) {
		color.Red("Folder does not exist. Create a folder and specify the path to it in the configuration")
		return nil
	}

	color.Green("folder exists")

	err := toDiskRecursive(cnf.AssistantURL, cnf.FolderPath, nil)
	if err != nil {
		color.Red(err.Error())
		return nil
	}

	return nil
}

func toDiskRecursive(domain string, localPath string, parentID *int) error {
	var convertParent int
	if parentID != nil {
		convertParent = *parentID
	} else {
		convertParent = 0
	}

	localEntries, err := os.ReadDir(localPath)
	if err != nil {
		color.Red(err.Error())
		return nil
	}

	color.Blue("localPath: %s", localPath)

	color.Blue("Делаем запрос дерева. parentID: %d", convertParent)
	tree, err := service.GetTree(domain, parentID)
	if err != nil {
		color.Red(err.Error())
		return nil
	}

	cloudNames := make(map[string]*dto.DriveTree)

	for _, node := range tree {
		cloudNames[node.Name] = node
	}

	for _, le := range localEntries {
		if strings.Contains(le.Name(), ":") {
			continue
		}
		color.Blue("=================== localEntry - %s ========================", le.Name())

		dirPath := filepath.Join(localPath, le.Name())

		localInfo, err := os.Stat(dirPath)
		if err != nil {
			color.Red(err.Error())
			return nil
		}
		localIsDir := localInfo.IsDir()

		if localIsDir {
			color.Blue("Это папка")
			node, ok := cloudNames[le.Name()]

			if !ok || node.Type != dict.StructTypeFolder {
				if ok && node.Type != dict.StructTypeFolder {
					color.Blue("На сервере НЕ ПАПКА. Удаляем")
					err := service.DeleteStruct(domain, node.ID)
					if err != nil {
						color.Red("error delete struct: %v", err)
						return nil
					}
				}

				color.Blue("На сервере папки нет. Создаем. Тут надо получать ID для рекурсивного проваливания. А может быть обновлять по ключу ноду из ответа")
				newTree, err := service.CreateDir(domain, le.Name(), parentID)
				if err != nil {
					color.Red("error create dir: %v", err)
					return nil
				}
				for _, newNode := range newTree {
					if newNode.Name == le.Name() {
						cloudNames[newNode.Name] = newNode
					}
				}
				if err := toDiskRecursive(domain, dirPath, &node.ID); err != nil {
					return err
				}
			} else {
				color.Blue("На сервере папка есть. Рекурсивное проваливание")
				if err := toDiskRecursive(domain, dirPath, &node.ID); err != nil {
					return err
				}
			}
		} else {
			color.Blue("Это файл")
			node, ok := cloudNames[le.Name()]

			sha256, hashErr := service.FileSHA256(dirPath)
			if hashErr != nil {
				color.Red(hashErr.Error())
				return nil
			}

			if !ok || node.Type != dict.StructTypeFile {
				if ok && node.Type != dict.StructTypeFile {
					color.Blue("На сервере НЕ ФАЙЛ. Удаляем")
					err := service.DeleteStruct(domain, node.ID)
					if err != nil {
						color.Red("error delete struct: %v", err)
						return nil
					}
				}

				color.Blue("На сервере файла нет. Загрузка. Обновление хэша. Добавляем ноду в cloudNames")

				newTree := make([]*dto.DriveTree, 0)
				if localInfo.Size() <= dict.ChunkSize {
					color.Blue("Грузим обычным способом")
					newTree, err = service.UploadFile(domain, dirPath, parentID)
					if err != nil {
						color.Red(fmt.Sprintf("error upload file: %v", err))
					}
				} else {
					color.Blue("Грузим чанками")
					newTree, err = service.UploadFileByChunks(domain, dirPath, parentID)
					if err != nil {
						color.Red(fmt.Sprintf("error upload file: %v", err))
					}
				}
				for _, newNode := range newTree {
					if newNode.Name == le.Name() {
						errUpdSha256 := service.UpdateHash(domain, newNode.ID, sha256)
						if errUpdSha256 != nil {
							color.Red(fmt.Sprintf("error update hash: %v", errUpdSha256))
							return nil
						}
						cloudNames[newNode.Name] = newNode
					}
				}
			} else {
				color.Blue("На сервере файл есть. Сверяем хэш. Если нету или разница - удаление, загрузка, обновление хэша, добавление в cloudNames")
				if node.SHA256 == nil || *node.SHA256 != sha256 {
					err := service.DeleteStruct(domain, node.ID)
					if err != nil {
						color.Red("error delete struct: %v", err)
						return nil
					}

					newTree := make([]*dto.DriveTree, 0)
					if localInfo.Size() <= dict.ChunkSize {
						newTree, err = service.UploadFile(domain, dirPath, parentID)
					} else {
						newTree, err = service.UploadFileByChunks(domain, dirPath, parentID)
					}

					for _, newNode := range newTree {
						if newNode.Name == le.Name() {
							errUpdSha256 := service.UpdateHash(domain, newNode.ID, sha256)
							if errUpdSha256 != nil {
								color.Red(fmt.Sprintf("error update hash: %v", errUpdSha256))
								return nil
							}
							cloudNames[newNode.Name] = newNode
						}
					}
				}
			}
		}
	}

	color.Blue("=== Обратное сканирование ===")
	for _, node := range cloudNames {
		dirPath := filepath.Join(localPath, node.Name)
		color.Blue("cloudPath: %s", dirPath)

		needDelete := false
		if node.Type == dict.StructTypeFile {
			if !service.FileExists(dirPath) {
				color.Blue("Локально файла нет. Удаляем на сервере")
				needDelete = true
			}
		} else if node.Type == dict.StructTypeFolder {
			if !service.PathExists(dirPath) {
				color.Blue("Локально папки нет. Удаляем на сервере")
				needDelete = true
			}
		}

		if needDelete {
			err := service.DeleteStruct(domain, node.ID)
			if err != nil {
				color.Red("error delete struct: %v", err)
				return nil
			}
		}
	}

	return nil
}
