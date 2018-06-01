package main

import (
	"os"
	"fmt"
	"io/ioutil"
	"bytes"
	"crypto/sha256"
	"io"
	"encoding/hex"
	"strings"
	"net/http"
	"errors"
	"os/exec"
	"context"
	"runtime"

	"github.com/qiniu/api.v7/auth/qbox"
	"github.com/qiniu/api.v7/storage"
	"path"
)

type QiNiuConfig struct {
	Access string `json:"access"`
	Secret string `json:"secret"`
	Bucket string `json:"bucket"`
	Domain string `json:"domain"`
}

var Config QiNiuConfig


func main() {

	configInit()

	if len(os.Args) < 2 {
		fmt.Println("Please input file path or url")
		return
	}

	path := os.Args[1]
	key, err := upload(path)
	if err != nil {
		fmt.Println("Upload failed: ",err.Error())
		return
	}

	url := "http://" + Config.Domain + "/" + key
	if err := head(url); err != nil {
		fmt.Println("Head failed: ", err.Error())
		return
	}

	fmt.Println()
	fmt.Println(url)

	fmt.Println()
	if err := clip(url); err != nil {
		fmt.Println("copy failed: ", err.Error())
	} else {
		fmt.Println("copied!")
	}

	return
}

func configInit() {
	Config = QiNiuConfig{
		Access: "xxx",
		Secret: "xxx",
		Bucket: "xxx",
		Domain: "xxx",
	}
}

func upload(filename string) (string, error) {
	var buffer []byte

	/* download network image */
	if strings.HasPrefix(filename, "https://") || strings.HasPrefix(filename, "http://") {
		client := http.Client{}
		req, err := http.NewRequest("GET", filename, nil)
		if err != nil {
			return "", err
		}
		resp, err := client.Do(req)
		if err != nil {
			return "", err
		}
		if resp.StatusCode != http.StatusOK {
			return "", errors.New(fmt.Sprintf("%v return %v", filename, resp.StatusCode))
		}
		buffer, err = ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return "", err
		}
	} else {
		file, err := os.Open(filename)
		if err != nil {
			return "", err
		}
		buffer, err = ioutil.ReadAll(file)
		file.Close()
		if err != nil {
			return "", err
		}
	}

	reader := bytes.NewReader(buffer)
	hasher := sha256.New()
	/* 把文件哈希 */
	if _, err := io.Copy(hasher, reader); err != nil {
		return "", err
	}
	hash := hex.EncodeToString(hasher.Sum(nil))
	reader.Seek(0, io.SeekStart)

	key := hash + path.Ext(filename)

	putPolicy := &storage.PutPolicy{
		Scope: Config.Bucket,
	}
	upToken := putPolicy.UploadToken(qbox.NewMac(Config.Access, Config.Secret))
	formUploader := storage.NewFormUploader(&storage.Config{
		Zone: &storage.ZoneHuanan,
		UseHTTPS: false,
		UseCdnDomains: false,
	})
	return key, formUploader.Put(context.Background(), nil, upToken, key, reader, int64(reader.Len()), nil)
}

// HTTP HEAD 方法判断上传路径有效性
func head(url string) error {
	client := http.Client{}
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return err
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("%v return %v", url, resp.StatusCode))
	}
	return nil
}

func clip(content string) error {
	if runtime.GOOS != "windows" {
		// yaourt xclip
		cmd := exec.Command("xclip", "-selection", "c")
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return err
		}
		go func() {
			defer stdin.Close()
			io.WriteString(stdin, content)
		}()
		err = cmd.Run()
		return err
	}
	cmd := exec.Command("clip")
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return err
	}
	go func() {
		defer stdin.Close()
		io.WriteString(stdin, content)
	}()

	err = cmd.Run()
	return err
}