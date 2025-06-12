package handler

import (
	"fmt"
	"net/http"
)

// package handler

// import (
// 	"fmt"
// 	"net/http"
// 	"time"
// )

// func Date(w http.ResponseWriter, r *http.Request) {
// 	currentTime := time.Now().Format(time.RFC850)
// 	fmt.Fprint(w, currentTime)
// }

package handler

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)


func UploadFile(w http.ResponseWriter, r *http.Request) {
	link := r.URL.Query().Get("link")
	file := r.URL.Query().Get("file")
	err = uploadFtpFile(link, file)
	if err != nil {
		log.Printf("failed to upload file %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}


var golbalFiles = make(map[string]map[string]*CustomTitleWatch, 0)
var golbalChanels = make(map[string]chan workerDownload, 0)
var globalDownloadedFiles = make(map[string]bool, 0)
var lockDownloadedFile sync.Mutex
var LOCAL_STORAGE = "/home/jwzwtyt/public_html"
var HOST_EXPOSE = "https://thuvu33.site"
var ROOT_EXPOSE = "https://phim.ge"

const (
	v1SkyUrl = "https://v1.kingcdn.xyz"
	v3SkyUrl = "https://v3.kingcdn.xyz"
	v4SkyUrl = "https://v4.kingcdn.xyz"

	v2SkyUrl  = "https://v2.kingcdn.xyz"
	v5SkyUrl  = "https://v5.kingcdn.xyz"
	v1filmUrl = "https://f1film.b-cdn.net"
	v3filmUrl = "https://f3film.b-cdn.net"
	v4filmUrl = "https://f4film.b-cdn.net"

	v2filmUrl = "https://f2film.b-cdn.net"
	v5filmUrl = "https://f5film.b-cdn.net"
)

func downloadFile(url string, filename string) error {
	start := time.Now().Unix()
	if false {
		log.Printf("download file %v %v", filename, url)

	}
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("failed to download file %v", err)
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filename)
	if err != nil {
		log.Printf("failed to create file %v", err)
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		log.Printf("failed to copy file %v", err)
	} else {
		end := time.Now().Unix()
		lockDownloadedFile.Lock()
		globalDownloadedFiles[filename] = true
		lockDownloadedFile.Unlock()
		if false {
			log.Printf("[XXX] download file %v success, time eslapse %v", filename, end-start)

		}
	}
	return err
}

func decodeBase64(line string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(line)
	if err != nil {
		return "", fmt.Errorf("base64 decode error: %w", err)
	}
	return string(decoded), nil
}

func decodeInternal(line string) string {
	s := ""
	for _, char := range line {
		decimalVal := int(char)

		switch {
		case 32 <= decimalVal && decimalVal <= 64:
			s += string(rune(decimalVal))
		case 65 <= decimalVal && decimalVal <= 84:
			s += string(rune(decimalVal + 6))
		case 85 <= decimalVal && decimalVal <= 90:
			s += string(rune(decimalVal - 20))
		case 91 <= decimalVal && decimalVal <= 116:
			s += string(rune(decimalVal + 6))
		case 117 <= decimalVal && decimalVal <= 122:
			s += string(rune(decimalVal - 20))
		}
	}
	return s
}

type workerDownload struct {
	LocalPath string
	RemoteUrl string
}

type CustomTitleWatch struct {
	OldEncodeJpgLink string
	OldDecodeJpgLink string
	NewDecodeJpgLink string
	NewEncodeJpgLink string
	CdnDecodeJpgLink string
	WorkerDownload   chan workerDownload
	Downloaded       bool
	LocalPath        string
}

func processFileM3u8(filmId string, oldHostName string, newHostName string, iFile string, oFile string, decodeFile string, linkFile string) (map[string]*CustomTitleWatch, error) {
	result := make(map[string]*CustomTitleWatch, 0)
	keyPattern := regexp.MustCompile(`URI="(.*?).key"`)
	dataPattern := regexp.MustCompile(`(.*)/([0-9a-zA-Z=]*).jpg`)

	// Suggested code may be subject to a license. Learn more: ~LicenseLog:4129918822.

	infile, err := os.Open(iFile)
	if err != nil {
		panic(err)
	}
	defer infile.Close()

	outfile, err := os.Create(oFile)
	if err != nil {
		panic(err)
	}
	defer outfile.Close()

	// Reading file line by line
	fileContentBytes, err := os.ReadFile(iFile)
	if err != nil {
		panic(err)
	}
	fileContent := string(fileContentBytes)
	lines := strings.Split(fileContent, "\n")
	indexJpg := 1

	zfile, err := os.Create(linkFile)
	if err != nil {
		panic(err)
	}
	defer zfile.Close()

	rfile, err := os.Create(decodeFile)
	if err != nil {
		panic(err)
	}
	defer rfile.Close()

	for _, line := range lines {
		// fmt.Println("1: ", line)

		m := keyPattern.FindStringSubmatch(line)
		n := dataPattern.FindStringSubmatch(line)

		if n != nil {
			// log.Printf("DEBUG 1: %v , %v", n[1], n[2])

			o := CustomTitleWatch{}
			url := n[2]
			// fmt.Println("1.1: ", line)
			// log.Printf("DEBUG 2: %v , %v", n[1], n[2])
			o.OldEncodeJpgLink = fmt.Sprintf("%v/%v.jpg", n[1], url)
			// respUrl, err := URL.Parse(url)
			// if err != nil {
			// 	panic(err)
			// }
			// h := fmt.Sprintf("%v://%v", respUrl.Scheme, respUrl.Hostname())

			newURL := decodeInternal(url)
			newURLDecoded, err := decodeBase64(newURL)
			o.OldDecodeJpgLink = fmt.Sprintf("%v/%v.jpg", n[1], newURLDecoded)
			zfile.WriteString(o.OldDecodeJpgLink + "\n")
			if err != nil {
				fmt.Println("error decoding base64:", err)
				_, err = outfile.WriteString(line + "\n")
				if err != nil {
					panic(err)
				}

				_, err = rfile.WriteString(line + "\n")
				if err != nil {
					panic(err)
				}

				continue
			}
			o.NewDecodeJpgLink = fmt.Sprintf("%v/%s/%d.jpg", newHostName, filmId, indexJpg)
			// newLine := dataPattern.ReplaceAllString(line, fmt.Sprintf(`%s/%s/%d.jpg`, newHostName, filmId, indexJpg))
			newLine := dataPattern.ReplaceAllString(line, fmt.Sprintf(`%s/%s.jpg`, v2filmUrl, newURLDecoded))

			if strings.Contains(line, v1SkyUrl) {
				newLine = dataPattern.ReplaceAllString(line, fmt.Sprintf(`%s/%s.jpg`, v1filmUrl, newURLDecoded))
			}
			if strings.Contains(line, v3SkyUrl) {
				newLine = dataPattern.ReplaceAllString(line, fmt.Sprintf(`%s/%s.jpg`, v3filmUrl, newURLDecoded))
			}
			if strings.Contains(line, v4SkyUrl) {
				newLine = dataPattern.ReplaceAllString(line, fmt.Sprintf(`%s/%s.jpg`, v4filmUrl, newURLDecoded))
			}

			if strings.Contains(line, v5SkyUrl) {
				newLine = dataPattern.ReplaceAllString(line, fmt.Sprintf(`%s/%s.jpg`, v5filmUrl, newURLDecoded))
			}
			o.CdnDecodeJpgLink = newLine

			_, err = outfile.WriteString(newLine + "\n")
			if err != nil {
				panic(err)
			}
			_, err = rfile.WriteString(o.OldDecodeJpgLink + "\n")
			if err != nil {
				panic(err)
			}

			result[fmt.Sprintf("%d", indexJpg)] = &o
			indexJpg += 1
		} else if m != nil {
			o := CustomTitleWatch{}
			url := m[1] // 			downloadFile(v1.OldDecodeJpgLink, fmt.Sprintf("%s/%s/%s.jpg", LOCAL_STORAGE, filmId, v))

			newURL := decodeInternal(url)
			newURLDecoded, err := decodeBase64(newURL)
			if err != nil {
				fmt.Println("error decoding base64:", err)
				_, err = outfile.WriteString(line + "\n")
				if err != nil {
					panic(err)
				}
				continue
			}
			o.OldEncodeJpgLink = fmt.Sprintf("%s/%s.key", oldHostName, url)
			o.OldDecodeJpgLink = fmt.Sprintf(`%v/%s.key`, oldHostName, newURLDecoded)
			zfile.WriteString(o.OldDecodeJpgLink + "\n")
			o.NewDecodeJpgLink = fmt.Sprintf(`%v/%s/%d.jpg`, newHostName, filmId, 0)

			// newLine := keyPattern.ReplaceAllString(line, fmt.Sprintf(`URI="%v/%s/%d.jpg"`, newHostName, filmId, 0))
			newLine := keyPattern.ReplaceAllString(line, fmt.Sprintf(`URI="%v.key"`, filmId))

			_, err = outfile.WriteString(newLine + "\n")
			if err != nil {
				panic(err)
			}
			rNewLine := keyPattern.ReplaceAllString(line, fmt.Sprintf(`URI="%v.key"`, newURLDecoded))

			rKey := fmt.Sprintf(rNewLine + "\n")
			_, err = rfile.WriteString(rKey + "\n")
			if err != nil {
				panic(err)
			}
			result[fmt.Sprintf("%v", 0)] = &o
		} else {
			_, err = outfile.WriteString(line + "\n")
			if err != nil {
				panic(err)
			}

			_, err = rfile.WriteString(line + "\n")
			if err != nil {
				panic(err)
			}
		}
	}

	// fmt.Println("Successfully converted python to go")
	return result, nil
}

func fastDownload(u string, filmId string, newHostName string, iFile string, oFile string, decodeFile string, linkFile string) (map[string]*CustomTitleWatch, error) {
	respUrl, err := url.Parse(u)
	if err != nil {
		panic(err)
	}
	h := fmt.Sprintf("%v://%v", respUrl.Scheme, respUrl.Hostname())

	e := downloadFile(u, iFile)
	if e != nil {
		log.Printf("failed to download file %v", e)
		return nil, e
	}
	log.Printf("download file %v success", iFile)
	return processFileM3u8(filmId, h, newHostName, iFile, oFile, decodeFile, linkFile)
}

func uploadToHost(fileName string, content []byte) error {
	url := "https://phim.ge/upload.php"
	method := "POST"

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)
	part1, errFile1 := writer.CreateFormFile("uploaded_file", fileName)
	if errFile1 != nil {
		log.Printf("error creating form file: %v", errFile1)
		return errFile1
	}
	file := bytes.NewReader(content)
	_, errFile1 = io.Copy(part1, file)
	if errFile1 != nil {
		log.Printf("error copying file content: %v", errFile1)
		return errFile1
	}
	err := writer.Close()
	if err != nil {
		log.Printf("error closing writer: %v", err)
		return err
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)

	if err != nil {
		log.Printf("error creating new request: %v", err)
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	res, err := client.Do(req)
	if err != nil {
		log.Printf("error making request: %v", err)
		return err
	}
	defer res.Body.Close()
	return nil

}

func uploadFtpFile(sourceUrl string, ftpFilePath string) error {
	if false {
		log.Printf("upload file %v %v", ftpFilePath, sourceUrl)

	}
	if !strings.HasPrefix(sourceUrl, "http://") && !strings.HasPrefix(sourceUrl, "https://") {
		c, err := os.ReadFile(sourceUrl)
		if err != nil {
			log.Printf("failed to read local file %v: %v", sourceUrl, err)
			return err
		}
		return uploadToHost(ftpFilePath, c)

	}

	resp, err := http.Get(sourceUrl)
	if err != nil {
		log.Printf("failed to download file %v", err)
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("failed to read response body %v", err)
		return err
	}
	return uploadToHost(ftpFilePath, body)
}

func downloadFileForCdn(url string) error {
	if false {
		log.Printf("download file %v ", url)
	}
	lockDownloadedFile.Lock()
	if _, ok := globalDownloadedFiles[url]; ok {
		if false {
			log.Printf("file %v already downloaded", url)
		}
		lockDownloadedFile.Unlock()
		return nil
	}
	globalDownloadedFiles[url] = true
	lockDownloadedFile.Unlock()
	client := &http.Client{
		Timeout: 300 * time.Second,
	}
	resp, err := client.Get(url)
	if err != nil {
		if false {
			log.Printf("[XXX] download failed: %v", err)
		}
	} else {
		// log.Printf("[ZZZ] download passed: %v", url)
		resp.Body.Close()
	}
	return nil
}

func worker(jobs <-chan workerDownload) {
	for j := range jobs {
		// log.Printf("worker download file %v %v", j.LocalPath, j.RemoteUrl)
		// downloadFile(j.RemoteUrl, j.LocalPath)
		downloadFileForCdn(j.RemoteUrl)
	}
}

func backgroundDownload(jobs chan workerDownload) {

	for w := 1; w <= 1000; w++ {
		go worker(jobs)
	}
}

func handleTask(filmId string, srcUrl interface{}) (interface{}, error) {
	LOCAL_STORAGE := "/tmp/"
	var err error
	oldFile := fmt.Sprintf("%s/old_%v.m3u8", LOCAL_STORAGE, filmId)
	newFile := fmt.Sprintf("%s/%v.m3u8", LOCAL_STORAGE, filmId)
	decodeFile := fmt.Sprintf("%s/decode_%v.m3u8", LOCAL_STORAGE, filmId)
	linkFile := fmt.Sprintf("%s/%v.link", LOCAL_STORAGE, filmId)

	result, e := fastDownload(srcUrl.(string), filmId, HOST_EXPOSE, oldFile, newFile, decodeFile, linkFile)

	if v, ok := result["0"]; ok && e == nil {

		// golbalFiles[filmId] = result
		if e := downloadFile(v.OldDecodeJpgLink, fmt.Sprintf("%s/%v.key", LOCAL_STORAGE, filmId)); e != nil {
			log.Printf("failed to download key file %v", e)
			return nil, e
		}
		t4 := time.Now()

		err = uploadFtpFile(v.OldDecodeJpgLink, fmt.Sprintf("%s.key", filmId))
		if err != nil {
			log.Printf("failed to upload key file %v", err)
			return nil, err
		}
		t5 := time.Now()
		log.Printf("upload key file took %v seconds for filmId %s", t5.Sub(t4).Seconds(), filmId)
		err = uploadFtpFile(fmt.Sprintf("%s/%v.m3u8", LOCAL_STORAGE, filmId), fmt.Sprintf("%s.m3u8", filmId))
		if err != nil {
			log.Printf("failed to upload m3u8 file %v", err)
			return nil, err
		}
		t6 := time.Now()
		log.Printf("upload m3u8 file took %v seconds for filmId %s", t6.Sub(t5).Seconds(), filmId)
		jobs := make(chan workerDownload, 2000)
		go backgroundDownload(jobs)

		for k, _ := range result {
			i, _ := strconv.Atoi(k)
			if false {
				log.Printf("push download file to chanel %v", result[fmt.Sprintf("%d", i)].OldDecodeJpgLink)
			}
			jobs <- workerDownload{LocalPath: fmt.Sprintf("%s/%s/%d.jpg", LOCAL_STORAGE, filmId, i),
				RemoteUrl: result[fmt.Sprintf("%d", i)].CdnDecodeJpgLink}

		}
		t7 := time.Now()
		log.Printf("background download took %v seconds for filmId %s", t7.Sub(t6).Seconds(), filmId)
	}
	return nil, nil
}
