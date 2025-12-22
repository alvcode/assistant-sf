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
)

func SyncRun(ctx context.Context, head string) error {
	cnf := config.MustLoad(ctx)

	if !service.PathExists(cnf.FolderPath) {
		color.Red("Folder does not exist. Create a folder and specify the path to it in the configuration")
		return nil
	}

	err := syncRecursive(cnf.AssistantURL, head, cnf.FolderPath, nil)
	if err != nil {
		return err
	}

	return nil
}

func syncRecursive(domain string, head string, localPath string, parentID *int) error {
	var convertParent int
	if parentID != nil {
		convertParent = *parentID
	} else {
		convertParent = 0
	}

	/**
	Грузим дерево с сервера. *
	Делаем проход по дереву облака *
		Если папка. Проверяем есть ли локально. Если нет - создаем, проваливаемся. Если да - проваливаемся *
		Если файл. Есть ли локально? Если нет - скачивание, обновление хэша (если его нет). Если есть:
			- смотрим хэш. Если на сервере нет или различается.
				При режиме local - удаляем на сервере, загружаем с локалки, обновляем хэш
				При режиме server - скачиваем с сервера, обновляем хэш
			- если хэш совпадает - пропуск
	Делаем проход по локальному дереву
		Если папка. Проверяем есть ли в облаке. Если нет - создаем, проваливаемся. ?Если да - проваливаемся.?
									(вот тут подумать, может надо разделить на 2 метода, чтобы не выполнять одно и тоже)
		Если файл. Есть ли на сервере? Если нет - грузим, обновляем хэш. Если есть - предполагаем, что на предыдущем уровне все уже сделано
	*/

	color.Blue("Делаем запрос дерева. parentID: %d", convertParent)
	tree, err := service.GetTree(domain, parentID)
	if err != nil {
		color.Red(err.Error())
		return nil
	}

	cloudNames := make(map[string]*dto.DriveTree)

	for _, node := range tree {
		color.Blue("================= Cloud - Обработка ноды: %s ===================", node.Name)
		cloudNames[node.Name] = node
		dirPath := filepath.Join(localPath, node.Name)

		if node.Type == dict.StructTypeFolder {
			color.Blue("Это папка. Путь до нее: %s", dirPath)

			// создать если нет
			if !service.PathExists(dirPath) {
				color.Blue("Создаем её")
				err := os.MkdirAll(dirPath, 0755)
				if err != nil {
					return fmt.Errorf("could not create folder %s: %w", node.Name, err)
				}

				// рекурсивно обойти вложенные
				id := node.ID
				if err := syncRecursive(domain, head, dirPath, &id); err != nil {
					return err
				}
			}
		} else if node.Type == dict.StructTypeFile {

		}
	}
}
