package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/shermp/go-osk/osk"
)

func main() {
	keymapJSON, _ := ioutil.ReadFile("../keymaps/keymap-en_us.json")
	km := osk.KeyMap{}
	json.Unmarshal(keymapJSON, &km)
	vk, _ := osk.New(&km, 1080, 1440)

	// svgFile, err := os.OpenFile("./osk-en_us.svg", os.O_WRONLY|os.O_CREATE, 0644)
	// if err != nil {
	// 	return
	// }
	// vk.CreateSVG(svgFile)
	// svgFile.Close()
	vk.CreateIMG()
}
