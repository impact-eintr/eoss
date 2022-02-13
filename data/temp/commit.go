package temp

import (
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/impact-eintr/eoss/data/locate"
	"github.com/impact-eintr/eoss/utils"
)

func (t *tempInfo) hash() string {
	s := strings.Split(t.Name, ".")
	return s[0]
}

func (t *tempInfo) id() int {
	s := strings.Split(t.Name, ".")
	id, _ := strconv.Atoi(s[1])
	return id
}

func commitTempObject(datFile string, tempinfo *tempInfo) {
	f, _ := os.Open(datFile)
	d := url.PathEscape(utils.CalculateHash(f))
	f.Close()
	// TODO 检测一下是否有重复文件
	if locate.Locate(tempinfo.Name) != -1 {
		os.Remove(datFile)
		log.Println("有重复文件!!!")
	} else {
		os.Rename(datFile, "/tmp/eoss/objects/"+tempinfo.Name+"."+d)
		locate.Add(tempinfo.hash(), tempinfo.id())
	}
}
