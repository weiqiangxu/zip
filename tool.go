package zip

import (
	"archive/tar"
	"compress/gzip"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type TgzPacker struct{}

func NewTgzPacker() *TgzPacker {
	return &TgzPacker{}
}

// removeTargetFile delete tar file
func (tp *TgzPacker) removeTargetFile(fileName string) (err error) {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return nil
	}
	return os.Remove(fileName)
}

// dirExists check dir
func (tp *TgzPacker) dirExists(dir string) bool {
	info, err := os.Stat(dir)
	return (err == nil || os.IsExist(err)) && info.IsDir()
}

// Pack sourceFullPath is dir or file path
func (tp *TgzPacker) Pack(sourceFullPath string, targetFilePath string) (err error) {
	sourceInfo, err := os.Stat(sourceFullPath)
	if err != nil {
		return err
	}
	if err = tp.removeTargetFile(targetFilePath); err != nil {
		return err
	}
	// create file fd
	file, err := os.Create(targetFilePath)
	if err != nil {
		return err
	}
	defer file.Close()
	gWriter := gzip.NewWriter(file)
	defer gWriter.Close()
	tarWriter := tar.NewWriter(gWriter)
	defer tarWriter.Close()
	if sourceInfo.IsDir() {
		return tp.tarFolder(sourceFullPath, filepath.Base(sourceFullPath), tarWriter)
	}
	return tp.tarFile(sourceFullPath, tarWriter)
}

// tarFile zip one file , sourceFullFile is a file
func (tp *TgzPacker) tarFile(sourceFullFile string, writer *tar.Writer) error {
	info, err := os.Stat(sourceFullFile)
	if err != nil {
		return err
	}
	header, err := tar.FileInfoHeader(info, "")
	if err != nil {
		return err
	}
	err = writer.WriteHeader(header)
	if err != nil {
		return err
	}
	fr, err := os.Open(sourceFullFile)
	if err != nil {
		return err
	}
	defer fr.Close()
	if _, err = io.Copy(writer, fr); err != nil {
		return err
	}
	return nil
}

// tarFolder sourceFullPath为待打包目录，baseName为待打包目录的根目录名称
func (tp *TgzPacker) tarFolder(sourceFullPath string, baseName string, writer *tar.Writer) error {
	// 保留最开始的原始目录，用于目录遍历过程中将文件由绝对路径改为相对路径
	baseFullPath := sourceFullPath
	return filepath.Walk(sourceFullPath, func(fileName string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Create header information
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		// Modify the name of the header according to the relative path
		// This is the root directory. Simply write the directory name into the header
		if fileName == baseFullPath {
			header.Name = baseName
		} else {
			// For no root directories, you need to process the path:
			// remove the first half of the absolute path
			// and then construct a relative path based on the root directory
			header.Name = filepath.Join(baseName, strings.TrimPrefix(fileName, baseFullPath))
		}

		if err = writer.WriteHeader(header); err != nil {
			return err
		}
		// There are many types of Linux files.
		// Here, only ordinary files are processed.
		// If the business needs to process other types of files
		// it is sufficient to add corresponding processing logic here
		if !info.Mode().IsRegular() {
			return nil
		}
		// Create a read handle to copy the content to the tarWriter
		fr, err := os.Open(fileName)
		if err != nil {
			return err
		}
		defer fr.Close()
		if _, err := io.Copy(writer, fr); err != nil {
			return err
		}
		return nil
	})
}

// UnPack TarFileName is the tar package to be uncompressed, and dstDir is the target directory to be uncompressed
func (tp *TgzPacker) UnPack(tarFileName string, dstDir string) (err error) {
	fr, err := os.Open(tarFileName)
	if err != nil {
		return err
	}
	defer func() {
		if err2 := fr.Close(); err2 != nil && err == nil {
			err = err2
		}
	}()
	gr, err := gzip.NewReader(fr)
	if err != nil {
		return err
	}
	defer func() {
		if err2 := gr.Close(); err2 != nil && err == nil {
			err = err2
		}
	}()
	tarReader := tar.NewReader(gr)
	for {
		header, err := tarReader.Next()
		switch {
		case err == io.EOF:
			return nil
		case err != nil:
			return err
		case header == nil:
			continue
		}
		targetFullPath := filepath.Join(dstDir, header.Name)
		switch header.Typeflag {
		case tar.TypeDir:
			// dir
			if exists := tp.dirExists(targetFullPath); !exists {
				if err = os.MkdirAll(targetFullPath, 0o755); err != nil {
					return err
				}
			}
		case tar.TypeReg:
			// Is a normal file, creating and writing the content to
			file, err := os.OpenFile(targetFullPath, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return err
			}
			_, err = io.Copy(file, tarReader)
			// Defer cannot be used within a loop. Close the file handle first
			if err2 := file.Close(); err2 != nil {
				return err2
			}
			// Here, we will judge the result of file copy
			if err != nil {
				return err
			}
		}
	}
}
