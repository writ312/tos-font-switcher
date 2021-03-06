// Demo code for the List primitive.
package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"

	"github.com/rivo/tview"
)

// ファイル操作部分

func makeSettingDir(basicSettingDir string) {
	if err := os.MkdirAll(basicSettingDir, 0777); err != nil {
		fmt.Println(err)
	}
}

func loadTosPath() string {
	basicSettingDir := os.Getenv("APPDATA") + "\\ToS-Font-Switcher"
	if _, err := os.Stat(basicSettingDir); err != nil {
		makeSettingDir(basicSettingDir)
	}
	file, _ := os.Open(basicSettingDir + "\\tos-path.txt")
	defer file.Close()
	bytes, _ := ioutil.ReadAll(file)
	tosPath := string(bytes)
	if len(tosPath) < 2 {
		if _, err := os.Stat("C:\\Program Files (x86)\\Steam\\steamapps\\common\\Tree of Savior (Japanese Ver.)"); err == nil {
			tosPath = "C:\\Program Files (x86)\\Steam\\steamapps\\common\\Tree of Savior (Japanese Ver.)\\release\\languageData"
		}
	}
	return tosPath
}

func saveFontListXML(tosPath string, defaultFont string) {
	fontListXML := `<?xml version="1.0" encoding="UTF-8"?>
	<!-- edited with XMLSPY v2004 rel. 3 U (http://www.xmlspy.com) by imc (imc) -->
	<Fontlist>
		<Font Name="ORIGINAL" Filename="%s"/>
		<Font Name="BOOK" Filename="%s"/>
	</Fontlist>
	`
	file, err := os.Create(tosPath + "\\Japanese\\fontlist.xml")
	if err != nil {
		//errorだと困るなぁ
		log.Fatal(err)
	}
	defer file.Close()
	// * UI簡略化のためにBookも同じフォントに。
	fmt.Fprintln(file, fmt.Sprintf(fontListXML, defaultFont, defaultFont))
}

// アドオン連携部分
func loadAddonSetting(tosPath string) string {
	var settings interface{}
	file, _ := os.OpenFile(tosPath+"\\settings.json", os.O_RDONLY|os.O_CREATE, 0666)
	defer file.Close()
	bytes, _ := ioutil.ReadAll(file)
	if len(bytes) == 0 {
		return ""
	}
	json.Unmarshal(bytes, &settings)
	return settings.(map[string]interface{})["Fontname"].(string)
}

func saveAddonSetting(tosPath string, defaultFontName string, fontlist []string) {
	type Settings struct {
		Fontname string   `json: "fontname"`
		Fontlist []string `json: "filelist"`
	}
	settings := new(Settings)
	settings.Fontname = defaultFontName
	settings.Fontlist = fontlist
	settingsJson, _ := json.Marshal(settings)
	ioutil.WriteFile(tosPath+"\\settings.json", settingsJson, os.ModePerm)
}

func getLoaclFontList(tosPath string) []string {
	var fontFiles []string
	if fileinfos, err := ioutil.ReadDir(tosPath + "\\Japanese\\font"); err == nil {
		for _, fileinfo := range fileinfos {
			fileName := fileinfo.Name()
			ext := fileName[len(fileName)-3:]
			if ext[:2] == "tt" {
				fontFiles = append(fontFiles, fileinfo.Name())
			}
		}
	}
	return fontFiles
}

// UI操作部分
// TODO　関数名にUPDATEとかでもつけたほうがわかりやすいかも

// func switchFont() {

// }

// func downloadFontList() {}

func downloadFontFile(fontUrl string) {
	response, err := http.Get(fontUrl)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fmt.Println("status:", response.Status)

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	u, _ := url.Parse(fontUrl)
	_, filename := path.Split(u.Path)

	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)

	if err != nil {
		fmt.Println(err)
	}

	defer func() {
		file.Close()
	}()

	file.Write(body)
	// TODO ZIP等なら展開して、フォルダ移動をゴニョゴニョ
}

// func unzip(file []byte) {

// 	// gzipの展開
// 	gzipReader, _ := gzip.NewReader(file)
// 	defer gzipReader.Close()

// 	// tarの展開
// 	tarReader := tar.NewReader(gzipReader)

// 	for {
// 		tarHeader, err := tarReader.Next()
// 		if err == io.EOF {
// 			break
// 		}

// 		// ファイルの特定
// 		if tarHeader.Name == "target.csv" {

// 			// あとはCSVの処理
// 			csvReader := csv.NewReader(tarReader)
// 			for {
// 				row, err := csvReader.Read()
// 				if err == io.EOF {
// 					break
// 				}
// 				fmt.Println("csv:", row)
// 			}
// 		}
// 	}
// }

func updateFontList(tosPath string, defaultFontName string, fontFileList []string, list *tview.List) {
	list.Clear()
	for _, fontname := range fontFileList {
		shortcut := '-'
		color := ""
		if defaultFontName == fontname {
			shortcut = '*'
			color = "[blue]"
		}

		selectFontName := fontname
		list.AddItem(color+fontname, "", shortcut, func() {
			saveFontListXML(tosPath, selectFontName)
			saveAddonSetting(tosPath, selectFontName, fontFileList)
			updateFontList(tosPath, selectFontName, fontFileList, list)
		})
	}
}

func initUI(list *tview.List) {
	tosPath := loadTosPath()
	if _, err := os.Stat(tosPath + "\\Japanese\\font"); err != nil {
		list.Clear()
		text := `ToSのインストール先が不明です
			Enterキーを押し、各自のインストール先を貼り付けてください
			例)C:\\Program Files (x86)\\Steam\\steamapps\\common\\Tree of Savior (Japanese Ver.)\\release\\languageData`
		list.AddItem(text, "", 's', func() {
			exec.Command("notepad", os.Getenv("APPDATA")+"\\ToS-Font-Switcher\\tos-path.txt").Run()
			initUI(list)
		})
	} else {
		defaultFontName := loadAddonSetting(tosPath)
		fontFileList := getLoaclFontList(tosPath)
		updateFontList(tosPath, defaultFontName, fontFileList, list)
	}

}

func main() {
	app := tview.NewApplication()
	list := tview.NewList().ShowSecondaryText(false)
	initUI(list)
	if err := app.SetRoot(list, true).Run(); err != nil {
		panic(err)
	}
}
