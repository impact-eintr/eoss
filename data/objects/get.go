package objects

import (
	"crypto/sha256"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/impact-eintr/eoss/data/locate"
	"github.com/impact-eintr/eoss/errmsg"
)

func Get(ctx *gin.Context) {
	file := getFile(ctx.Param("filehash")[1:])
	if file == "" {
		errmsg.ErrRawLog(ctx, http.StatusNotFound, "cannot find file")
		return
	}
	sendFile(ctx.Writer, file)
}

func getFile(name string) string {
	files, _ := filepath.Glob(os.Getenv("STORAGE_ROOT") + "/tmp/eoss/objects/" + name + ".*")
	if len(files) != 1 {
		return ""
	}
	file := files[0]
	h := sha256.New()
	sendFile(h, file)
	d := url.PathEscape(base64.StdEncoding.EncodeToString(h.Sum(nil)))
	hash := strings.Split(file, ".")[2]
	if d != hash {
		log.Println("object hash mismatch, remove", file)
		locate.Del(hash)
		os.Remove(file)
		return ""
	}
	return file
}

func sendFile(w io.Writer, file string) {
	f, _ := os.Open(file)
	defer f.Close()
	io.Copy(w, f)
}
