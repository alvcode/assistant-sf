package command

import (
	"assistant-sf/internal/config"
	"assistant-sf/internal/dict"
	"assistant-sf/internal/dto"
	"assistant-sf/internal/service"
	"context"
	"github.com/fatih/color"
	"os"
	"path/filepath"
)

const ChunkSize = 45 * 1024 * 1024

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
						cloudNames[node.Name] = newNode
					}
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
			if !ok || node.Type != dict.StructTypeFile {
				if ok && node.Type != dict.StructTypeFile {
					color.Blue("На сервере НЕ ФАЙЛ. Удаляем")
					err := service.DeleteStruct(domain, node.ID)
					if err != nil {
						color.Red("error delete struct: %v", err)
						return nil
					}
				}

				newTree := make([]*dto.DriveTree, 0)
				if localInfo.Size() <= ChunkSize {
					newTree, err = service.UploadFile(domain, dirPath, parentID)
				} else {

				}
				color.Blue("На сервере файла нет. Загрузка. Обновление хэша. Добавляем ноду в cloudNames")
			} else {
				color.Blue("На сервере файл есть. Сверяем хэш. Если нету или разница - удаление, загрузка, обновление хэша, добавление в cloudNames")
			}
		}
	}

	color.Blue("=== Обратное сканирование ===")
	for _, node := range cloudNames {
		dirPath := filepath.Join(localPath, node.Name)
		color.Blue("cloudPath: %s", dirPath)

		if node.Type == dict.StructTypeFile {
			if !service.FileExists(dirPath) {
				color.Blue("Локально файла нет. Удаляем на сервере")
			}
		} else if node.Type == dict.StructTypeFolder {
			if !service.PathExists(dirPath) {
				color.Blue("Локально папки нет. Удаляем на сервере")
			}
		}
	}

	/**
	Движемся по дереву локальной папки. Запрашиваем дерево в начале итерации.
	Если папка:
		если на сервере папки нет - создаем
		делаем рекурсивное проваливание
	если файл:
		если на сервере файла нет - загрузка, обновление хэша
		если есть - сверяем хэш:
			Если хэша на сервере нет или есть разница - удаляем, загружаем по новой, обновляем хэш
			Если хэш одинаковый - пропуск

	Запрашиваем дерево еще раз. Вторым циклом проходимся по дереву сервера
		если папки или файла локально нет - удаляем на сервере
	*/

	return nil
}
