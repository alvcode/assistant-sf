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

func SyncRun(ctx context.Context, head string, isDebug bool) error {
	cnf := config.MustLoad(ctx)

	if !service.PathExists(cnf.FolderPath) {
		return errors.New("folder does not exist. Create a folder and specify the path to it in the configuration")
	}

	if isDebug {
		color.Yellow("sync folder exists")
	}

	err := syncRecursive(cnf.AssistantURL, head, cnf.FolderPath, nil, isDebug)
	if err != nil {
		return err
	}

	return nil
}

func syncRecursive(domain string, head string, localPath string, parentID *int, isDebug bool) error {
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
			color.Yellow("========= Cloud - Обработка ноды: %s ==========", node.Name)
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
					color.Yellow("Её нет. Создаем")
				}
				err := os.MkdirAll(dirPath, 0755)
				if err != nil {
					return fmt.Errorf("could not create folder %s: %w", node.Name, err)
				}
			}

			// рекурсивно обойти вложенные
			id := node.ID
			if err := syncRecursive(domain, head, dirPath, &id, isDebug); err != nil {
				return err
			}
		} else if node.Type == dict.StructTypeFile {
			needDownload := false
			needUpload := false
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
						color.Yellow("Хэша на сервере нет или они различаются. Надо загружать")
					}
					if head == "local" {
						needUpload = true
					} else if head == "server" {
						needDownload = true
					}
				} else {
					if isDebug {
						color.Yellow("Хэш совпадает, не грузим")
					}
				}
			} else {
				needDownload = true
				if isDebug {
					color.Yellow("Файла локально нет. Надо загружать")
				}
			}

			if needDownload {
				if !node.IsChunk {
					if isDebug {
						color.Yellow("Загружаем обычным способом")
					}
					fileBytes, err := service.GetFullFile(domain, node.ID)
					if err != nil {
						return fmt.Errorf("error get file: %w", err)
					}

					err = os.WriteFile(dirPath, fileBytes, 0644)
					if err != nil {
						return fmt.Errorf("error save file: %s, %w", node.Name, err)
					}
				} else {
					if isDebug {
						color.Yellow("Файл с чанками: %s", node.Name)
					}
					maxChunk, err := service.GetMaxChunk(domain, node.ID)
					if err != nil {
						return fmt.Errorf("error get max chunk: %w", err)
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
					return fmt.Errorf("error calculating file hash: %w", err)
				}

				err = service.UpdateHash(domain, node.ID, fileHash)
				if err != nil {
					return fmt.Errorf("error update hash: %s, %w", node.Name, err)
				}
			}

			if needUpload {
				if isDebug {
					color.Yellow("меняем на сервере")
				}
				err := service.DeleteStruct(domain, node.ID)
				if err != nil {
					return fmt.Errorf("error delete struct: %v", err)
				}

				localInfo, err := os.Stat(dirPath)
				if err != nil {
					return fmt.Errorf("cannot stat dir: %w", err)
				}

				newTree := make([]*dto.DriveTree, 0)
				if localInfo.Size() <= dict.ChunkSize {
					if isDebug {
						color.Yellow("Грузим обычным способом")
					}
					newTree, err = service.UploadFile(domain, dirPath, parentID)
					if err != nil {
						return fmt.Errorf("upload fail: %w", err)
					}
				} else {
					if isDebug {
						color.Yellow("Грузим чанками")
					}
					newTree, err = service.UploadFileByChunks(domain, dirPath, parentID)
					if err != nil {
						return fmt.Errorf("upload fail: %w", err)
					}
				}

				for _, newNode := range newTree {
					if newNode.Name == node.Name {
						fileHash, err = service.FileSHA256(dirPath)
						if err != nil {
							return fmt.Errorf("error calculating file hash: %w", err)
						}

						err = service.UpdateHash(domain, newNode.ID, fileHash)
						if err != nil {
							return fmt.Errorf("error update hash: %s, %w", node.Name, err)
						}

						cloudNames[newNode.Name] = newNode
					}
				}
			}
		}
	}

	localEntries, err := os.ReadDir(localPath)
	if err != nil {
		return fmt.Errorf("cannot read local directory: %w", err)
	}

	if isDebug {
		color.Yellow("============== Делаем обратное сканирование =============")
	}
	for _, le := range localEntries {
		if strings.Contains(le.Name(), ":") {
			continue
		}
		dirPath := filepath.Join(localPath, le.Name())

		if isDebug {
			color.Yellow("Обработка ноды: %s", dirPath)
		}

		localInfo, err := os.Stat(dirPath)
		if err != nil {
			return fmt.Errorf("error stat local directory: %w", err)
		}
		localIsDir := localInfo.IsDir()
		node, ok := cloudNames[le.Name()]

		if localIsDir {
			if !ok {
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
						break
					}
				}
				node, ok = cloudNames[le.Name()]
				if ok {
					if err := syncRecursive(domain, head, dirPath, &node.ID, isDebug); err != nil {
						return err
					}
				}
			}
		} else {
			if isDebug {
				color.Yellow("Это файл")
			}

			sha256, hashErr := service.FileSHA256(dirPath)
			if hashErr != nil {
				return fmt.Errorf("error calculating file hash: %w", err)
			}

			if !ok {
				newTree := make([]*dto.DriveTree, 0)

				if localInfo.Size() <= dict.ChunkSize {
					if isDebug {
						color.Yellow("Грузим обычным способом")
					}
					newTree, err = service.UploadFile(domain, dirPath, parentID)
					if err != nil {
						return fmt.Errorf("upload fail: %w", err)
					}
				} else {
					if isDebug {
						color.Yellow("Грузим чанками")
					}
					newTree, err = service.UploadFileByChunks(domain, dirPath, parentID)
					if err != nil {
						return fmt.Errorf(fmt.Sprintf("error upload file: %v", err))
					}
				}
				for _, newNode := range newTree {
					if newNode.Name == le.Name() {
						errUpdSha256 := service.UpdateHash(domain, newNode.ID, sha256)
						if errUpdSha256 != nil {
							return fmt.Errorf(fmt.Sprintf("error update hash: %v", errUpdSha256))
						}
						cloudNames[newNode.Name] = newNode
					}
				}
			} else {
				if isDebug {
					color.Yellow("Уже есть на сервере. Ничего не делаем")
				}
			}
		}
	}

	return nil
}
