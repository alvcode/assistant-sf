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

func FromDiskRun(ctx context.Context) error {
	cnf := config.MustLoad(ctx)

	if !service.PathExists(cnf.FolderPath) {
		color.Red("Folder does not exist. Create a folder and specify the path to it in the configuration")
		return nil
	}

	color.Green("folder exists")

	err := fromDiskRecursive(cnf.AssistantURL, cnf.FolderPath, nil)
	if err != nil {
		color.Red(err.Error())
		return nil
	}

	return nil
}

func fromDiskRecursive(domain string, localPath string, parentID *int) error {
	var convertParent int
	if parentID != nil {
		convertParent = *parentID
	} else {
		convertParent = 0
	}
	color.Blue("Делаем запрос дерева. parentID: %d", convertParent)
	tree, err := service.GetTree(domain, parentID)
	if err != nil {
		color.Red(err.Error())
		return nil
	}

	cloudNames := make(map[string]*dto.DriveTree)

	for _, node := range tree {
		color.Blue("================= Обработка ноды: %s ===================", node.Name)
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
			}

			// рекурсивно обойти вложенные
			id := node.ID
			if err := fromDiskRecursive(domain, dirPath, &id); err != nil {
				return err
			}
		} else if node.Type == dict.StructTypeFile {
			needDownload := false
			var fileHash string

			if service.FileExists(dirPath) {
				color.Blue("Файл %s существует", node.Name)

				fileHash, err = service.FileSHA256(dirPath)
				if err != nil {
					color.Red("error calculating file hash: %s", err)
					return nil
				}
				if node.SHA256 == nil || fileHash != *node.SHA256 {
					color.Blue("Хэша на сервере нет или они различаются. Надо загружать")
					needDownload = true
				} else {
					color.Blue("Хэш совпадает, не грузим")
				}
			} else {
				needDownload = true
				color.Blue("Файла локально нет. Надо загружать")
			}

			if needDownload {
				if !node.IsChunk {
					color.Blue("Загружаем обычным способом")
					fileBytes, err := service.GetFullFile(domain, node.ID)
					if err != nil {
						color.Red(err.Error())
						return nil
					}

					err = os.WriteFile(dirPath, fileBytes, 0644)
					if err != nil {
						color.Red("error save file: %s, %s", node.Name, err.Error())
						return nil
					}
				} else {
					color.Blue("Файл с чанками: %s", node.Name)
					maxChunk, err := service.GetMaxChunk(domain, node.ID)
					if err != nil {
						color.Red(err.Error())
						return nil
					}

					out, err := os.Create(dirPath)
					if err != nil {
						color.Red("cannot create file: %w", err)
						return nil
					}
					defer func(out *os.File) {
						err := out.Close()
						if err != nil {
							fmt.Printf("error closing file: %s, %s", node.Name, err.Error())
						}
					}(out)

					for i := 0; i <= maxChunk; i++ {
						color.Blue("Грузим чанк номер %d", i)
						chunkData, err := service.GetChunk(domain, node.ID, i)
						if err != nil {
							color.Red("failed to load chunk %d: %w", i, err)
							return nil
						}

						_, err = out.Write(chunkData)
						if err != nil {
							color.Red("failed to write chunk %d: %w", i, err)
							return nil
						}
					}
				}

				fileHash, err = service.FileSHA256(dirPath)
				if err != nil {
					color.Red("error calculating file hash: %s", err)
					return nil
				}

				err = service.UpdateHash(domain, node.ID, fileHash)
				if err != nil {
					color.Red("error update hash: %s, %s", node.Name, err.Error())
					return nil
				}
			}
		}
	}

	// теперь удаляем лишнее локально
	localEntries, err := os.ReadDir(localPath)
	if err != nil {
		return err
	}

	color.Blue("localPath: %s", localPath)

	for _, le := range localEntries {
		color.Blue("localEntries - %s", le.Name())
		if _, ok := cloudNames[le.Name()]; !ok {
			color.Blue("Локально есть файл или папка %s, которой нет в облаке, удаляем путь: %s", le.Name(), filepath.Join(localPath, le.Name()))
			// папки или файла нет в облаке, удаляем
			err := os.RemoveAll(filepath.Join(localPath, le.Name()))
			if err != nil {
				return err
			}
		}
	}

	return nil
}
