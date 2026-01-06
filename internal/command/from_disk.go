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
)

func FromDiskRun(ctx context.Context, isDebug bool) error {
	cnf := config.MustLoad(ctx)

	if !service.PathExists(cnf.FolderPath) {
		return errors.New("folder does not exist. Create a folder and specify the path to it in the configuration")
	}

	err := service.ValidateSyncPath(cnf.FolderPath)
	if err != nil {
		return err
	}

	if isDebug {
		color.Yellow("sync folder exists")
	}

	err = fromDiskRecursive(cnf.AssistantURL, cnf.FolderPath, nil, cnf.ExcludeFolders, isDebug)
	if err != nil {
		return err
	}

	return nil
}

func fromDiskRecursive(domain string, localPath string, parentID *int, excludeFolders []string, isDebug bool) error {
	var convertParent int
	if parentID != nil {
		convertParent = *parentID
	} else {
		convertParent = 0
	}

	if isDebug {
		color.Yellow("Делаем запрос дерева. parentID: %d", convertParent)
	}

	tree, err := service.GetTree(domain, parentID)
	if err != nil {
		return fmt.Errorf("error get tree from server: %v", err)
	}

	cloudNames := make(map[string]*dto.DriveTree)

	for _, node := range tree {
		if isDebug {
			color.Yellow("======= Обработка ноды: %s ==========", node.Name)
		}

		if node.Type == dict.StructTypeFolder {
			needContinue := false
			for _, excludeFName := range excludeFolders {
				if excludeFName == node.Name {
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
		}

		cloudNames[node.Name] = node
		dirPath := filepath.Join(localPath, node.Name)

		if node.Type == dict.StructTypeFolder {
			if isDebug {
				color.Yellow("Это папка. Путь до нее: %s", dirPath)
			}

			// создать если нет
			if !service.PathExists(dirPath) {
				if isDebug {
					color.Yellow("Создаем папку")
				}
				err := os.MkdirAll(dirPath, 0755)
				if err != nil {
					return fmt.Errorf("could not create folder %s: %w", node.Name, err)
				}
			}

			// рекурсивно обойти вложенные
			id := node.ID
			if err := fromDiskRecursive(domain, dirPath, &id, excludeFolders, isDebug); err != nil {
				return err
			}
		} else if node.Type == dict.StructTypeFile {
			needDownload := false
			var fileHash string

			if service.FileExists(dirPath) {
				if isDebug {
					color.Yellow("Файл %s существует", node.Name)
				}

				fileHash, err = service.FileSHA256(dirPath)
				if err != nil {
					return fmt.Errorf("error calculating file hash: %s", err)
				}
				if node.SHA256 == nil || fileHash != *node.SHA256 {
					if isDebug {
						color.Yellow("Хэша на сервере нет или они различаются. Загружаем")
					}
					needDownload = true
				} else {
					if isDebug {
						color.Yellow("Хэш совпадает, не грузим")
					}
				}
			} else {
				needDownload = true
				if isDebug {
					color.Yellow("Файла локально нет. Загружаем")
				}
			}

			if needDownload {
				if !node.IsChunk {
					if isDebug {
						color.Yellow("Загружаем обычным способом")
					}
					fileBytes, err := service.GetFullFile(domain, node.ID)
					if err != nil {
						return fmt.Errorf("could not get file from server: %v", err)
					}

					err = os.WriteFile(dirPath, fileBytes, 0644)
					if err != nil {
						return fmt.Errorf("error save file: %s, %s", node.Name, err.Error())
					}
				} else {
					if isDebug {
						color.Yellow("Загружаем чанками")
					}
					maxChunk, err := service.GetMaxChunk(domain, node.ID)
					if err != nil {
						return fmt.Errorf("could not get max chunk from server: %v", err)
					}

					out, err := os.Create(dirPath)
					if err != nil {
						return fmt.Errorf("cannot create file: %w", err)
					}
					defer func(out *os.File) {
						err := out.Close()
						if err != nil {
							fmt.Printf("error closing file: %s, %s", node.Name, err.Error())
						}
					}(out)

					for i := 0; i <= maxChunk; i++ {
						if isDebug {
							color.Yellow("Грузим чанк номер %d", i)
						}
						chunkData, err := service.GetChunk(domain, node.ID, i)
						if err != nil {
							return fmt.Errorf("failed to load chunk %d: %w", i, err)
						}

						_, err = out.Write(chunkData)
						if err != nil {
							return fmt.Errorf("failed to write chunk %d: %w", i, err)
						}
					}
				}

				fileHash, err = service.FileSHA256(dirPath)
				if err != nil {
					return fmt.Errorf("error calculating file hash: %s", err)
				}

				err = service.UpdateHash(domain, node.ID, fileHash)
				if err != nil {
					return fmt.Errorf("error update hash: %s, %s", node.Name, err.Error())
				}
			}
		}
	}

	// теперь удаляем лишнее локально
	localEntries, err := os.ReadDir(localPath)
	if err != nil {
		return err
	}

	if isDebug {
		color.Yellow("localPath: %s", localPath)
	}

	for _, le := range localEntries {
		if isDebug {
			color.Yellow("localEntries - %s", le.Name())
		}
		if _, ok := cloudNames[le.Name()]; !ok {
			if isDebug {
				color.Yellow("Локально есть файл или папка %s, которой нет в облаке, удаляем путь: %s", le.Name(), filepath.Join(localPath, le.Name()))
			}
			// папки или файла нет в облаке, удаляем
			err := os.RemoveAll(filepath.Join(localPath, le.Name()))
			if err != nil {
				return err
			}
		}
	}

	return nil
}
