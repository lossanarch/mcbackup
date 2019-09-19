package targz

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

func Decompress(gz io.Reader) (*gzip.Reader, error) {
	return gzip.NewReader(gz)
}

func Compress(data io.Writer) *gzip.Writer {
	return gzip.NewWriter(data)
}

func Tar(data io.Writer) *tar.Writer {
	t := tar.NewWriter(data)

	return t
}

func GetFilesMultiReader(files []*os.File) (io.Reader, error) {
	var rs []io.Reader
	for _, f := range files {
		rs = append(rs, getFileReader(f))
	}

	return io.MultiReader(rs...), nil
}

func getFileReader(file *os.File) io.Reader {
	return bufio.NewReader(file)
}

func BufferedReadWrite(basePath string, writePath string, files []*os.File) error {
	outfile, err := os.Create(writePath + ".tgz")
	if err != nil {
		return err
	}
	defer outfile.Close()

	gz := Compress(outfile)
	defer gz.Close()
	w := Tar(gz)
	defer w.Close()
	// w := GetTarGZWriter(outfile)

	// buf := make([]byte, 4194304) //Read 4MB blocks
	for _, f := range files {
		info, err := os.Stat(f.Name())
		if err != nil {
			return err
		}
		switch mode := info.Mode(); {
		case mode.IsDir():
			hdr, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			hdr.Name = strings.TrimPrefix(f.Name(), basePath+"/")
			// fmt.Println("Header:", hdr)
			err = w.WriteHeader(hdr)
			if err != nil {
				return err
			}
			// err = WriteData(w, nil)
			// if err != nil {
			// 	return err
			// }
		case mode.IsRegular():
			hdr, err := tar.FileInfoHeader(info, "")
			if err != nil {
				return err
			}
			hdr.Name = strings.TrimPrefix(f.Name(), basePath+"/")
			// fmt.Println("Header:", hdr)
			err = w.WriteHeader(hdr)
			if err != nil {
				return err
			}

			r := getFileReader(f)

			// for {

			buf, err := ioutil.ReadAll(r)
			// n, err := r.Read(buf)
			if err != nil && err != io.EOF {
				return err
			}
			// if n == 0 {
			// continue
			// }

			if err := WriteData(w, &buf); err != nil {
				return err
			}

			// if err == io.EOF {
			// break
			// }
			// }

		}

	}

	err = w.Flush()
	if err != nil {
		return err
	}

	err = gz.Flush()
	if err != nil {
		return err
	}

	err = outfile.Sync()
	if err != nil {
		return err
	}

	return nil
}

func GetTarGZWriter(f *os.File) *tar.Writer {
	g := Compress(f)
	t := Tar(g)

	return t
}

func WriteData(t *tar.Writer, data *[]byte) error {
	// fmt.Println("Size:", len(*data))
	_, err := t.Write(*data)
	if err != nil {
		return err
	}
	return nil
}

func ReadDir(path string) ([]string, error) {
	var list []string

	err := filepath.Walk(path, func(path string, f os.FileInfo, err error) error {
		list = append(list, path)
		return err
	})

	if err != nil {
		return nil, err
	}

	return list, nil
}
