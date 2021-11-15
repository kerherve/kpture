package utils

import "fmt"

const Version = "0.0.1"
const App = `
      __  ___       __   ___ 
|__/ |__)  |  |  | |__) |__  
|  \ |     |  \__/ |  \ |___ ` + "v" + Version

func PrintAppName() {
	fmt.Println(App)
	fmt.Println()
}
