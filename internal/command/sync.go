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

func SyncRun(ctx context.Context, head string) error {
	cnf := config.MustLoad(ctx)

	if !service.PathExists(cnf.FolderPath) {
		return errors.New("folder does not exist. Create a folder and specify the path to it in the configuration")
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
			- смотрим хэш. Если на сервере нет или различается. *
				При режиме local - удаляем на сервере, загружаем с локалки, обновляем хэш *
				При режиме server - скачиваем с сервера, обновляем хэш *
			- если хэш совпадает - пропуск *
	Делаем проход по локальному дереву
		* Если папка. Проверяем есть ли в облаке. Если нет - создаем, проваливаемся. ?Если да - проваливаемся.?
									(вот тут подумать, может надо разделить на 2 метода, чтобы не выполнять одно и тоже)
		Если файл. Есть ли на сервере? Если нет - грузим, обновляем хэш. Если есть - предполагаем, что на предыдущем уровне все уже сделано
	*/

	color.Blue("Делаем запрос дерева. parentID: %d", convertParent)
	tree, err := service.GetTree(domain, parentID)
	if err != nil {
		return fmt.Errorf("error get tree from server: %v", err)
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
			}

			// рекурсивно обойти вложенные
			id := node.ID
			if err := syncRecursive(domain, head, dirPath, &id); err != nil {
				return err
			}
		} else if node.Type == dict.StructTypeFile {
			needDownload := false
			needUpload := false
			var fileHash string

			if service.FileExists(dirPath) {
				color.Blue("Файл %s существует", node.Name)

				fileHash, err = service.FileSHA256(dirPath)
				if err != nil {
					return fmt.Errorf("error calculating file hash: %s", err)
				}
				if node.SHA256 == nil || fileHash != *node.SHA256 {
					color.Blue("Хэша на сервере нет или они различаются. Надо загружать")
					if head == "local" {
						needUpload = true
					} else if head == "server" {
						needDownload = true
					}
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
						return fmt.Errorf("error get file: %w", err)
					}

					err = os.WriteFile(dirPath, fileBytes, 0644)
					if err != nil {
						return fmt.Errorf("error save file: %s, %w", node.Name, err)
					}
				} else {
					color.Blue("Файл с чанками: %s", node.Name)
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
						color.Blue("Грузим чанк номер %d", i)
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
				color.Blue("меняем на сервере")
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
					color.Blue("Грузим обычным способом")
					newTree, err = service.UploadFile(domain, dirPath, parentID)
					if err != nil {
						return fmt.Errorf("upload fail: %w", err)
					}
				} else {
					color.Blue("Грузим чанками")
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

	for _, le := range localEntries {
		if strings.Contains(le.Name(), ":") {
			continue
		}
		dirPath := filepath.Join(localPath, le.Name())

		localInfo, err := os.Stat(dirPath)
		if err != nil {
			return fmt.Errorf("error stat local directory: %w", err)
		}
		localIsDir := localInfo.IsDir()
		node, ok := cloudNames[le.Name()]

		if localIsDir {
			if !ok {
				color.Blue("На сервере папки нет. Создаем. Тут надо получать ID для рекурсивного проваливания. А может быть обновлять по ключу ноду из ответа")
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
					if err := syncRecursive(domain, head, dirPath, &node.ID); err != nil {
						return err
					}
				}
			}
		} else {
			color.Blue("Это файл")

			sha256, hashErr := service.FileSHA256(dirPath)
			if hashErr != nil {
				return fmt.Errorf("error calculating file hash: %w", err)
			}

			if !ok {
				newTree := make([]*dto.DriveTree, 0)

				if localInfo.Size() <= dict.ChunkSize {
					color.Blue("Грузим обычным способом")
					newTree, err = service.UploadFile(domain, dirPath, parentID)
					if err != nil {
						return fmt.Errorf("upload fail: %w", err)
					}
				} else {
					color.Blue("Грузим чанками")
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
			}
		}
	}

	return nil
}
