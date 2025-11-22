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

	//tree, err := service.GetTree(cnf.AssistantURL, nil)
	//if err != nil {
	//	color.Red(err.Error())
	//	return nil
	//}
	//
	//for _, node := range tree {
	//	fmt.Println(node.Name)
	//	if node.Type == dict.StructTypeFolder {
	//		if !service.PathExists(filepath.Join(cnf.FolderPath, node.Name)) {
	//			err := os.MkdirAll(filepath.Join(cnf.FolderPath, node.Name), 0755)
	//			if err != nil {
	//				color.Red(fmt.Sprintf("could not create folder %s: %w", node.Name, err.Error()))
	//				return nil
	//			}
	//		}
	//	} else if node.Type == dict.StructTypeFile {
	//
	//	}
	//}
	//
	//entries, err := os.ReadDir(cnf.FolderPath)
	//if err != nil {
	//	return err
	//}
	//
	//for _, e := range entries {
	//	full := filepath.Join(cnf.FolderPath, e.Name())
	//
	//	if e.IsDir() {
	//		fmt.Println("DIR:", full)
	//		//scan(full)
	//	} else {
	//		fmt.Println("FILE:", full)
	//	}
	//}
	//
	//fmt.Println(tree)

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
		color.Blue("Обработка ноды: %s", node.Name)
		cloudNames[node.Name] = node

		if node.Type == dict.StructTypeFolder {
			dirPath := filepath.Join(localPath, node.Name)
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
			if !node.IsChunk {
				filePath := filepath.Join(localPath, node.Name)
				if service.FileExists(filePath) {
					// сравнение хэшей если различаются - скачиваем.
				} else {
					// файла нет, скачиваем
					fileBytes, err := service.GetFullFile(domain, node.ID)
					if err != nil {
						color.Red(err.Error())
						return nil
					}

					err = os.WriteFile(filePath, fileBytes, 0644)
					if err != nil {
						color.Red("error save file: %s, %s", node.Name, err.Error())
						return nil
					}
				}
			} else {
				maxChunk, err := service.GetMaxChunk(domain, node.ID)
				if err != nil {
					return err
				}

				/**
				// 2) Создаём файл
				    out, err := os.Create(dstPath)
				    if err != nil {
				        return fmt.Errorf("cannot create file: %w", err)
				    }
				    defer out.Close()

				    // 3) Последовательно грузим чанки
				    for i := 0; i <= maxChunk; i++ {
				        chunkData, err := getChunk(domain, structID, i)
				        if err != nil {
				            return fmt.Errorf("failed to load chunk %d: %w", i, err)
				        }

				        _, err = out.Write(chunkData)
				        if err != nil {
				            return fmt.Errorf("failed to write chunk %d: %w", i, err)
				        }
				    }
				*/
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
